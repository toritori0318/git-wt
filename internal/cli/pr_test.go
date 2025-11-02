package cli

import (
	"strings"
	"testing"
)

func TestGhNotFoundError(t *testing.T) {
	err := &GhNotFoundError{}
	errMsg := err.Error()

	if !strings.Contains(errMsg, "GitHub CLI") {
		t.Errorf("GhNotFoundError should mention GitHub CLI, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "gh") {
		t.Errorf("GhNotFoundError should mention 'gh' command, got: %s", errMsg)
	}
}

func TestInvalidPRNumberError(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid string",
			input: "abc",
		},
		{
			name:  "negative number",
			input: "-5",
		},
		{
			name:  "zero",
			input: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &InvalidPRNumberError{Input: tt.input}
			errMsg := err.Error()

			if !strings.Contains(errMsg, tt.input) {
				t.Errorf("InvalidPRNumberError should contain input '%s', got: %s", tt.input, errMsg)
			}
			if !strings.Contains(errMsg, "invalid PR number") {
				t.Errorf("InvalidPRNumberError should explain the issue, got: %s", errMsg)
			}
		})
	}
}

func TestValidatePRNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:    "valid PR number",
			input:   "123",
			want:    123,
			wantErr: false,
		},
		{
			name:    "valid single digit",
			input:   "1",
			want:    1,
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   "abc",
			want:    0,
			wantErr: true,
		},
		{
			name:    "negative number",
			input:   "-5",
			want:    0,
			wantErr: true,
		},
		{
			name:    "zero",
			input:   "0",
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validatePRNumber(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePRNumber(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validatePRNumber(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfirmNavigate(t *testing.T) {
	// Note: This function requires stdin interaction, so we test the interface only
	// Full integration tests would need mock stdin
	t.Run("function signature", func(t *testing.T) {
		// Verify function exists and has correct signature
		var buf strings.Builder
		_, err := confirmNavigate(&buf, "test-branch", "/path/to/worktree")
		// We expect an error because there's no stdin input in tests
		if err == nil {
			t.Log("confirmNavigate completed (possibly with default behavior)")
		}
	})
}

func TestConfirmUseExisting(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		cdMode   bool
		quiet    bool
		wantSkip bool // Should skip prompt and auto-confirm
	}{
		{
			name:     "cd mode skips prompt",
			branch:   "feature/auth",
			cdMode:   true,
			quiet:    false,
			wantSkip: true,
		},
		{
			name:     "quiet mode skips prompt",
			branch:   "feature/login",
			cdMode:   false,
			quiet:    true,
			wantSkip: true,
		},
		{
			name:     "normal mode requires prompt",
			branch:   "bugfix/123",
			cdMode:   false,
			quiet:    false,
			wantSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			confirmed, err := confirmUseExisting(&buf, tt.branch, tt.cdMode, tt.quiet)

			if tt.wantSkip {
				// Should auto-confirm without error
				if err != nil {
					t.Errorf("expected no error with skip=true, got %v", err)
				}
				if !confirmed {
					t.Errorf("expected confirmed=true with skip, got false")
				}
			} else {
				// Would require stdin in normal mode
				// Just verify the function was called
				if err == nil {
					t.Log("confirmUseExisting completed (possibly with default behavior)")
				}
			}
		})
	}
}

func TestValidatePRBranchName(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		wantErr    bool
	}{
		{
			name:       "valid branch name",
			branchName: "feature/auth",
			wantErr:    false,
		},
		{
			name:       "valid branch name with numbers",
			branchName: "feature-123",
			wantErr:    false,
		},
		{
			name:       "branch name starting with dash",
			branchName: "-bad",
			wantErr:    true,
		},
		{
			name:       "branch name with double dots",
			branchName: "feature..auth",
			wantErr:    true,
		},
		{
			name:       "empty branch name",
			branchName: "",
			wantErr:    true,
		},
		{
			name:       "whitespace only branch name",
			branchName: "   ",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBranchName(tt.branchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBranchName(%q) error = %v, wantErr %v", tt.branchName, err, tt.wantErr)
			}
		})
	}
}

