package gitx

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "test-repo")

	// Initialize git repo
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create test repo directory: %v", err)
	}

	// Run git init
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Configure git user (required for commits)
	configCmd := exec.Command("git", "config", "user.name", "Test User")
	configCmd.Dir = repoPath
	if err := configCmd.Run(); err != nil {
		t.Fatalf("Failed to config git user: %v", err)
	}

	emailCmd := exec.Command("git", "config", "user.email", "test@example.com")
	emailCmd.Dir = repoPath
	if err := emailCmd.Run(); err != nil {
		t.Fatalf("Failed to config git email: %v", err)
	}

	// Create initial commit
	readmeFile := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(readmeFile, []byte("# Test Repo\n"), 0644); err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	addCmd := exec.Command("git", "add", "README.md")
	addCmd.Dir = repoPath
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to add README: %v", err)
	}

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = repoPath
	if err := commitCmd.Run(); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		// t.TempDir() handles cleanup automatically
	}

	return repoPath, cleanup
}

func TestGetCurrentBranch(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Change to repo directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to change to repo directory: %v", err)
	}

	ctx := context.Background()

	// Test: Get current branch (should be main or master)
	branch, err := GetCurrentBranch(ctx)
	if err != nil {
		t.Fatalf("GetCurrentBranch() error = %v", err)
	}

	// Git init creates either 'main' or 'master' depending on git version
	if branch != "main" && branch != "master" {
		t.Errorf("GetCurrentBranch() = %q, want 'main' or 'master'", branch)
	}
}

func TestBranchExists(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to change to repo directory: %v", err)
	}

	ctx := context.Background()

	// Create a test branch
	createBranchCmd := exec.Command("git", "branch", "test-branch")
	createBranchCmd.Dir = repoPath
	if err := createBranchCmd.Run(); err != nil {
		t.Fatalf("Failed to create test branch: %v", err)
	}

	tests := []struct {
		name       string
		branchName string
		want       bool
	}{
		{
			name:       "existing branch",
			branchName: "test-branch",
			want:       true,
		},
		{
			name:       "non-existing branch",
			branchName: "non-existent",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := BranchExists(ctx, tt.branchName)
			if err != nil {
				t.Fatalf("BranchExists() error = %v", err)
			}

			if exists != tt.want {
				t.Errorf("BranchExists(%q) = %v, want %v", tt.branchName, exists, tt.want)
			}
		})
	}
}

func TestDeleteBranch(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to change to repo directory: %v", err)
	}

	ctx := context.Background()

	// Create a test branch
	createBranchCmd := exec.Command("git", "branch", "to-delete")
	createBranchCmd.Dir = repoPath
	if err := createBranchCmd.Run(); err != nil {
		t.Fatalf("Failed to create test branch: %v", err)
	}

	// Verify branch exists
	exists, err := BranchExists(ctx, "to-delete")
	if err != nil {
		t.Fatalf("BranchExists() error = %v", err)
	}
	if !exists {
		t.Fatal("Branch 'to-delete' should exist before deletion")
	}

	// Delete the branch
	if err := DeleteBranch(ctx, "to-delete", false); err != nil {
		t.Fatalf("DeleteBranch() error = %v", err)
	}

	// Verify branch no longer exists
	exists, err = BranchExists(ctx, "to-delete")
	if err != nil {
		t.Fatalf("BranchExists() after deletion error = %v", err)
	}
	if exists {
		t.Error("Branch 'to-delete' should not exist after deletion")
	}
}

func TestIsBranchMerged(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to change to repo directory: %v", err)
	}

	ctx := context.Background()

	// Get current branch name (main or master)
	currentBranch, err := GetCurrentBranch(ctx)
	if err != nil {
		t.Fatalf("GetCurrentBranch() error = %v", err)
	}

	// Create and merge a branch
	createBranchCmd := exec.Command("git", "branch", "merged-branch")
	createBranchCmd.Dir = repoPath
	if err := createBranchCmd.Run(); err != nil {
		t.Fatalf("Failed to create merged-branch: %v", err)
	}

	// Create an unmerged branch
	createUnmergedCmd := exec.Command("git", "branch", "unmerged-branch")
	createUnmergedCmd.Dir = repoPath
	if err := createUnmergedCmd.Run(); err != nil {
		t.Fatalf("Failed to create unmerged-branch: %v", err)
	}

	// Checkout unmerged branch and make a commit
	checkoutCmd := exec.Command("git", "checkout", "unmerged-branch")
	checkoutCmd.Dir = repoPath
	if err := checkoutCmd.Run(); err != nil {
		t.Fatalf("Failed to checkout unmerged-branch: %v", err)
	}

	// Create a new file in unmerged branch
	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	addTestCmd := exec.Command("git", "add", "test.txt")
	addTestCmd.Dir = repoPath
	if err := addTestCmd.Run(); err != nil {
		t.Fatalf("Failed to add test file: %v", err)
	}

	commitTestCmd := exec.Command("git", "commit", "-m", "Add test file")
	commitTestCmd.Dir = repoPath
	if err := commitTestCmd.Run(); err != nil {
		t.Fatalf("Failed to commit test file: %v", err)
	}

	// Checkout back to current branch
	checkoutBackCmd := exec.Command("git", "checkout", currentBranch)
	checkoutBackCmd.Dir = repoPath
	if err := checkoutBackCmd.Run(); err != nil {
		t.Fatalf("Failed to checkout back to %s: %v", currentBranch, err)
	}

	tests := []struct {
		name       string
		branchName string
		want       bool
	}{
		{
			name:       "merged branch",
			branchName: "merged-branch",
			want:       true,
		},
		{
			name:       "unmerged branch",
			branchName: "unmerged-branch",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged, err := IsBranchMerged(ctx, tt.branchName)
			if err != nil {
				t.Fatalf("IsBranchMerged() error = %v", err)
			}

			if merged != tt.want {
				t.Errorf("IsBranchMerged(%q) = %v, want %v", tt.branchName, merged, tt.want)
			}
		})
	}
}
