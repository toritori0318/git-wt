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

func TestConfirmWithIO(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "y answers yes", input: "y\n", want: true},
		{name: "Y uppercase answers yes", input: "Y\n", want: true},
		{name: "yes answers yes", input: "yes\n", want: true},
		{name: "n answers no", input: "n\n", want: false},
		{name: "empty line answers no", input: "\n", want: false},
		{name: "no input (EOF) answers no", input: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			got := confirmWithIO(strings.NewReader(tt.input), &out, "Proceed?")
			if got != tt.want {
				t.Errorf("confirmWithIO(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfirmWithIOWritesPromptToGivenWriter(t *testing.T) {
	var out strings.Builder
	confirmWithIO(strings.NewReader("y\n"), &out, "Proceed?")

	if !strings.Contains(out.String(), "Proceed? (y/N): ") {
		t.Errorf("expected prompt to be written to writer, got: %q", out.String())
	}
}
