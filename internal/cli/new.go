package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toritori0318/git-wt/internal/gitx"
	"github.com/toritori0318/git-wt/internal/naming"
)

// BranchInUseError represents an error when a branch is already in use
type BranchInUseError struct {
	Branch string
	Path   string
}

func (e *BranchInUseError) Error() string {
	return fmt.Sprintf("branch '%s' is already in use at %s.\nNavigate: wt go %s\nOpen: wt open %s",
		e.Branch, e.Path, e.Branch, e.Branch)
}

type newCmdConfig struct {
	baseDir string
	cd      bool
}

func newNewCmd() *cobra.Command {
	cfg := &newCmdConfig{}

	cmd := &cobra.Command{
		Use:   "new <branch> [<start-point>]",
		Short: "Create a new worktree",
		Long: `Create a new worktree in a sibling directory.

Worktrees are automatically placed using the naming convention <repo>-<branch>.
If an existing branch is specified, that branch will be checked out.
For new branches, they are created from start-point (defaults to current HEAD if omitted).`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 || len(args) > 2 {
				cmd.Help()
				if len(args) == 0 {
					return fmt.Errorf("\nError: missing required argument <branch>")
				}
				return fmt.Errorf("\nError: too many arguments (expected 1-2, got %d)", len(args))
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			return runNewWithConfig(c, args, cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.baseDir, "base-dir", "", "Base directory for worktree placement (defaults to repository parent)")
	cmd.Flags().BoolVar(&cfg.cd, "cd", false, "Output worktree path to stdout after creation (for cd with shell function)")

	return cmd
}

var newCmd = newNewCmd()

func init() {
	newCmd = newNewCmd()
}

func runNewWithConfig(cmd *cobra.Command, args []string, cfg *newCmdConfig) error {
	ctx := cmd.Context()

	// Parse and validate arguments
	branch := args[0]
	if err := validateBranchName(branch); err != nil {
		return err
	}

	var startPoint string
	if len(args) > 1 {
		startPoint = args[1]
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

	// Sanitize branch name
	sanitized := naming.Sanitize(branch)

	// Generate worktree path
	worktreePath, err := naming.GenerateWorktreePath(baseDir, repo.Name, sanitized)
	if err != nil {
		return fmt.Errorf("failed to generate worktree path: %w", err)
	}

	// Check if branch is already in use by another worktree
	if err := checkBranchNotInUse(ctx, branch); err != nil {
		return err
	}

	// Check if branch exists
	branchExists, err := gitx.BranchExists(ctx, branch)
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}

	// Create worktree
	createNewBranch := !branchExists
	if err := gitx.Add(ctx, worktreePath, branch, startPoint, createNewBranch); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	// Success message
	printSuccess(cmd.OutOrStdout(), worktreePath, branch, cfg.cd, flagQuiet)

	return nil
}

func validateBranchName(branch string) error {
	if strings.TrimSpace(branch) == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Check for Git-prohibited patterns
	if strings.Contains(branch, "..") || strings.HasPrefix(branch, "-") {
		return fmt.Errorf("invalid branch name: %s", branch)
	}

	return nil
}

func resolveAndValidateBaseDir(customBaseDir, defaultBaseDir string) (string, error) {
	baseDir := customBaseDir
	if baseDir == "" {
		return defaultBaseDir, nil
	}

	// Validate user-specified base directory
	info, err := os.Stat(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("base directory does not exist: %s", baseDir)
		}
		return "", fmt.Errorf("failed to access base directory: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("base directory is not a directory: %s", baseDir)
	}

	return baseDir, nil
}

func checkBranchNotInUse(ctx context.Context, branch string) error {
	existingWT, err := gitx.FindWorktreeByBranch(ctx, branch)
	if err != nil {
		return fmt.Errorf("failed to search worktrees: %w", err)
	}

	if existingWT != nil {
		return &BranchInUseError{
			Branch: branch,
			Path:   existingWT.Path,
		}
	}

	return nil
}

func printSuccess(w io.Writer, worktreePath, branch string, cdMode, quiet bool) {
	if cdMode {
		fmt.Fprintln(w, worktreePath)
		return
	}

	if quiet {
		return
	}

	fmt.Fprintf(w, "✓ Created worktree\n")
	fmt.Fprintf(w, "  Branch: %s\n", branch)
	fmt.Fprintf(w, "  Path: %s\n", worktreePath)
}
