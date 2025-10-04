package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/toritsuyo/gwt/internal/gitx"
)

var (
	// Global flags
	flagRepo  string
	flagQuiet bool
	flagDebug bool

	// Version information (set by main package)
	versionInfo = "dev"
	commitInfo  = "unknown"
	dateInfo    = "unknown"
)

// ExitCodeError wraps an error with an exit code
type ExitCodeError struct {
	Code int
	Err  error
}

func (e *ExitCodeError) Error() string { return e.Err.Error() }
func (e *ExitCodeError) Unwrap() error { return e.Err }

// SetVersionInfo sets version information from main package
func SetVersionInfo(version, commit, date string) {
	versionInfo = version
	commitInfo = commit
	dateInfo = date
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}

var rootCmd = &cobra.Command{
	Use:   "gwt",
	Short: "Git worktree helper CLI",
	Long: `gwt is a CLI tool that wraps git worktree commands with conventions and shortcuts.
It manages worktrees in sibling directories with automatic naming conventions.`,
	Version: "dev",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set debug mode
		if flagDebug {
			gitx.Debug = true
		}

		// Check if git command is available
		if err := gitx.CheckGitInstalled(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagRepo, "repo", "", "Manually specify repository root path")
	rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", false, "Minimal output")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Debug mode (show command execution)")

	// Set custom help template with passthrough section only for root
	rootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}{{if not .HasParent}}

Git Worktree Passthrough:
  Any unknown command will be passed through to 'git worktree'.

  Examples:
    gwt list              -> git worktree list
    gwt add <path> <ref>  -> git worktree add <path> <ref>
    gwt remove <path>     -> git worktree remove <path>
    gwt lock <path>       -> git worktree lock <path>
    gwt prune             -> git worktree prune
{{end}}`)

	// Register subcommands
	rootCmd.AddCommand(newCmd)
}

// Execute runs the root command
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		// Pass through to git worktree for unknown command/flag errors
		if shouldPassthrough(err) {
			if pe := passthroughToGitWorktree(rootCmd, os.Args[1:]); pe != nil {
				return pe
			}
			return nil
		}
		fmt.Fprintln(rootCmd.ErrOrStderr(), err)
		return err
	}
	return nil
}

// shouldPassthrough checks if the error should trigger passthrough to git worktree
func shouldPassthrough(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unknown command") ||
		strings.Contains(msg, "unknown flag") ||
		strings.Contains(msg, "unknown shorthand flag")
}

// filterPassthroughArgs removes internal flags from args before passing to git worktree
func filterPassthroughArgs(args []string) []string {
	out := make([]string, 0, len(args))
	skipNext := false
	for i := 0; i < len(args); i++ {
		if skipNext {
			skipNext = false
			continue
		}
		a := args[i]
		switch {
		// Stop stripping after --
		case a == "--":
			out = append(out, args[i+1:]...)
			return out

		// Boolean persistent flags (do not forward)
		case a == "--debug", a == "--quiet":
			continue

		// Value persistent flag forms
		case strings.HasPrefix(a, "--repo="):
			continue
		case a == "--repo":
			// Skip the value token
			skipNext = true
			continue

		default:
			out = append(out, a)
		}
	}
	return out
}

// passthroughToGitWorktree passes unknown commands to git worktree
func passthroughToGitWorktree(cmd *cobra.Command, rawArgs []string) error {
	// Resolve git path
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git command not found: %w", err)
	}

	args := append([]string{"worktree"}, filterPassthroughArgs(rawArgs)...)

	// Context that cancels on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	c := exec.CommandContext(ctx, gitPath, args...)
	c.Stdin = cmd.InOrStdin()
	c.Stdout = cmd.OutOrStdout()
	c.Stderr = cmd.ErrOrStderr()

	if flagDebug {
		fmt.Fprintf(cmd.ErrOrStderr(), "+ %s %s\n", gitPath, strings.Join(args, " "))
	}

	if err := c.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &ExitCodeError{Code: exitErr.ExitCode(), Err: err}
		}
		return err
	}

	return nil
}
