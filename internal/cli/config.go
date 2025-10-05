package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/toritori0318/git-wt/internal/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage wt configuration",
		Long: `Manage wt configuration settings.

Configuration file location: ~/.config/wt/config.yaml

Available settings:
  worktree.directory_format     - "subdirectory" or "sibling"
  worktree.subdirectory_suffix  - Suffix for subdirectory mode (default: "-wt")`,
	}

	// Disable interspersed flags to allow arguments that start with '-'
	// This prevents "-wttt" from being interpreted as a flag
	cmd.Flags().SetInterspersed(false)

	cmd.AddCommand(newConfigListCmd())
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigResetCmd())

	return cmd
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration settings",
		RunE:  runConfigList,
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE:  runConfigGet,
	}
}

func newConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE:  runConfigSet,
		// Allow unknown flags to pass through as values
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	// Disable flag parsing to treat all arguments as positional
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func newConfigResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		RunE:  runConfigReset,
	}
}

func runConfigList(cmd *cobra.Command, args []string) error {
	configPath, err := config.GetDefaultConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	w := cmd.OutOrStdout()
	printConfigList(w, cfg, configPath)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	configPath, err := config.GetDefaultConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	value, err := getConfigValue(cfg, key)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), value)
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	configPath, err := config.GetDefaultConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := setConfigValue(cfg, key, value); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Set %s = %s\n", key, value)
	return nil
}

func runConfigReset(cmd *cobra.Command, args []string) error {
	configPath, err := config.GetDefaultConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Reset(); err != nil {
		return fmt.Errorf("failed to reset config: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "✓ Configuration reset to defaults")
	return nil
}

func printConfigList(w io.Writer, cfg *config.Config, configPath string) {
	// Check if config file exists
	fileStatus := "not found (using defaults)"
	if _, err := os.Stat(configPath); err == nil {
		fileStatus = "found"
	}

	fmt.Fprintf(w, "Configuration file: %s (%s)\n\n", configPath, fileStatus)
	fmt.Fprintln(w, "Settings:")
	fmt.Fprintf(w, "  worktree.directory_format     = %s\n", cfg.GetDirectoryFormat())
	fmt.Fprintf(w, "  worktree.subdirectory_suffix  = %s\n", cfg.GetSubdirectorySuffix())
}

func getConfigValue(cfg *config.Config, key string) (string, error) {
	switch key {
	case "worktree.directory_format":
		return cfg.GetDirectoryFormat(), nil
	case "worktree.subdirectory_suffix":
		return cfg.GetSubdirectorySuffix(), nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

func setConfigValue(cfg *config.Config, key, value string) error {
	switch key {
	case "worktree.directory_format":
		return cfg.SetDirectoryFormat(value)
	case "worktree.subdirectory_suffix":
		return cfg.SetSubdirectorySuffix(value)
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
}

var configCmd = newConfigCmd()

func init() {
	rootCmd.AddCommand(configCmd)
}
