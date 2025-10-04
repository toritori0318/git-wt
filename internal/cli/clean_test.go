package cli

import (
	"strings"
	"testing"
)

func TestNoRemovableWorktreesError(t *testing.T) {
	err := &NoRemovableWorktreesError{}
	errMsg := err.Error()

	if !strings.Contains(errMsg, "no removable worktrees") {
		t.Errorf("NoRemovableWorktreesError should contain 'no removable worktrees', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "main worktree") {
		t.Errorf("NoRemovableWorktreesError should explain about main worktree, got: %s", errMsg)
	}
}

func TestWorktreeRemovalCancelledError(t *testing.T) {
	err := &WorktreeRemovalCancelledError{}
	errMsg := err.Error()

	if !strings.Contains(errMsg, "cancelled") {
		t.Errorf("WorktreeRemovalCancelledError should contain 'cancelled', got: %s", errMsg)
	}
}

func TestConfirm(t *testing.T) {
	// Note: This function reads from os.Stdin, so it's difficult to test without mocking.
	// In a real test environment, you would use dependency injection or interfaces to make this testable.
	// For now, we'll skip this test or use a mock stdin.
	t.Skip("confirm() requires stdin interaction, skipping for now")
}
