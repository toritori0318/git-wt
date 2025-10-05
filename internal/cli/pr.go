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
}

func newPrCmd() *cobra.Command {
	cfg := &prCmdConfig{}

	cmd := &cobra.Command{
		Use:   "pr <pr-number>",
		Short: "Create worktree for PR review",
		Long: `Create worktree for reviewing GitHub Pull Requests.

Uses GitHub CLI (gh) to fetch PR information and creates a dedicated worktree.
Supports PRs from forks.

Prerequisites:
  - GitHub CLI (gh) must be installed
  - Must be authenticated with gh auth login

Examples:
  wt pr 123                          # Review PR #123
  wt pr 123 --branch review/pr-123   # Specify local branch name
  wt pr 123 --cd                     # Move immediately after creation`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runPRWithConfig(c, args, cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.branch, "branch", "", "Local branch name (default: wt/pr-<num>)")
	cmd.Flags().StringVar(&cfg.remote, "remote", "", "Remote name (default: auto-detect)")
	cmd.Flags().BoolVar(&cfg.cd, "cd", false, "Output only worktree path (for shell function)")

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
	localBranch := determineLocalBranch(cfg.branch, prNumber)

	// Check if branch is already in use
	if err := checkBranchNotInUse(ctx, localBranch); err != nil {
		return err
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

func determineLocalBranch(userBranch string, prNumber int) string {
	if userBranch != "" {
		return userBranch
	}
	return fmt.Sprintf("wt/pr-%d", prNumber)
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
