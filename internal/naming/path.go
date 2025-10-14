package naming

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/toritori0318/git-wt/internal/config"
)

// GenerateWorktreePath generates a unique worktree path using default configuration
// Uses subdirectory mode by default: <baseDir>/.<repoName>-wt/<sanitizedBranch>
func GenerateWorktreePath(baseDir, repoName, sanitizedBranch string) (string, error) {
	// Load default config (or from default config path if available)
	configPath, err := config.GetDefaultConfigPath()
	if err != nil {
		// If we can't get config path, use defaults
		cfg := &config.Config{
			Worktree: config.WorktreeConfig{
				DirectoryFormat:    config.DefaultDirectoryFormat,
				SubdirectoryPrefix: config.DefaultSubdirectoryPrefix,
				SubdirectorySuffix: config.DefaultSubdirectorySuffix,
			},
		}
		return GenerateWorktreePathWithConfig(baseDir, repoName, sanitizedBranch, cfg)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		// If config load fails, use defaults
		cfg = &config.Config{
			Worktree: config.WorktreeConfig{
				DirectoryFormat:    config.DefaultDirectoryFormat,
				SubdirectoryPrefix: config.DefaultSubdirectoryPrefix,
				SubdirectorySuffix: config.DefaultSubdirectorySuffix,
			},
		}
	}

	return GenerateWorktreePathWithConfig(baseDir, repoName, sanitizedBranch, cfg)
}

// GenerateWorktreePathWithConfig generates a unique worktree path using the provided configuration
func GenerateWorktreePathWithConfig(baseDir, repoName, sanitizedBranch string, cfg *config.Config) (string, error) {
	const maxAttempts = 100

	if cfg.GetDirectoryFormat() == config.DirectoryFormatSubdirectory {
		// Subdirectory mode: <baseDir>/<prefix><repoName><suffix>/<sanitizedBranch>
		worktreeDir := cfg.GetSubdirectoryPrefix() + repoName + cfg.GetSubdirectorySuffix()
		return generateUniquePathInSubdir(baseDir, worktreeDir, sanitizedBranch, maxAttempts)
	}

	// Sibling mode (legacy): <baseDir>/<repoName>-<sanitizedBranch>
	baseName := fmt.Sprintf("%s-%s", repoName, sanitizedBranch)
	candidate := filepath.Join(baseDir, baseName)

	// Check for duplicates
	if !pathExists(candidate) {
		return candidate, nil
	}

	// Retry with numbered suffix
	for i := 2; i < maxAttempts; i++ {
		candidate = filepath.Join(baseDir, fmt.Sprintf("%s-%d", baseName, i))
		if !pathExists(candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("could not generate unique path after %d attempts", maxAttempts)
}

// generateUniquePathInSubdir generates a unique path in a subdirectory
func generateUniquePathInSubdir(baseDir, worktreeDir, branchName string, maxAttempts int) (string, error) {
	// Base path: <baseDir>/<worktreeDir>/<branchName>
	candidate := filepath.Join(baseDir, worktreeDir, branchName)

	// Check for duplicates
	if !pathExists(candidate) {
		return candidate, nil
	}

	// Retry with numbered suffix on the branch name
	for i := 2; i < maxAttempts; i++ {
		candidate = filepath.Join(baseDir, worktreeDir, fmt.Sprintf("%s-%d", branchName, i))
		if !pathExists(candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("could not generate unique path after %d attempts", maxAttempts)
}

// pathExists checks if a path exists
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
