package gitx

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// Repo represents repository information
type Repo struct {
	Root   string // Absolute path to repository root
	Name   string // Repository name (directory name)
	Parent string // Parent directory of repository root (for sibling placement)
}

// getMainWorktreeRoot returns the root of the main worktree
// When called from a worktree, it returns the main repository root, not the worktree path
func getMainWorktreeRoot(ctx context.Context, dir string) (string, error) {
	output, err := RunGitInDir(ctx, dir, "worktree", "list", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("failed to get worktree list: %w", err)
	}

	// Parse the first worktree entry (main worktree)
	// Format: "worktree /path/to/repo"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "worktree ") {
			return strings.TrimPrefix(line, "worktree "), nil
		}
	}

	return "", fmt.Errorf("could not find main worktree in output")
}

// GetRepo returns repository information for the current or specified directory
// If called from a worktree, it returns the main repository information
func GetRepo(ctx context.Context, dir string) (*Repo, error) {
	// First check if we're in a git repository
	_, err := RunGitInDir(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}

	// Get the main worktree root (works from both main repo and worktrees)
	root, err := getMainWorktreeRoot(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get main worktree root: %w", err)
	}

	name := filepath.Base(root)
	parent := filepath.Dir(root)

	return &Repo{
		Root:   root,
		Name:   name,
		Parent: parent,
	}, nil
}

// IsInsideWorktree checks if the current directory is inside a git worktree
func IsInsideWorktree(ctx context.Context, dir string) bool {
	_, err := RunGitInDir(ctx, dir, "rev-parse", "--is-inside-work-tree")
	return err == nil
}
