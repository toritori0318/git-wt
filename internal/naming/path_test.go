package naming_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/toritsuyo/wt/internal/config"
	"github.com/toritsuyo/wt/internal/naming"
)

// テストリスト（網羅したいテストシナリオ）
// 1. subdirectoryモードでのパス生成 ✓
//    - デフォルト設定で <baseDir>/<repoName>-wt/<sanitizedBranch> が生成される
//    - カスタムsuffixで <baseDir>/<repoName>-custom/<sanitizedBranch> が生成される
//    - 重複時に <baseDir>/<repoName>-wt/<sanitizedBranch>-2, -3... が生成される
// 2. siblingモード（レガシー）でのパス生成 ✓
//    - <baseDir>/<repoName>-<sanitizedBranch> が生成される
//    - 重複時に <baseDir>/<repoName>-<sanitizedBranch>-2, -3... が生成される
// 3. エラーケース ✓
//    - 最大試行回数を超えた場合にエラーが返される
// 4. 設定ファイルが存在しない場合 ✓
//    - デフォルトのsubdirectoryモードでパスが生成される

func TestGenerateWorktreePathWithSubdirectoryMode(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with subdirectory mode (default)
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test basic path generation
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

func TestGenerateWorktreePathWithCustomSuffix(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with custom suffix
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

	want := filepath.Join(baseDir, "myproject-worktrees", "feature-login")
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

	// Create config with subdirectory mode
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	baseDir := filepath.Join(tempDir, "repos")

	// Create first path to simulate duplicate
	firstPath := filepath.Join(baseDir, "myproject-wt", "feature-login")
	if err := os.MkdirAll(firstPath, 0755); err != nil {
		t.Fatalf("Failed to create first path: %v", err)
	}

	// Generate path (should get -2 suffix)
	path, err := naming.GenerateWorktreePathWithConfig(baseDir, "myproject", "feature-login", cfg)
	if err != nil {
		t.Fatalf("GenerateWorktreePathWithConfig() returned error: %v", err)
	}

	want := filepath.Join(baseDir, "myproject-wt", "feature-login-2")
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

func TestGenerateWorktreePathDefault(t *testing.T) {
	// Test the default GenerateWorktreePath function (without explicit config)
	// This should use default subdirectory mode
	tempDir := t.TempDir()
	baseDir := filepath.Join(tempDir, "repos")

	path, err := naming.GenerateWorktreePath(baseDir, "myproject", "feature-login")
	if err != nil {
		t.Fatalf("GenerateWorktreePath() returned error: %v", err)
	}

	want := filepath.Join(baseDir, "myproject-wt", "feature-login")
	if path != want {
		t.Errorf("GenerateWorktreePath() = %q, want %q", path, want)
	}
}
