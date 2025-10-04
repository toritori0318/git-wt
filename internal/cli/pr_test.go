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

func TestDetermineLocalBranch(t *testing.T) {
	tests := []struct {
		name       string
		userBranch string
		prNumber   int
		want       string
	}{
		{
			name:       "user specified branch",
			userBranch: "review/pr-123",
			prNumber:   123,
			want:       "review/pr-123",
		},
		{
			name:       "default branch name",
			userBranch: "",
			prNumber:   456,
			want:       "gwt/pr-456",
		},
		{
			name:       "user empty string uses default",
			userBranch: "",
			prNumber:   1,
			want:       "gwt/pr-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineLocalBranch(tt.userBranch, tt.prNumber)
			if got != tt.want {
				t.Errorf("determineLocalBranch(%q, %d) = %q, want %q", tt.userBranch, tt.prNumber, got, tt.want)
			}
		})
	}
}
