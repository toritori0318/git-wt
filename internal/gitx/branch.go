package gitx

import (
	"context"
	"fmt"
	"strings"
)

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(ctx context.Context) (string, error) {
	branch, err := RunGit(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(branch), nil
}

// BranchExists checks if a branch exists locally
func BranchExists(ctx context.Context, branch string) (bool, error) {
	ref := fmt.Sprintf("refs/heads/%s", branch)
	_, err := RunGit(ctx, "show-ref", "--verify", ref)
	if err != nil {
		// show-ref returns an error when the branch doesn't exist
		if strings.Contains(err.Error(), "not a valid ref") ||
		   strings.Contains(err.Error(), "fatal:") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// DeleteBranch deletes a local branch
func DeleteBranch(ctx context.Context, branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}

	_, err := RunGit(ctx, "branch", flag, branch)
	return err
}

// IsBranchMerged checks if a branch is merged into the current branch
func IsBranchMerged(ctx context.Context, branch string) (bool, error) {
	output, err := RunGit(ctx, "branch", "--merged")
	if err != nil {
		return false, err
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Branch name format: "  branch" or "* branch"
		branchName := strings.TrimSpace(strings.TrimPrefix(line, "*"))
		if branchName == branch {
			return true, nil
		}
	}

	return false, nil
}

// IsUsingBranch checks if any worktree (except the specified path) is using the branch
func IsUsingBranch(ctx context.Context, branch string, excludePath string) (bool, error) {
	worktrees, err := List(ctx)
	if err != nil {
		return false, err
	}

	for _, wt := range worktrees {
		if wt.Branch == branch && wt.Path != excludePath {
			return true, nil
		}
	}

	return false, nil
}
