package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/toritori0318/git-wt/internal/config"
)

// Test scenarios to cover:
// 1. Get default configuration ✓
//    - Returns default values when file doesn't exist
//    - directory_format is "subdirectory"
//    - subdirectory_prefix is "."
//    - subdirectory_suffix is "-wt"
// 2. Load configuration from file
//    - Can load settings from valid YAML file
//    - Can load when directory_format is "sibling"
//    - Can load custom subdirectory_prefix
//    - Can load custom subdirectory_suffix
// 3. Save configuration to file
//    - Can save settings to YAML file
//    - Automatically creates directory if it doesn't exist
// 4. Validate configuration values
//    - Returns error for invalid directory_format (other than "subdirectory" or "sibling")
// 5. Reset configuration
//    - Can delete configuration file
// 6. Get individual configuration values
//    - Can get format with GetDirectoryFormat()
//    - Can get prefix with GetSubdirectoryPrefix()
//    - Can get suffix with GetSubdirectorySuffix()

// TestDefaultConfig tests that default configuration is returned when no config file exists
func TestDefaultConfig(t *testing.T) {
	// Use temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Load config from non-existent file (should return defaults)
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Verify default values
	if got := cfg.GetDirectoryFormat(); got != "subdirectory" {
		t.Errorf("GetDirectoryFormat() = %q, want %q", got, "subdirectory")
	}

	if got := cfg.GetSubdirectoryPrefix(); got != "." {
		t.Errorf("GetSubdirectoryPrefix() = %q, want %q", got, ".")
	}

	if got := cfg.GetSubdirectorySuffix(); got != "-wt" {
		t.Errorf("GetSubdirectorySuffix() = %q, want %q", got, "-wt")
	}
}

