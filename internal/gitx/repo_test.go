package gitx

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGetMainWorktreeRoot(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("from main repository", func(t *testing.T) {
		// Get main worktree root from main repository
		root, err := getMainWorktreeRoot(ctx, repoPath)
		if err != nil {
			t.Fatalf("getMainWorktreeRoot() error = %v", err)
		}

		// Resolve symlinks for comparison (macOS /var -> /private/var)
		rootResolved, err := filepath.EvalSymlinks(root)
		if err != nil {
			rootResolved = root
		}
		repoPathResolved, err := filepath.EvalSymlinks(repoPath)
		if err != nil {
			repoPathResolved = repoPath
		}

		if rootResolved != repoPathResolved {
			t.Errorf("getMainWorktreeRoot() = %q, want %q", rootResolved, repoPathResolved)
		}
	})

	t.Run("from worktree directory", func(t *testing.T) {
		// Create a worktree
		worktreePath := filepath.Join(t.TempDir(), "test-worktree")

		// Create a new branch and worktree
		createBranchCmd := exec.Command("git", "branch", "test-wt")
		createBranchCmd.Dir = repoPath
		if err := createBranchCmd.Run(); err != nil {
			t.Fatalf("Failed to create branch: %v", err)
		}

		addWorktreeCmd := exec.Command("git", "worktree", "add", worktreePath, "test-wt")
		addWorktreeCmd.Dir = repoPath
		if err := addWorktreeCmd.Run(); err != nil {
			t.Fatalf("Failed to add worktree: %v", err)
		}

		// Get main worktree root from worktree directory
		root, err := getMainWorktreeRoot(ctx, worktreePath)
		if err != nil {
			t.Fatalf("getMainWorktreeRoot() error = %v", err)
		}

		// Resolve symlinks for comparison (macOS /var -> /private/var)
		rootResolved, err := filepath.EvalSymlinks(root)
		if err != nil {
			rootResolved = root
		}
		repoPathResolved, err := filepath.EvalSymlinks(repoPath)
		if err != nil {
			repoPathResolved = repoPath
		}

		// Should return the main repository path, not the worktree path
		if rootResolved != repoPathResolved {
			t.Errorf("getMainWorktreeRoot() from worktree = %q, want %q", rootResolved, repoPathResolved)
		}
	})
}

func TestGetRepo(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	repoName := filepath.Base(repoPath)
	repoParent := filepath.Dir(repoPath)

	t.Run("from main repository", func(t *testing.T) {
		repo, err := GetRepo(ctx, repoPath)
		if err != nil {
			t.Fatalf("GetRepo() error = %v", err)
		}

		// Resolve symlinks for comparison (macOS /var -> /private/var)
		rootResolved, _ := filepath.EvalSymlinks(repo.Root)
		repoPathResolved, _ := filepath.EvalSymlinks(repoPath)
		parentResolved, _ := filepath.EvalSymlinks(repo.Parent)
		repoParentResolved, _ := filepath.EvalSymlinks(repoParent)

		if rootResolved != repoPathResolved {
			t.Errorf("GetRepo().Root = %q, want %q", rootResolved, repoPathResolved)
		}
		if repo.Name != repoName {
			t.Errorf("GetRepo().Name = %q, want %q", repo.Name, repoName)
		}
		if parentResolved != repoParentResolved {
			t.Errorf("GetRepo().Parent = %q, want %q", parentResolved, repoParentResolved)
		}
	})

	t.Run("from worktree directory", func(t *testing.T) {
		// Create a worktree
		worktreePath := filepath.Join(t.TempDir(), "test-worktree-2")

		// Create a new branch and worktree
		createBranchCmd := exec.Command("git", "branch", "test-wt-2")
		createBranchCmd.Dir = repoPath
		if err := createBranchCmd.Run(); err != nil {
			t.Fatalf("Failed to create branch: %v", err)
		}

		addWorktreeCmd := exec.Command("git", "worktree", "add", worktreePath, "test-wt-2")
		addWorktreeCmd.Dir = repoPath
		if err := addWorktreeCmd.Run(); err != nil {
			t.Fatalf("Failed to add worktree: %v", err)
		}

		// Get repo info from worktree directory
		repo, err := GetRepo(ctx, worktreePath)
		if err != nil {
			t.Fatalf("GetRepo() error = %v", err)
		}

		// Resolve symlinks for comparison (macOS /var -> /private/var)
		rootResolved, _ := filepath.EvalSymlinks(repo.Root)
		repoPathResolved, _ := filepath.EvalSymlinks(repoPath)
		parentResolved, _ := filepath.EvalSymlinks(repo.Parent)
		repoParentResolved, _ := filepath.EvalSymlinks(repoParent)

		// Should return main repository info, not worktree info
		if rootResolved != repoPathResolved {
			t.Errorf("GetRepo().Root from worktree = %q, want %q (main repo)", rootResolved, repoPathResolved)
		}
		if repo.Name != repoName {
			t.Errorf("GetRepo().Name from worktree = %q, want %q", repo.Name, repoName)
		}
		if parentResolved != repoParentResolved {
			t.Errorf("GetRepo().Parent from worktree = %q, want %q", parentResolved, repoParentResolved)
		}
	})
}
