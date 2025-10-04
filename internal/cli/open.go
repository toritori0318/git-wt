package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/toritsuyo/gwt/internal/editor"
	"github.com/toritsuyo/gwt/internal/gitx"
)

type openCmdConfig struct {
	editor string
}

func newOpenCmd() *cobra.Command {
	cfg := &openCmdConfig{}

	cmd := &cobra.Command{
		Use:   "open [query]",
		Short: "Open worktrees in editor",
		Long: `Open worktrees in editor.

If query is not specified, select interactively.
Editor is determined by the following priority:
  1. --editor flag
  2. GWT_EDITOR environment variable
  3. VISUAL environment variable
  4. EDITOR environment variable
  5. code, idea, subl, vim, vi (in order of availability)
  6. macOS: open, Linux: xdg-open

Examples:
  gwt open                      # Select interactively and open with default editor
  gwt open feature              # Open worktree containing "feature"
  gwt open --editor code main   # Open main with VS Code`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runOpenWithConfig(c, args, cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.editor, "editor", "", "Specify editor to use")
	return cmd
}

var openCmd = newOpenCmd()

func init() {
	openCmd = newOpenCmd()
	rootCmd.AddCommand(openCmd)
}

func runOpenWithConfig(cmd *cobra.Command, args []string, cfg *openCmdConfig) error {
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

	// Create display items (reuse from go.go)
	items := createDisplayItems(worktrees)

	// Select worktree
	selectedIndex, err := selectWorktreeByQueryOrInteractive(items, query, "Select worktree to open")
	if err != nil {
		return err
	}

	// Selected worktree
	selected := worktrees[selectedIndex]

	// Find editor
	editorPath, err := editor.FindEditor(cfg.editor)
	if err != nil {
		return err
	}

	// Output message
	printOpeningMessage(cmd.OutOrStdout(), selected.Path, editorPath, flagQuiet)

	// Open in editor (using resolved path to avoid duplicate FindEditor call)
	if err := editor.OpenWithPath(selected.Path, editorPath); err != nil {
		return err
	}

	return nil
}

func selectWorktreeByQueryOrInteractive(items []string, query string, prompt string) (int, error) {
	if query != "" {
		return selectByQuery(items, query)
	}
	return selectWorktree(items, prompt)
}

func printOpeningMessage(w io.Writer, path, editorPath string, quiet bool) {
	if quiet {
		return
	}
	fmt.Fprintf(w, "Opening %s with '%s'...\n", path, editorPath)
}
