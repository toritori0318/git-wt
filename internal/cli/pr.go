package cli

import (
	"fmt"
	"io"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/toritori0318/git-wt/internal/ghx"
	"github.com/toritori0318/git-wt/internal/gitx"
	"github.com/toritori0318/git-wt/internal/naming"
)

// GhNotFoundError represents an error when GitHub CLI is not found
type GhNotFoundError struct{}

func (e *GhNotFoundError) Error() string {
	return "GitHub CLI (gh) not found\n\nInstallation:\n  macOS: brew install gh\n  Linux: https://cli.github.com/\n\nAuthentication: gh auth login"
}

// InvalidPRNumberError represents an error when PR number is invalid
type InvalidPRNumberError struct {
	Input string
}

func (e *InvalidPRNumberError) Error() string {
	return fmt.Sprintf("invalid PR number: %s", e.Input)
}

type prCmdConfig struct {
	branch string
	remote string
	cd     bool
	force  bool
}

func newPrCmd() *cobra.Command {
	cfg := &prCmdConfig{}

	cmd := &cobra.Command{
		Use:   "pr <pr-number>",
		Short: "Create worktree for PR review",
		Long: `Create worktree for reviewing GitHub Pull Requests.

Uses GitHub CLI (gh) to fetch PR information and creates a dedicated worktree.
Supports PRs from forks.

Branch Naming:
  By default, uses the PR's original branch name (e.g., feature/auth).

Existing Branch Handling:
  - If branch exists in a worktree:
    - Without --cd: Shows info and exits
    - With --cd: Prompts to navigate to existing worktree
  - If branch exists locally (not in worktree): Prompts to create worktree with existing branch
  - Use --force to skip all prompts

Prerequisites:
  - GitHub CLI (gh) must be installed
  - Must be authenticated with gh auth login

Examples:
  wt pr 123                          # Review PR #123 (uses PR's branch name)
  wt pr 123 --branch review/pr-123   # Specify custom local branch name
  wt pr 123 --cd                     # Move immediately after creation
  wt pr 123 --force                  # Skip all prompts, auto-use existing branches`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runPRWithConfig(c, args, cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.branch, "branch", "", "Local branch name (default: PR's original branch name)")
	cmd.Flags().StringVar(&cfg.remote, "remote", "", "Remote name (default: auto-detect)")
	cmd.Flags().BoolVar(&cfg.cd, "cd", false, "Output only worktree path (for shell function)")
	cmd.Flags().BoolVar(&cfg.force, "force", false, "Skip all prompts and use existing branches")

	return cmd
}

var prCmd = newPrCmd()

func init() {
	prCmd = newPrCmd()
	rootCmd.AddCommand(prCmd)
}

func runPRWithConfig(cmd *cobra.Command, args []string, cfg *prCmdConfig) error {
	ctx := cmd.Context()
	w := cmd.OutOrStdout()

	// Check if shell function is configured when using --cd
	if err := checkShellFunction(cfg.cd); err != nil {
		return err
	}

	// Validate PR number
	prNumber, err := validatePRNumber(args[0])
	if err != nil {
		return err
	}

	// Check GitHub CLI
	if !ghx.IsGhAvailable() {
		return &GhNotFoundError{}
	}

	// Get repository info
	repo, err := gitx.GetRepo(ctx, flagRepo)
	if err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}

	// Fetch PR info
	printPRProgress(w, "Fetching PR #%d info...\n", prNumber, cfg.cd, flagQuiet)
	prInfo, err := ghx.GetPRInfo(prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR info: %w", err)
	}

	printPRInfo(w, prInfo, cfg.cd, flagQuiet)

	// Determine local branch name
	localBranch := cfg.branch
	if localBranch == "" {
		localBranch = prInfo.HeadRefName
	}

	// Check if branch is already in use by a worktree
	existingWT, err := gitx.FindWorktreeByBranch(ctx, localBranch)
	if err != nil {
		return fmt.Errorf("failed to search worktrees: %w", err)
	}

	if existingWT != nil {
		// Branch is in use by worktree
		if cfg.cd {
			// With --cd: prompt to navigate (or auto-navigate with --force)
			if cfg.force {
				// Force mode: auto-navigate without prompt
				fmt.Fprintln(w, existingWT.Path)
				return nil
			}
			if !flagQuiet {
				fmt.Fprintf(w, "Branch '%s' is already in use by worktree.\n", localBranch)
			}
			if confirmed, err := confirmNavigate(w, localBranch, existingWT.Path); err != nil {
				return err
			} else if confirmed {
				// User wants to navigate - output path for shell function
				fmt.Fprintln(w, existingWT.Path)
				return nil
			}
			// User declined navigation
			return fmt.Errorf("operation cancelled")
		}
		// Without --cd: show info and exit
		fmt.Fprintf(w, "Branch '%s' is already in use by worktree: %s\n", localBranch, existingWT.Path)
		fmt.Fprintf(w, "Path: %s\n", existingWT.Path)
		return nil
	}

	// Check if branch already exists locally
	branchExists, err := gitx.BranchExists(ctx, localBranch)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if branchExists {
		// Branch exists but not in worktree - prompt to use it (or auto-use with --force)
		if !cfg.force {
			if confirmed, err := confirmUseExisting(w, localBranch, cfg.cd, flagQuiet); err != nil {
				return err
			} else if !confirmed {
				return fmt.Errorf("operation cancelled")
			}
		}
		// User confirmed or force mode - will use existing branch for worktree
	}

	// Determine remote and setup temporary remote if needed
	remote, tempRemote, err := determineRemote(w, cfg.remote, prInfo, prNumber, cfg.cd, flagQuiet)
	if err != nil {
		return err
	}

	// Ensure temporary remote cleanup
	if tempRemote != "" {
		defer func() {
			printPRProgress(w, "Removing temporary remote: %s\n", tempRemote, cfg.cd, flagQuiet)
			_ = ghx.RemoveRemote(tempRemote) // Ignore error: cleanup is best-effort
		}()
	}

	// Fetch branch
	printPRProgress(w, "Fetching branch: %s/%s -> %s\n", remote, prInfo.HeadRefName, localBranch, cfg.cd, flagQuiet)
	if err := ghx.FetchPRBranch(remote, prInfo.HeadRefName, localBranch); err != nil {
		return fmt.Errorf("failed to fetch PR branch: %w", err)
	}

	// Generate worktree path
	sanitized := naming.Sanitize(fmt.Sprintf("pr-%d-%s", prNumber, prInfo.HeadRefName))
	worktreePath, err := naming.GenerateWorktreePath(repo.Parent, repo.Name, sanitized)
	if err != nil {
		return fmt.Errorf("failed to generate worktree path: %w", err)
	}

	// Create worktree
	printPRProgress(w, "Creating worktree: %s\n", worktreePath, cfg.cd, flagQuiet)
	if err := gitx.Add(ctx, worktreePath, localBranch, "", false); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	// Output result
	printPRSuccess(w, worktreePath, prNumber, localBranch, cfg.cd, flagQuiet)

	return nil
}

