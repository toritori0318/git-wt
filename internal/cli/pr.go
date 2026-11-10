package cli

import (
	"fmt"
	"io"
	"os"
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
	yes bool
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

Navigation:
  By default, prints only the worktree path so the shell function can cd there.
  Pass --no-cd to stay in the current directory and print a summary instead.

Existing Branch Handling:
  - If the branch is already checked out in a worktree:
    - Default: ask before navigating. Pressing Enter accepts (Y/n)
    - --yes or --quiet: navigate without asking
    - With --no-cd: print the path and exit (does not navigate, even with --yes)
  - If the branch exists locally but is not in a worktree:
    - Default: ask before reusing it. Pressing Enter accepts (Y/n)
    - --yes or --quiet: reuse it without asking

Prerequisites:
  - GitHub CLI (gh) must be installed
  - Must be authenticated with gh auth login

Examples:
  wt pr 123                          # Review PR #123 and navigate there
  wt pr 123 --branch review/pr-123   # Specify custom local branch name
  wt pr 123 --no-cd                  # Create the worktree but stay put
  wt pr 123 --no-cd --yes            # Stay put and skip confirmation prompts`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runPRWithConfig(c, args, cfg)
		},
	}

	cmd.Flags().BoolVar(&cfg.yes, "yes", false, "Skip confirmation prompts and accept them")

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

	// Navigation is the default; --no-cd stays put and does not need the shell function.
	if err := checkShellFunction(!flagNoCD); err != nil {
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

	// pathOnly is the default: stdout is a single path for the shell function to cd.
	pathOnly := !flagNoCD

	// Fetch PR info
	printPRProgress(w, "Fetching PR #%d info...\n", prNumber, pathOnly, flagQuiet)
	prInfo, err := ghx.GetPRInfo(prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR info: %w", err)
	}

	printPRInfo(w, prInfo, pathOnly, flagQuiet)

	// Determine local branch name
	localBranch := flagBranch
	if localBranch == "" {
		localBranch = prInfo.HeadRefName
	}

	// Validate branch name
	if err := validateBranchName(localBranch); err != nil {
		return fmt.Errorf("invalid branch name '%s': %w", localBranch, err)
	}

	// Check if branch is already in use by a worktree
	existingWT, err := gitx.FindWorktreeByBranch(ctx, localBranch)
	if err != nil {
		return fmt.Errorf("failed to search worktrees: %w", err)
	}

	if existingWT != nil {
		// Already checked out. --no-cd reports the path and stays put.
		// Otherwise ask, with Enter meaning yes. --yes and --quiet skip the question.
		if flagNoCD {
			fmt.Fprintf(w, "Branch '%s' is already in use by worktree: %s\n", localBranch, existingWT.Path)
			fmt.Fprintf(w, "Path: %s\n", existingWT.Path)
			return nil
		}
		printExistingWorktreeNotice(os.Stderr, localBranch, flagQuiet)
		if !cfg.yes && !flagQuiet && !confirmDefaultYes("Navigate to existing worktree?") {
			return fmt.Errorf("operation cancelled")
		}
		fmt.Fprintln(w, existingWT.Path)
		return nil
	}

	// Check if branch already exists locally
	branchExists, err := gitx.BranchExists(ctx, localBranch)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if branchExists {
		// Ask before reusing the branch. Enter means yes. --yes and --quiet skip the question.
		if confirmed, err := confirmUseExisting(os.Stderr, localBranch, cfg.yes, flagQuiet); err != nil {
			return err
		} else if !confirmed {
			return fmt.Errorf("operation cancelled")
		}
	}

	// Determine remote and setup temporary remote if needed
	remote, tempRemote, err := determineRemote(w, flagRemote, prInfo, prNumber, pathOnly, flagQuiet)
	if err != nil {
		return err
	}

	// Ensure temporary remote cleanup
	if tempRemote != "" {
		defer func() {
			printPRProgress(w, "Removing temporary remote: %s\n", tempRemote, pathOnly, flagQuiet)
			_ = ghx.RemoveRemote(tempRemote) // Ignore error: cleanup is best-effort
		}()
	}

	// Fetch branch
	printPRProgress(w, "Fetching branch: %s/%s -> %s\n", remote, prInfo.HeadRefName, localBranch, pathOnly, flagQuiet)
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
	printPRProgress(w, "Creating worktree: %s\n", worktreePath, pathOnly, flagQuiet)
	if err := gitx.Add(ctx, worktreePath, localBranch, "", false); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	// Output result
	printPRSuccess(w, worktreePath, prNumber, localBranch, pathOnly, flagQuiet)

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

// printExistingWorktreeNotice writes the "already in use" notice to w.
// Callers must pass os.Stderr, not the command's stdout writer: when navigating,
// stdout is captured by the shell wrapper via command substitution, so
// anything written there would silently vanish instead of reaching the
// terminal.
func printExistingWorktreeNotice(w io.Writer, branch string, quiet bool) {
	if quiet {
		return
	}
	fmt.Fprintf(w, "Branch '%s' is already in use by worktree.\n", branch)
}

// confirmUseExisting asks whether to reuse a local branch that is not in a worktree.
// skip (--yes) and quiet accept without asking. The notice goes to w, which callers
// must set to stderr: navigating commands capture stdout for the shell.
func confirmUseExisting(w io.Writer, branch string, skip, quiet bool) (bool, error) {
	if skip || quiet {
		return true, nil
	}
	fmt.Fprintf(w, "Branch '%s' already exists locally.\n", branch)
	return confirmDefaultYes("Create new worktree using existing branch?"), nil
}

func determineRemote(w io.Writer, userRemote string, prInfo *ghx.PRInfo, prNumber int, pathOnly, quiet bool) (remote, tempRemote string, err error) {
	if userRemote != "" {
		return userRemote, "", nil
	}

	if prInfo.IsCrossRepository {
		// For fork PRs, add temporary remote if needed
		if !ghx.RemoteExists(prInfo.HeadOwner) {
			tempRemote = fmt.Sprintf("wt-pr-%d", prNumber)
			printPRProgress(w, "Adding temporary remote: %s (%s/%s)\n", tempRemote, prInfo.HeadOwner, prInfo.HeadRepo, pathOnly, quiet)
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
	// Extract pathOnly and quiet from the end of args
	if len(args) < 2 {
		return
	}
	pathOnly, ok1 := args[len(args)-2].(bool)
	quiet, ok2 := args[len(args)-1].(bool)
	if !ok1 || !ok2 {
		return
	}

	if pathOnly || quiet {
		return
	}

	fmt.Fprintf(w, format, args[:len(args)-2]...)
}

func printPRInfo(w io.Writer, prInfo *ghx.PRInfo, pathOnly, quiet bool) {
	if pathOnly || quiet {
		return
	}
	fmt.Fprintf(w, "  Branch: %s\n", prInfo.HeadRefName)
	fmt.Fprintf(w, "  Owner: %s\n", prInfo.HeadOwner)
}

func printPRSuccess(w io.Writer, worktreePath string, prNumber int, localBranch string, pathOnly, quiet bool) {
	if pathOnly {
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
