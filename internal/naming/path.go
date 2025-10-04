package naming

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateWorktreePath generates a unique worktree path in the parent directory
// Naming convention: <baseDir>/<repoName>-<sanitizedBranch>
// On duplicate: <baseDir>/<repoName>-<sanitizedBranch>-2, -3, ...
func GenerateWorktreePath(baseDir, repoName, sanitizedBranch string) (string, error) {
	const maxAttempts = 100

	// Base path
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

// pathExists checks if a path exists
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
