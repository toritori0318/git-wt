package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// DirectoryFormatSubdirectory uses <repo>-wt/<branch> format
	DirectoryFormatSubdirectory = "subdirectory"
	// DirectoryFormatSibling uses <repo>-<branch> format (legacy)
	DirectoryFormatSibling = "sibling"

	// DefaultDirectoryFormat is the default directory format
	DefaultDirectoryFormat = DirectoryFormatSubdirectory
	// DefaultSubdirectoryPrefix is the default prefix for subdirectory mode
	DefaultSubdirectoryPrefix = "."
	// DefaultSubdirectorySuffix is the default suffix for subdirectory mode
	DefaultSubdirectorySuffix = "-wt"
)

// Config represents the application configuration
type Config struct {
	Worktree WorktreeConfig `yaml:"worktree"`
	path     string         // Path to config file (not serialized)
}

// WorktreeConfig represents worktree-specific configuration
type WorktreeConfig struct {
	DirectoryFormat     string `yaml:"directory_format"`
	SubdirectoryPrefix  string `yaml:"subdirectory_prefix"`
	SubdirectorySuffix  string `yaml:"subdirectory_suffix"`
}

// Load loads configuration from the specified path
// If the file doesn't exist, returns default configuration
func Load(path string) (*Config, error) {
	cfg := &Config{
		path: path,
		Worktree: WorktreeConfig{
			DirectoryFormat:    DefaultDirectoryFormat,
			SubdirectoryPrefix: DefaultSubdirectoryPrefix,
			SubdirectorySuffix: DefaultSubdirectorySuffix,
		},
	}

	// If file doesn't exist, return defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetDirectoryFormat returns the directory format setting
func (c *Config) GetDirectoryFormat() string {
	return c.Worktree.DirectoryFormat
}

// GetSubdirectoryPrefix returns the subdirectory prefix setting
func (c *Config) GetSubdirectoryPrefix() string {
	return c.Worktree.SubdirectoryPrefix
}

// GetSubdirectorySuffix returns the subdirectory suffix setting
func (c *Config) GetSubdirectorySuffix() string {
	return c.Worktree.SubdirectorySuffix
}

// Validate validates the configuration
func (c *Config) Validate() error {
	format := c.Worktree.DirectoryFormat
	if format != DirectoryFormatSubdirectory && format != DirectoryFormatSibling {
		return fmt.Errorf("invalid directory_format: %q (must be %q or %q)",
			format, DirectoryFormatSubdirectory, DirectoryFormatSibling)
	}

	// Validate subdirectory suffix starts with hyphen
	suffix := c.Worktree.SubdirectorySuffix
	if suffix != "" && suffix[0] != '-' {
		return fmt.Errorf("subdirectory_suffix must start with '-', got %q", suffix)
	}

	return nil
}

// SetDirectoryFormat sets and validates the directory format
func (c *Config) SetDirectoryFormat(format string) error {
	if format != DirectoryFormatSubdirectory && format != DirectoryFormatSibling {
		return fmt.Errorf("invalid value for directory_format: %s (must be 'subdirectory' or 'sibling')", format)
	}
	c.Worktree.DirectoryFormat = format
	return nil
}

// SetSubdirectoryPrefix sets the subdirectory prefix
func (c *Config) SetSubdirectoryPrefix(prefix string) error {
	// No validation needed - any string is valid as prefix
	c.Worktree.SubdirectoryPrefix = prefix
	return nil
}

// SetSubdirectorySuffix sets and validates the subdirectory suffix
func (c *Config) SetSubdirectorySuffix(suffix string) error {
	if suffix != "" && suffix[0] != '-' {
		return fmt.Errorf("subdirectory_suffix must start with '-'")
	}
	c.Worktree.SubdirectorySuffix = suffix
	return nil
}

// Save saves the configuration to the file
func (c *Config) Save() error {
	// Validate before saving
	if err := c.Validate(); err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(c.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Reset removes the configuration file
func (c *Config) Reset() error {
	if _, err := os.Stat(c.path); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to do
	}

	if err := os.Remove(c.path); err != nil {
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	return nil
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() (string, error) {
	// Use XDG_CONFIG_HOME if set, otherwise use ~/.config
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		configHome = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configHome, "wt", "config.yaml"), nil
}
