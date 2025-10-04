package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/toritsuyo/gwt/internal/gitx"
	"github.com/toritsuyo/gwt/internal/selectx"
)

// NoWorktreesError represents an error when no worktrees are found
type NoWorktreesError struct{}

func (e *NoWorktreesError) Error() string {
	return "no worktrees found"
}

// IndexOutOfRangeError represents an error when index is out of range
type IndexOutOfRangeError struct {
	Index int
	Max   int
}

func (e *IndexOutOfRangeError) Error() string {
	return fmt.Sprintf("index out of range: %d (max: %d)", e.Index, e.Max)
}

// NoMatchError represents an error when no matching worktree is found
type NoMatchError struct {
	Query string
}

func (e *NoMatchError) Error() string {
	return fmt.Sprintf("no matching worktree found: %s", e.Query)
}

type goCmdConfig struct {
	noFzf bool
	index int
}

func newGoCmd() *cobra.Command {
	cfg := &goCmdConfig{}

	cmd := &cobra.Command{
		Use:   "go [query]",
		Short: "Navigate between worktrees",
		Long: `Navigate between worktrees.

If query is not specified, select interactively (using fzf or numbered selection).
If query is specified, filter by partial match.

Examples:
  gwt go                    # Interactive selection
  gwt go feature            # Select worktree containing "feature"
  gwt go --quiet feature    # Output path only (for shell function)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runGoWithConfig(c, args, cfg)
		},
	}

	cmd.Flags().BoolVar(&cfg.noFzf, "no-fzf", false, "Don't use fzf")
	cmd.Flags().IntVar(&cfg.index, "index", -1, "Non-interactive mode: select specified index")

	return cmd
}

var goCmd = newGoCmd()

func init() {
	goCmd = newGoCmd()
	rootCmd.AddCommand(goCmd)
}

func runGoWithConfig(cmd *cobra.Command, args []string, cfg *goCmdConfig) error {
	ctx := cmd.Context()

	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	// Get worktree list
	worktrees, err := gitx.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to get worktrees: %w", err)
	}

	if len(worktrees) == 0 {
		return &NoWorktreesError{}
	}

	// Create display items
	items := createDisplayItems(worktrees)

	// Select worktree
	selectedIndex, err := selectWorktreeIndex(worktrees, items, cfg, query)
	if err != nil {
		return err
	}

	// Selected worktree
	selected := worktrees[selectedIndex]

	// Output result
	printGoResult(cmd.OutOrStdout(), &selected, query, flagQuiet)

	return nil
}

func createDisplayItems(worktrees []gitx.Worktree) []string {
	items := make([]string, len(worktrees))
	for i, wt := range worktrees {
		branch := formatBranch(wt)
		items[i] = fmt.Sprintf("%s\t%s", branch, wt.Path)
	}
	return items
}

func formatBranch(wt gitx.Worktree) string {
	if !wt.IsDetached {
		return wt.Branch
	}

	headShort := wt.HEAD
	if len(headShort) > 7 {
		headShort = headShort[:7]
	}
	return fmt.Sprintf("(detached: %s)", headShort)
}

func selectWorktreeIndex(
	worktrees []gitx.Worktree,
	items []string,
	cfg *goCmdConfig,
	query string,
) (int, error) {
	// Case 1: Direct index selection
	if cfg.index >= 0 {
		if cfg.index >= len(worktrees) {
			return 0, &IndexOutOfRangeError{
				Index: cfg.index,
				Max:   len(worktrees) - 1,
			}
		}
		return cfg.index, nil
	}

	// Case 2: Query-based selection
	if query != "" {
		return selectByQuery(items, query, cfg.noFzf)
	}

	// Case 3: Interactive selection
	return selectWorktree(items, "Select worktree", cfg.noFzf)
}

func selectByQuery(items []string, query string, noFzf bool) (int, error) {
	filtered, err := selectx.FilterByQuery(items, query)
	if err != nil {
		return 0, &NoMatchError{Query: query}
	}

	if len(filtered) == 1 {
		// Auto-select if only one match
		return filtered[0].Index, nil
	}

	// Select from multiple matches
	filteredItems := make([]string, len(filtered))
	for i, f := range filtered {
		filteredItems[i] = f.Text
	}

	idx, err := selectWorktree(filteredItems, "Select worktree", noFzf)
	if err != nil {
		return 0, err
	}

	return filtered[idx].Index, nil
}

func selectWorktree(items []string, prompt string, noFzf bool) (int, error) {
	if !noFzf && selectx.IsFzfAvailable() {
		return selectx.SelectWithFzf(items, prompt)
	}
	return selectx.SelectWithPrompt(items, prompt)
}

func printGoResult(w io.Writer, selected *gitx.Worktree, query string, quiet bool) {
	if quiet {
		fmt.Fprintln(w, selected.Path)
		return
	}

	fmt.Fprintf(w, "Destination: %s\n", selected.Path)
	fmt.Fprintf(w, "\nHint: To actually navigate, use shell function\n")
	if query != "" {
		fmt.Fprintf(w, "  gwt go %s\n", query)
	}
}
