package gitx

import (
	"context"
	"fmt"
	"path/filepath"
)

// Repo represents repository information
type Repo struct {
	Root   string // Absolute path to repository root
	Name   string // Repository name (directory name)
	Parent string // Parent directory of repository root (for sibling placement)
}

// GetRepo returns repository information for the current or specified directory
func GetRepo(ctx context.Context, dir string) (*Repo, error) {
	root, err := RunGitInDir(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
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
