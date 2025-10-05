package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toritsuyo/gwt/internal/gitx"
	"github.com/toritsuyo/gwt/internal/naming"
	"github.com/toritsuyo/gwt/internal/tmux"
)

type tmuxNewConfig struct {
	baseDir     string
	count       int
	layout      string
	syncPanes   bool
	noAttach    bool
	sessionName string
}

// newTmuxCmd creates the root tmux command
func newTmuxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tmux",
		Short: "Manage tmux sessions with worktrees",
		Long:  `Manage tmux sessions with worktrees.`,
	}

	// Add subcommands
	cmd.AddCommand(newTmuxNewCmd())

	return cmd
}

func newTmuxNewCmd() *cobra.Command {
	cfg := &tmuxNewConfig{}

	cmd := &cobra.Command{
		Use:   "new <branch> [<start-point>]",
		Short: "Create worktree(s) and launch tmux session",
		Long: `Create one or more worktrees and launch them in a tmux session.

Creates worktrees with numbered suffixes (branch-1, branch-2, etc.) and opens them in tmux panes.

Examples:
  gwt tmux new feature/auth
  gwt tmux new feature/auth --count 3
  gwt tmux new feature/auth main --count 3 --sync-panes`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(c *cobra.Command, args []string) error {
			return runTmuxNew(c, args, cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.baseDir, "base-dir", "", "Base directory for worktree placement (defaults to repository parent)")
	cmd.Flags().IntVar(&cfg.count, "count", 1, "Number of worktrees to create")
	cmd.Flags().StringVar(&cfg.layout, "layout", "tiled", "Tmux layout (tiled/horizontal/vertical)")
	cmd.Flags().BoolVar(&cfg.syncPanes, "sync-panes", false, "Enable tmux synchronize-panes (send same input to all panes)")
	cmd.Flags().BoolVar(&cfg.noAttach, "no-attach", false, "Don't attach to tmux session")
	cmd.Flags().StringVar(&cfg.sessionName, "session-name", "", "Custom tmux session name")

	return cmd
}

var tmuxCmd = newTmuxCmd()

func init() {
	tmuxCmd = newTmuxCmd()
	rootCmd.AddCommand(tmuxCmd)
}

func validateLayout(layout string) error {
	if layout == "" {
		return nil // Empty is allowed (uses tmux default)
	}

	validLayouts := []string{
		"tiled",
		"horizontal",
		"vertical",
		"even-horizontal",
		"even-vertical",
		"main-horizontal",
		"main-vertical",
	}

	for _, valid := range validLayouts {
		if layout == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid layout: %s (must be one of: %s)", layout, "tiled, horizontal, vertical, even-horizontal, even-vertical, main-horizontal, main-vertical")
}

func runTmuxNew(cmd *cobra.Command, args []string, cfg *tmuxNewConfig) error {
	ctx := cmd.Context()
	w := cmd.OutOrStdout()

	// Parse arguments
	branchPrefix := args[0]
	if err := validateBranchName(branchPrefix); err != nil {
		return err
	}

	var startPoint string
	if len(args) > 1 {
		startPoint = args[1]
	}

	// Check tmux availability
	if !tmux.IsTmuxAvailable() {
		return fmt.Errorf("tmux is not installed. Install with: brew install tmux (macOS) or apt install tmux (Linux)")
	}

	// Validate layout
	if err := validateLayout(cfg.layout); err != nil {
		return err
	}

	// Validate count
	if cfg.count < 1 {
		return fmt.Errorf("count must be at least 1")
	}

	// Get repository information
	repo, err := gitx.GetRepo(ctx, flagRepo)
	if err != nil {
		return fmt.Errorf("failed to get repository information: %w", err)
	}

	// Determine and validate base directory
	baseDir, err := resolveAndValidateBaseDir(cfg.baseDir, repo.Parent)
	if err != nil {
		return err
	}

	// Create worktrees
	fmt.Fprintf(w, "Creating worktrees...\n")
	panes, err := createMultipleWorktrees(ctx, branchPrefix, startPoint, cfg.count, repo, baseDir, w)
	if err != nil {
		return err
	}

	// Setup tmux session
	tmuxName := cfg.sessionName
	if tmuxName == "" {
		// Generate session name from branch prefix
		tmuxName = fmt.Sprintf("gwt-%s-%s", repo.Name, naming.Sanitize(branchPrefix))
	} else {
		// Sanitize user-specified session name to prevent command injection
		tmuxName = naming.Sanitize(tmuxName)
	}

	tm := tmux.NewManager(tmuxName)

	// Kill existing session if it exists
	if tm.SessionExists() {
		if err := tm.KillSession(); err != nil {
			return err
		}
	}

	fmt.Fprintf(w, "\nStarting tmux session...\n")
	tmuxCfg := tmux.SessionConfig{
		SessionName: tmuxName,
		Panes:       panes,
		Layout:      cfg.layout,
		SyncPanes:   cfg.syncPanes,
		NoAttach:    cfg.noAttach,
		Debug:       flagDebug,
	}

	if err := tm.CreateSession(tmuxCfg); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	fmt.Fprintf(w, "✓ Tmux session created: %s\n", tmuxName)

	if cfg.noAttach {
		fmt.Fprintf(w, "\nSession running in background\n")
		fmt.Fprintf(w, "Attach with: tmux attach -t %s\n", tmuxName)
	} else {
		fmt.Fprintf(w, "\nAttaching to tmux session (Ctrl-b d to detach)...\n")
		if err := tm.AttachSession(); err != nil {
			return err
		}
	}

	return nil
}

func createMultipleWorktrees(
	ctx context.Context,
	branchPrefix string,
	startPoint string,
	count int,
	repo *gitx.Repo,
	baseDir string,
	w interface{ Write([]byte) (int, error) },
) ([]tmux.Pane, error) {
	var panes []tmux.Pane

	for i := 1; i <= count; i++ {
		// Generate branch name with number suffix
		branchName := fmt.Sprintf("%s-%d", branchPrefix, i)

		// Check if branch is already in use
		if err := checkBranchNotInUse(ctx, branchName); err != nil {
			return nil, err
		}

		// Sanitize branch name
		sanitized := naming.Sanitize(branchName)

		// Generate worktree path
		worktreePath, err := naming.GenerateWorktreePath(baseDir, repo.Name, sanitized)
		if err != nil {
			return nil, fmt.Errorf("failed to generate worktree path for %s: %w", branchName, err)
		}

		// Check if branch exists
		exists, err := gitx.BranchExists(ctx, branchName)
		if err != nil {
			return nil, fmt.Errorf("failed to check branch existence for %s: %w", branchName, err)
		}

		// Create worktree
		if err := gitx.Add(ctx, worktreePath, branchName, startPoint, !exists); err != nil {
			return nil, fmt.Errorf("failed to create worktree for %s: %w", branchName, err)
		}

		fmt.Fprintf(w, "  ✓ %s -> %s\n", branchName, worktreePath)

		panes = append(panes, tmux.Pane{
			WorktreePath: worktreePath,
			BranchName:   branchName,
		})
	}

	return panes, nil
}
