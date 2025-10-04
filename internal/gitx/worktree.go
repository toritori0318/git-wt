package gitx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Worktree represents a git worktree
type Worktree struct {
	Path      string // Worktree path
	Branch    string // Branch name (empty if detached)
	HEAD      string // HEAD commit SHA
	IsDetached bool   // Whether in detached HEAD state
	IsLocked  bool   // Whether locked
	IsPrunable bool   // Whether prunable
}

// List returns all worktrees in the repository
func List(ctx context.Context) ([]Worktree, error) {
	output, err := RunGit(ctx, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	return parseWorktreePorcelain(output)
}

// parseWorktreePorcelain parses the output of 'git worktree list --porcelain'
func parseWorktreePorcelain(output string) ([]Worktree, error) {
	var worktrees []Worktree
	var current *Worktree

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if current != nil {
				worktrees = append(worktrees, *current)
				current = nil
			}
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "worktree":
			current = &Worktree{Path: value}
		case "HEAD":
			if current != nil {
				current.HEAD = value
			}
		case "branch":
			if current != nil {
				// branch refs/heads/main -> main
				current.Branch = strings.TrimPrefix(value, "refs/heads/")
			}
		case "detached":
			if current != nil {
				current.IsDetached = true
			}
		case "locked":
			if current != nil {
				current.IsLocked = true
			}
		case "prunable":
			if current != nil {
				current.IsPrunable = true
			}
		}
	}

	// Add the last entry
	if current != nil {
		worktrees = append(worktrees, *current)
	}

	return worktrees, nil
}

// Add creates a new worktree
func Add(ctx context.Context, path, branch, startPoint string, createBranch bool) error {
	args := []string{"worktree", "add"}

	if createBranch {
		args = append(args, "-b", branch)
	}

	args = append(args, path)

	if !createBranch {
		args = append(args, branch)
	} else if startPoint != "" {
		args = append(args, startPoint)
	}

	_, err := RunGit(ctx, args...)
	if err != nil {
		return err
	}

	return nil
}

// Remove removes a worktree
func Remove(ctx context.Context, path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	_, err := RunGit(ctx, args...)
	return err
}

// Prune removes worktree information for deleted directories
func Prune(ctx context.Context) error {
	_, err := RunGit(ctx, "worktree", "prune")
	return err
}

// IsMainWorktree checks if the given path is the main worktree
func IsMainWorktree(ctx context.Context, path string) (bool, error) {
	repo, err := GetRepo(ctx, "")
	if err != nil {
		return false, err
	}

	// Convert path to absolute path for comparison
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}

	return absPath == repo.Root, nil
}

// GetCurrentWorktree returns the worktree for the current directory
func GetCurrentWorktree(ctx context.Context) (*Worktree, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	worktrees, err := List(ctx)
	if err != nil {
		return nil, err
	}

	// Find the worktree containing the current directory
	for _, wt := range worktrees {
		absPath, err := filepath.Abs(wt.Path)
		if err != nil {
			continue
		}

		// Check if cwd is under the worktree path
		relPath, err := filepath.Rel(absPath, cwd)
		if err != nil {
			continue
		}

		if !strings.HasPrefix(relPath, "..") {
			return &wt, nil
		}
	}

	return nil, fmt.Errorf("current directory is not in any worktree")
}

// FindWorktreeByBranch finds a worktree by branch name
func FindWorktreeByBranch(ctx context.Context, branch string) (*Worktree, error) {
	worktrees, err := List(ctx)
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if wt.Branch == branch {
			return &wt, nil
		}
	}

	return nil, nil // not found
}
