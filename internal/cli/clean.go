package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toritsuyo/gwt/internal/gitx"
)

// NoRemovableWorktreesError represents an error when no removable worktrees are found
type NoRemovableWorktreesError struct{}

func (e *NoRemovableWorktreesError) Error() string {
	return "no removable worktrees found (main worktree cannot be removed)"
}

// WorktreeRemovalCancelledError represents an error when removal is cancelled
type WorktreeRemovalCancelledError struct{}

func (e *WorktreeRemovalCancelledError) Error() string {
	return "worktree removal cancelled"
}

type cleanCmdConfig struct {
	force      bool
	keepBranch bool
	yes        bool
}

func newCleanCmd() *cobra.Command {
	cfg := &cleanCmdConfig{}

	cmd := &cobra.Command{
		Use:   "clean [query]",
		Short: "Remove worktrees",
		Long: `Remove worktrees.

If query is not specified, select interactively.
After removal, prompts to delete the branch (can be suppressed with --keep-branch).

Warning: Main worktree (repository root) cannot be removed.

Options:
  --force        Force removal even with uncommitted changes
  --keep-branch  Keep the branch
  --yes          Skip all confirmations`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runCleanWithConfig(c, args, cfg)
		},
	}

	cmd.Flags().BoolVar(&cfg.force, "force", false, "Force removal even with uncommitted changes (WARNING: may lose work)")
	cmd.Flags().BoolVar(&cfg.keepBranch, "keep-branch", false, "Keep the branch")
	cmd.Flags().BoolVar(&cfg.yes, "yes", false, "Skip all confirmations")

	return cmd
}

var cleanCmd = newCleanCmd()

func init() {
	cleanCmd = newCleanCmd()
	rootCmd.AddCommand(cleanCmd)
}

func runCleanWithConfig(cmd *cobra.Command, args []string, cfg *cleanCmdConfig) error {
	ctx := cmd.Context()
	w := cmd.OutOrStdout()

	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	// Get removable worktrees
	validWorktrees, items, err := getRemovableWorktrees(ctx)
	if err != nil {
		return err
	}

	// Select worktree to remove
	selectedIndex, err := selectWorktreeByQueryOrInteractive(items, query, "Select worktree to remove", false)
	if err != nil {
		return err
	}

	selected := validWorktrees[selectedIndex]

	// Confirm removal
	if !cfg.yes {
		if !confirmRemoval(w, selected) {
			return &WorktreeRemovalCancelledError{}
		}
	}

	// Remove worktree
	if err := removeWorktree(ctx, w, selected, cfg); err != nil {
		return err
	}

	// Handle branch deletion
	if err := handleBranchDeletion(ctx, w, selected, cfg); err != nil {
		return err
	}

	// Clean up stale worktree administrative files
	_ = gitx.Prune(ctx) // Ignore error: prune is best-effort cleanup

	return nil
}

func getRemovableWorktrees(ctx context.Context) ([]gitx.Worktree, []string, error) {
	// Get worktree list
	worktrees, err := gitx.List(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get worktrees: %w", err)
	}

	if len(worktrees) == 0 {
		return nil, nil, &NoWorktreesError{}
	}

	// Get repository root
	repo, err := gitx.GetRepo(ctx, flagRepo)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get repository information: %w", err)
	}

	// Create list for display (exclude main worktree)
	var items []string
	var validWorktrees []gitx.Worktree

	for _, wt := range worktrees {
		// Skip main worktree
		if wt.Path == repo.Root {
			continue
		}

		branch := formatBranch(wt)
		items = append(items, fmt.Sprintf("%s\t%s", branch, wt.Path))
		validWorktrees = append(validWorktrees, wt)
	}

	if len(validWorktrees) == 0 {
		return nil, nil, &NoRemovableWorktreesError{}
	}

	return validWorktrees, items, nil
}

func confirmRemoval(w io.Writer, wt gitx.Worktree) bool {
	printRemovalConfirmation(w, wt)
	return confirm("Are you sure?")
}

func removeWorktree(ctx context.Context, w io.Writer, wt gitx.Worktree, cfg *cleanCmdConfig) error {
	if err := gitx.Remove(ctx, wt.Path, cfg.force); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	printRemovalSuccess(w, wt.Path, flagQuiet)
	return nil
}

func handleBranchDeletion(ctx context.Context, w io.Writer, wt gitx.Worktree, cfg *cleanCmdConfig) error {
	if cfg.keepBranch || wt.Branch == "" {
		return nil
	}

	// Check if branch is in use
	inUse, err := gitx.IsUsingBranch(ctx, wt.Branch, wt.Path)
	if err != nil {
		return fmt.Errorf("failed to check branch usage: %w", err)
	}

	if inUse {
		printBranchInUseWarning(w, wt.Branch, flagQuiet)
		return nil
	}

	// Ask user if they want to delete the branch
	shouldDelete := cfg.yes || confirm(fmt.Sprintf("Also delete branch '%s'?", wt.Branch))
	if !shouldDelete {
		return nil
	}

	// Check if branch is merged and determine if force delete is needed
	forceDelete, shouldProceed := shouldForceDeleteBranch(ctx, w, wt.Branch, cfg.yes)
	if !shouldProceed {
		printBranchKeptMessage(w, wt.Branch, flagQuiet)
		return nil
	}

	// Delete branch
	if err := gitx.DeleteBranch(ctx, wt.Branch, forceDelete); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	printBranchDeletionSuccess(w, wt.Branch, flagQuiet)
	return nil
}

func shouldForceDeleteBranch(ctx context.Context, w io.Writer, branch string, autoYes bool) (forceDelete bool, shouldProceed bool) {
	merged, err := gitx.IsBranchMerged(ctx, branch)
	if err != nil {
		if !flagQuiet {
			fmt.Fprintf(w, "Warning: failed to check if branch is merged: %v\n", err)
		}
		merged = false
	}

	if merged {
		return false, true
	}

	printBranchNotMergedWarning(w, branch)
	if autoYes {
		return true, true
	}

	if confirm("Force delete? (git branch -D)") {
		return true, true
	}

	return false, false
}

// Output functions

func printRemovalConfirmation(w io.Writer, wt gitx.Worktree) {
	fmt.Fprintf(w, "The following worktree will be removed:\n")
	fmt.Fprintf(w, "  Path: %s\n", wt.Path)
	if wt.Branch != "" {
		fmt.Fprintf(w, "  Branch: %s\n", wt.Branch)
	}
}

func printRemovalSuccess(w io.Writer, path string, quiet bool) {
	if quiet {
		return
	}
	fmt.Fprintf(w, "✓ Worktree removed: %s\n", path)
}

func printBranchDeletionSuccess(w io.Writer, branch string, quiet bool) {
	if quiet {
		return
	}
	fmt.Fprintf(w, "✓ Branch deleted: %s\n", branch)
}

func printBranchInUseWarning(w io.Writer, branch string, quiet bool) {
	if quiet {
		return
	}
	fmt.Fprintf(w, "⚠ Branch '%s' is in use by other worktrees, keeping it\n", branch)
}

func printBranchNotMergedWarning(w io.Writer, branch string) {
	fmt.Fprintf(w, "⚠ Branch '%s' is not merged\n", branch)
}

func printBranchKeptMessage(w io.Writer, branch string, quiet bool) {
	if quiet {
		return
	}
	fmt.Fprintf(w, "Branch '%s' will be kept\n", branch)
}

// confirm prompts user for confirmation
func confirm(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (y/N): ", message)

	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}