// TestLoadFromFile tests loading configuration from a YAML file
func TestLoadFromFile(t *testing.T) {
	tests := []struct {
		name           string
		yamlContent    string
		wantFormat     string
		wantPrefix     string
		wantSuffix     string
		wantErr        bool
	}{
		{
			name: "subdirectory format with default prefix and suffix",
			yamlContent: `worktree:
  directory_format: subdirectory
  subdirectory_prefix: .
  subdirectory_suffix: -wt`,
			wantFormat: "subdirectory",
			wantPrefix: ".",
			wantSuffix: "-wt",
			wantErr:    false,
		},
		{
			name: "sibling format",
			yamlContent: `worktree:
  directory_format: sibling`,
			wantFormat: "sibling",
			wantPrefix: ".", // Default prefix is preserved (not used in sibling mode)
			wantSuffix: "-wt", // Default suffix is preserved (not used in sibling mode)
			wantErr:    false,
		},
		{
			name: "custom subdirectory prefix and suffix",
			yamlContent: `worktree:
  directory_format: subdirectory
  subdirectory_prefix: _
  subdirectory_suffix: -worktrees`,
			wantFormat: "subdirectory",
			wantPrefix: "_",
			wantSuffix: "-worktrees",
			wantErr:    false,
		},
		{
			name: "empty subdirectory prefix",
			yamlContent: `worktree:
  directory_format: subdirectory
  subdirectory_prefix: ""
  subdirectory_suffix: -wt`,
			wantFormat: "subdirectory",
			wantPrefix: "",
			wantSuffix: "-wt",
			wantErr:    false,
		},
		{
			name: "invalid directory format",
			yamlContent: `worktree:
  directory_format: invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config file
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.yaml")

			if err := os.WriteFile(configPath, []byte(tt.yamlContent), 0644); err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			// Load config
			cfg, err := config.Load(configPath)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() error = nil, wantErr = true")
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() returned unexpected error: %v", err)
			}

			// Verify values
			if got := cfg.GetDirectoryFormat(); got != tt.wantFormat {
				t.Errorf("GetDirectoryFormat() = %q, want %q", got, tt.wantFormat)
			}

			if got := cfg.GetSubdirectoryPrefix(); got != tt.wantPrefix {
				t.Errorf("GetSubdirectoryPrefix() = %q, want %q", got, tt.wantPrefix)
			}

			if got := cfg.GetSubdirectorySuffix(); got != tt.wantSuffix {
				t.Errorf("GetSubdirectorySuffix() = %q, want %q", got, tt.wantSuffix)
			}
		})
	}
}

// TestSaveConfig tests saving configuration to file
func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "subdir", "config.yaml")

	// Create config
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Modify config
	cfg.Worktree.DirectoryFormat = "sibling"
	cfg.Worktree.SubdirectoryPrefix = "_"
	cfg.Worktree.SubdirectorySuffix = "-custom"

	// Save config (should create directory if needed)
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	// Load again and verify
	cfg2, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() after Save() returned error: %v", err)
	}

	if got := cfg2.GetDirectoryFormat(); got != "sibling" {
		t.Errorf("After Save/Load: GetDirectoryFormat() = %q, want %q", got, "sibling")
	}

	if got := cfg2.GetSubdirectoryPrefix(); got != "_" {
		t.Errorf("After Save/Load: GetSubdirectoryPrefix() = %q, want %q", got, "_")
	}

	if got := cfg2.GetSubdirectorySuffix(); got != "-custom" {
		t.Errorf("After Save/Load: GetSubdirectorySuffix() = %q, want %q", got, "-custom")
	}
}

// TestResetConfig tests resetting (deleting) configuration
func TestResetConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create and save config
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	// Reset config
	if err := cfg.Reset(); err != nil {
		t.Fatalf("Reset() returned error: %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Errorf("Config file still exists after Reset()")
	}
}

// TestResetNonexistentConfig tests resetting when config file doesn't exist
func TestResetNonexistentConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Reset should not error even if file doesn't exist
	if err := cfg.Reset(); err != nil {
		t.Fatalf("Reset() returned error for non-existent file: %v", err)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		config config.Config
		wantErr bool
	}{
		{
			name: "valid subdirectory format",
			config: config.Config{
				Worktree: config.WorktreeConfig{
					DirectoryFormat:    config.DirectoryFormatSubdirectory,
					SubdirectorySuffix: "-wt",
				},
			},
			wantErr: false,
		},
		{
			name: "valid sibling format",
			config: config.Config{
				Worktree: config.WorktreeConfig{
					DirectoryFormat:    config.DirectoryFormatSibling,
					SubdirectorySuffix: "-wt",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid directory format",
			config: config.Config{
				Worktree: config.WorktreeConfig{
					DirectoryFormat: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid subdirectory suffix (no hyphen)",
			config: config.Config{
				Worktree: config.WorktreeConfig{
					DirectoryFormat:    config.DirectoryFormatSubdirectory,
					SubdirectorySuffix: "wt",
				},
			},
			wantErr: true,
		},
		{
			name: "empty subdirectory suffix is valid",
			config: config.Config{
				Worktree: config.WorktreeConfig{
					DirectoryFormat:    config.DirectoryFormatSubdirectory,
					SubdirectorySuffix: "",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetDirectoryFormat(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid subdirectory",
			value:   config.DirectoryFormatSubdirectory,
			wantErr: false,
		},
		{
			name:    "valid sibling",
			value:   config.DirectoryFormatSibling,
			wantErr: false,
		},
		{
			name:    "invalid format",
			value:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Worktree: config.WorktreeConfig{
					DirectoryFormat:    config.DefaultDirectoryFormat,
					SubdirectorySuffix: config.DefaultSubdirectorySuffix,
				},
			}

			err := cfg.SetDirectoryFormat(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetDirectoryFormat() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && cfg.Worktree.DirectoryFormat != tt.value {
				t.Errorf("DirectoryFormat = %v, want %v", cfg.Worktree.DirectoryFormat, tt.value)
			}
		})
	}
}

func TestSetSubdirectoryPrefix(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid prefix with dot",
			value:   ".",
			wantErr: false,
		},
		{
			name:    "valid prefix with underscore",
			value:   "_",
			wantErr: false,
		},
		{
			name:    "empty prefix is valid",
			value:   "",
			wantErr: false,
		},
		{
			name:    "arbitrary string is valid",
			value:   "prefix-",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Worktree: config.WorktreeConfig{
					DirectoryFormat:    config.DefaultDirectoryFormat,
					SubdirectoryPrefix: config.DefaultSubdirectoryPrefix,
					SubdirectorySuffix: config.DefaultSubdirectorySuffix,
				},
			}

			err := cfg.SetSubdirectoryPrefix(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetSubdirectoryPrefix() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && cfg.Worktree.SubdirectoryPrefix != tt.value {
				t.Errorf("SubdirectoryPrefix = %v, want %v", cfg.Worktree.SubdirectoryPrefix, tt.value)
			}
		})
	}
}

func TestSetSubdirectorySuffix(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid suffix with hyphen",
			value:   "-wt",
			wantErr: false,
		},
		{
			name:    "valid suffix with longer hyphen prefix",
			value:   "-worktrees",
			wantErr: false,
		},
		{
			name:    "invalid suffix without hyphen",
			value:   "wt",
			wantErr: true,
		},
		{
			name:    "empty suffix is valid",
			value:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Worktree: config.WorktreeConfig{
					DirectoryFormat:    config.DefaultDirectoryFormat,
					SubdirectorySuffix: config.DefaultSubdirectorySuffix,
				},
			}

			err := cfg.SetSubdirectorySuffix(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetSubdirectorySuffix() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && cfg.Worktree.SubdirectorySuffix != tt.value {
				t.Errorf("SubdirectorySuffix = %v, want %v", cfg.Worktree.SubdirectorySuffix, tt.value)
			}
		})
	}
}
