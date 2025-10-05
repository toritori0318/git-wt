package cli

import (
	_ "embed"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

// Embed shell scripts
// embed requires relative paths, so specify paths from internal/cli
var (
	//go:embed hook_bash_core.sh
	bashHook string

	//go:embed hook_zsh_core.sh
	zshHook string

	//go:embed hook_fish.fish
	fishHook string
)

var supportedShells = []string{"bash", "zsh", "fish"}

// UnsupportedShellError represents an error when an unsupported shell is specified
type UnsupportedShellError struct {
	Shell           string
	SupportedShells []string
}

func (e *UnsupportedShellError) Error() string {
	return fmt.Sprintf("unsupported shell: %s\nSupported shells: %s",
		e.Shell, strings.Join(e.SupportedShells, ", "))
}

type hookCmdConfig struct {
	// Future extensions (e.g., --format, --output-file, etc.)
}

func newHookCmd() *cobra.Command {
	cfg := &hookCmdConfig{}

	cmd := &cobra.Command{
		Use:   "hook <shell>",
		Short: "Output shell hook scripts",
		Long: `Output shell hook scripts to stdout.

To enable actual directory navigation with wt go command,
this script must be added to your shell configuration file.

Supported shells: bash, zsh, fish

Examples:
  # Bash
  wt hook bash >> ~/.bashrc
  source ~/.bashrc

  # Zsh
  wt hook zsh >> ~/.zshrc
  source ~/.zshrc

  # Fish
  wt hook fish > ~/.config/fish/functions/wt.fish
  exec fish`,
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return supportedShells, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(c *cobra.Command, args []string) error {
			return runHookWithConfig(c, args, cfg)
		},
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	return cmd
}

var hookCmd = newHookCmd()

func init() {
	hookCmd = newHookCmd()
	rootCmd.AddCommand(hookCmd)
}

func runHookWithConfig(cmd *cobra.Command, args []string, cfg *hookCmdConfig) error {
	shell := args[0]

	// Validate shell
	if err := validateShell(shell); err != nil {
		return err
	}

	// Get shell script
	script, err := getShellScript(shell)
	if err != nil {
		return err
	}

	// Output script
	printHookScript(cmd.OutOrStdout(), script)

	return nil
}

func validateShell(shell string) error {
	normalizedShell := strings.ToLower(strings.TrimSpace(shell))

	for _, s := range supportedShells {
		if normalizedShell == s {
			return nil
		}
	}

	return &UnsupportedShellError{
		Shell:           shell,
		SupportedShells: supportedShells,
	}
}

func getShellScript(shell string) (string, error) {
	normalizedShell := strings.ToLower(strings.TrimSpace(shell))

	switch normalizedShell {
	case "bash":
		return bashHook, nil
	case "zsh":
		return zshHook, nil
	case "fish":
		return fishHook, nil
	default:
		// Should not reach here as validateShell already checked
		return "", &UnsupportedShellError{
			Shell:           shell,
			SupportedShells: supportedShells,
		}
	}
}

func printHookScript(w io.Writer, script string) {
	fmt.Fprint(w, script)
}
