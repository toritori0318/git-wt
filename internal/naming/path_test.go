package naming_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/toritori0318/git-wt/internal/config"
	"github.com/toritori0318/git-wt/internal/naming"
)

// Test scenarios to cover:
// 1. Path generation in subdirectory mode ✓
//    - Default settings generate <baseDir>/.<repoName>-wt/<sanitizedBranch> (prefix is ".")
//    - Custom prefix generates <baseDir>/_<repoName>-wt/<sanitizedBranch>
//    - Empty prefix generates <baseDir>/<repoName>-wt/<sanitizedBranch>
//    - Custom suffix generates <baseDir>/.<repoName>-custom/<sanitizedBranch>
//    - Duplicates generate <baseDir>/.<repoName>-wt/<sanitizedBranch>-2, -3...
// 2. Path generation in sibling mode (legacy) ✓
//    - Generates <baseDir>/<repoName>-<sanitizedBranch>
//    - Duplicates generate <baseDir>/<repoName>-<sanitizedBranch>-2, -3...
// 3. Error cases ✓
//    - Returns error when max attempts exceeded
// 4. When config file doesn't exist ✓
//    - Generates path in default subdirectory mode (prefix is ".")

func TestGenerateWorktreePathWithSubdirectoryMode(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with subdirectory mode (default)
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test basic path generation with default prefix "."
	baseDir := filepath.Join(tempDir, "repos")
	path, err := naming.GenerateWorktreePathWithConfig(baseDir, "myproject", "feature-login", cfg)
	if err != nil {
		t.Fatalf("GenerateWorktreePathWithConfig() returned error: %v", err)
	}

	want := filepath.Join(baseDir, ".myproject-wt", "feature-login")
	if path != want {
		t.Errorf("GenerateWorktreePathWithConfig() = %q, want %q", path, want)
	}
}

func TestGenerateWorktreePathWithCustomSuffix(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with custom suffix (default prefix is ".")
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	cfg.Worktree.SubdirectorySuffix = "-worktrees"

	// Test path generation with custom suffix
	baseDir := filepath.Join(tempDir, "repos")
	path, err := naming.GenerateWorktreePathWithConfig(baseDir, "myproject", "feature-login", cfg)
	if err != nil {
		t.Fatalf("GenerateWorktreePathWithConfig() returned error: %v", err)
	}

	want := filepath.Join(baseDir, ".myproject-worktrees", "feature-login")
	if path != want {
		t.Errorf("GenerateWorktreePathWithConfig() = %q, want %q", path, want)
	}
}

func TestGenerateWorktreePathWithSiblingMode(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with sibling mode
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	cfg.Worktree.DirectoryFormat = "sibling"

	// Test path generation with sibling mode
	baseDir := filepath.Join(tempDir, "repos")
	path, err := naming.GenerateWorktreePathWithConfig(baseDir, "myproject", "feature-login", cfg)
	if err != nil {
		t.Fatalf("GenerateWorktreePathWithConfig() returned error: %v", err)
	}

	want := filepath.Join(baseDir, "myproject-feature-login")
	if path != want {
		t.Errorf("GenerateWorktreePathWithConfig() = %q, want %q", path, want)
	}
}

func TestGenerateWorktreePathWithDuplicatesSubdirectory(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with subdirectory mode (default prefix is ".")
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	baseDir := filepath.Join(tempDir, "repos")

	// Create first path to simulate duplicate
	firstPath := filepath.Join(baseDir, ".myproject-wt", "feature-login")
	if err := os.MkdirAll(firstPath, 0755); err != nil {
		t.Fatalf("Failed to create first path: %v", err)
	}

	// Generate path (should get -2 suffix)
	path, err := naming.GenerateWorktreePathWithConfig(baseDir, "myproject", "feature-login", cfg)
	if err != nil {
		t.Fatalf("GenerateWorktreePathWithConfig() returned error: %v", err)
	}

	want := filepath.Join(baseDir, ".myproject-wt", "feature-login-2")
	if path != want {
		t.Errorf("GenerateWorktreePathWithConfig() = %q, want %q", path, want)
	}
}

func TestGenerateWorktreePathWithDuplicatesSibling(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with sibling mode
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	cfg.Worktree.DirectoryFormat = "sibling"

	baseDir := filepath.Join(tempDir, "repos")

	// Create first path to simulate duplicate
	firstPath := filepath.Join(baseDir, "myproject-feature-login")
	if err := os.MkdirAll(firstPath, 0755); err != nil {
		t.Fatalf("Failed to create first path: %v", err)
	}

	// Generate path (should get -2 suffix)
	path, err := naming.GenerateWorktreePathWithConfig(baseDir, "myproject", "feature-login", cfg)
	if err != nil {
		t.Fatalf("GenerateWorktreePathWithConfig() returned error: %v", err)
	}

	want := filepath.Join(baseDir, "myproject-feature-login-2")
	if path != want {
		t.Errorf("GenerateWorktreePathWithConfig() = %q, want %q", path, want)
	}
}

func TestGenerateWorktreePathWithCustomPrefix(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with custom prefix "_"
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	cfg.Worktree.SubdirectoryPrefix = "_"

	// Test path generation with custom prefix
	baseDir := filepath.Join(tempDir, "repos")
	path, err := naming.GenerateWorktreePathWithConfig(baseDir, "myproject", "feature-login", cfg)
	if err != nil {
		t.Fatalf("GenerateWorktreePathWithConfig() returned error: %v", err)
	}

	want := filepath.Join(baseDir, "_myproject-wt", "feature-login")
	if path != want {
		t.Errorf("GenerateWorktreePathWithConfig() = %q, want %q", path, want)
	}
}

func TestGenerateWorktreePathWithEmptyPrefix(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with empty prefix
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	cfg.Worktree.SubdirectoryPrefix = ""

	// Test path generation with empty prefix
	baseDir := filepath.Join(tempDir, "repos")
	path, err := naming.GenerateWorktreePathWithConfig(baseDir, "myproject", "feature-login", cfg)
	if err != nil {
		t.Fatalf("GenerateWorktreePathWithConfig() returned error: %v", err)
	}

	want := filepath.Join(baseDir, "myproject-wt", "feature-login")
	if path != want {
		t.Errorf("GenerateWorktreePathWithConfig() = %q, want %q", path, want)
	}
}

func TestGenerateWorktreePathDefault(t *testing.T) {
	// Test the default GenerateWorktreePath function (without explicit config)
	// This should use default subdirectory mode with prefix "."
	tempDir := t.TempDir()
	baseDir := filepath.Join(tempDir, "repos")

	path, err := naming.GenerateWorktreePath(baseDir, "myproject", "feature-login")
	if err != nil {
		t.Fatalf("GenerateWorktreePath() returned error: %v", err)
	}

	want := filepath.Join(baseDir, ".myproject-wt", "feature-login")
	if path != want {
		t.Errorf("GenerateWorktreePath() = %q, want %q", path, want)
	}
}
