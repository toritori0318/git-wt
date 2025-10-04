package gitx

import (
	"context"
	"fmt"
	"path/filepath"
)

// Repo represents repository information
type Repo struct {
	Root   string // リポジトリルートの絶対パス
	Name   string // リポジトリ名（ディレクトリ名）
	Parent string // リポジトリルートの親ディレクトリ（兄弟配置用）
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