func validatePRNumber(input string) (int, error) {
	prNumber, err := strconv.Atoi(input)
	if err != nil {
		return 0, &InvalidPRNumberError{Input: input}
	}
	if prNumber <= 0 {
		return 0, &InvalidPRNumberError{Input: input}
	}
	return prNumber, nil
}

// confirmNavigate asks user if they want to navigate to an existing worktree
func confirmNavigate(w io.Writer, branch, path string) (bool, error) {
	confirmed := confirm("Navigate to existing worktree?")
	return confirmed, nil
}

// confirmUseExisting asks user if they want to use an existing branch for new worktree
func confirmUseExisting(w io.Writer, branch string, cdMode, quiet bool) (bool, error) {
	if cdMode || quiet {
		// In cd or quiet mode, assume yes
		return true, nil
	}
	fmt.Fprintf(w, "Branch '%s' already exists locally.\n", branch)
	confirmed := confirm("Create new worktree using existing branch?")
	return confirmed, nil
}

func determineRemote(w io.Writer, userRemote string, prInfo *ghx.PRInfo, prNumber int, cdMode, quiet bool) (remote, tempRemote string, err error) {
	if userRemote != "" {
		return userRemote, "", nil
	}

	if prInfo.IsCrossRepository {
		// For fork PRs, add temporary remote if needed
		if !ghx.RemoteExists(prInfo.HeadOwner) {
			tempRemote = fmt.Sprintf("wt-pr-%d", prNumber)
			printPRProgress(w, "Adding temporary remote: %s (%s/%s)\n", tempRemote, prInfo.HeadOwner, prInfo.HeadRepo, cdMode, quiet)
			if err := ghx.AddRemote(tempRemote, prInfo.HeadOwner, prInfo.HeadRepo); err != nil {
				return "", "", fmt.Errorf("failed to add temporary remote: %w", err)
			}
			return tempRemote, tempRemote, nil
		}
		return prInfo.HeadOwner, "", nil
	}

	// For same-repo PRs, use origin
	return "origin", "", nil
}

func printPRProgress(w io.Writer, format string, args ...interface{}) {
	// Extract cdMode and quiet from the end of args
	if len(args) < 2 {
		return
	}
	cdMode, ok1 := args[len(args)-2].(bool)
	quiet, ok2 := args[len(args)-1].(bool)
	if !ok1 || !ok2 {
		return
	}

	if cdMode || quiet {
		return
	}

	fmt.Fprintf(w, format, args[:len(args)-2]...)
}

func printPRInfo(w io.Writer, prInfo *ghx.PRInfo, cdMode, quiet bool) {
	if cdMode || quiet {
		return
	}
	fmt.Fprintf(w, "  Branch: %s\n", prInfo.HeadRefName)
	fmt.Fprintf(w, "  Owner: %s\n", prInfo.HeadOwner)
}

func printPRSuccess(w io.Writer, worktreePath string, prNumber int, localBranch string, cdMode, quiet bool) {
	if cdMode {
		fmt.Fprintln(w, worktreePath)
		return
	}

	if quiet {
		return
	}

	fmt.Fprintf(w, "\n✓ PR review worktree created\n")
	fmt.Fprintf(w, "  PR: #%d\n", prNumber)
	fmt.Fprintf(w, "  Branch: %s\n", localBranch)
	fmt.Fprintf(w, "  Path: %s\n", worktreePath)
	fmt.Fprintf(w, "\nNavigate: cd %s\n", worktreePath)
	fmt.Fprintf(w, "Or: wt go pr-%d\n", prNumber)
}
