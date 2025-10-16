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
	"github.com/toritori0318/git-wt/internal/gitx"
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

// ShellFunctionNotConfiguredError represents an error when shell function is not configured
type ShellFunctionNotConfiguredError struct{}

func (e *ShellFunctionNotConfiguredError) Error() string {
	return `Cannot change directory: shell function not configured.

To enable directory navigation with --cd flag, configure your shell:

  Bash:   echo 'eval "$(wt hook bash)"' >> ~/.bashrc
  Zsh:    echo 'eval "$(wt hook zsh)"' >> ~/.zshrc
  Fish:   wt hook fish > ~/.config/fish/functions/wt.fish

Then restart your shell or run: exec $SHELL`
}

// checkShellFunction checks if shell function is configured when using --cd flag
func checkShellFunction(cdMode bool) error {
	if !cdMode {
		return nil
	}

	// Check if WT_SHELL_FUNCTION environment variable is set
	if os.Getenv("WT_SHELL_FUNCTION") == "" {
		return &ShellFunctionNotConfiguredError{}
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "wt",
	Short: "Git worktree helper CLI",
	Long: `wt is a CLI tool that wraps git worktree commands with conventions and shortcuts.
It manages worktrees in sibling directories with automatic naming conventions.`,
	Version: "dev",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	SilenceErrors: true,
	SilenceUsage:  true,
	// Allow unknown flags to pass through to subcommands
	// This enables arguments like "-wttt" to be used as values
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
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
	// Enable TraverseChildren to parse flags on both root and subcommands
	// This allows subcommands to handle their own arguments correctly
	rootCmd.TraverseChildren = true

	rootCmd.PersistentFlags().StringVar(&flagRepo, "repo", "", "Manually specify repository root path")
	rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", false, "Minimal output")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Debug mode (show command execution)")

	// Disable interspersed flags to allow subcommand arguments that start with '-'
	// This prevents arguments like "-wttt" from being interpreted as global flags
	rootCmd.Flags().SetInterspersed(false)

	// Set custom help template with passthrough section only for root
	rootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}{{if not .HasParent}}

Git Worktree Passthrough:
  Any unknown command will be passed through to 'git worktree'.

  Examples:
    wt list              -> git worktree list
    wt add <path> <ref>  -> git worktree add <path> <ref>
    wt remove <path>     -> git worktree remove <path>
    wt lock <path>       -> git worktree lock <path>
    wt prune             -> git worktree prune
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

	// Don't passthrough if user is trying to use a known subcommand
	// This prevents issues with arguments that look like flags (e.g., "-wttt")
	if len(os.Args) > 1 {
		subcommand := os.Args[1]
		knownSubcommands := []string{"config", "new", "go", "clean", "pr", "open", "hook", "tmux"}
		for _, known := range knownSubcommands {
			if subcommand == known {
				return false
			}
		}
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
