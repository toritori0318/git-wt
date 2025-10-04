package cli

import (
	"strings"
	"testing"
)

func TestBranchInUseError(t *testing.T) {
	branch := "feature/test"
	path := "/path/to/worktree"
	err := &BranchInUseError{
		Branch: branch,
		Path:   path,
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, branch) {
		t.Errorf("BranchInUseError should contain branch name '%s', got: %s", branch, errMsg)
	}
	if !strings.Contains(errMsg, path) {
		t.Errorf("BranchInUseError should contain path '%s', got: %s", path, errMsg)
	}
	if !strings.Contains(errMsg, "already in use") {
		t.Errorf("BranchInUseError should explain the issue, got: %s", errMsg)
	}
}

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		wantErr    bool
	}{
		{
			name:       "valid branch name",
			branchName: "feature/test",
			wantErr:    false,
		},
		{
			name:       "valid simple name",
			branchName: "main",
			wantErr:    false,
		},
		{
			name:       "empty string",
			branchName: "",
			wantErr:    true,
		},
		{
			name:       "whitespace only",
			branchName: "   ",
			wantErr:    true,
		},
		{
			name:       "double dots",
			branchName: "feature..test",
			wantErr:    true,
		},
		{
			name:       "starts with hyphen",
			branchName: "-feature",
			wantErr:    true,
		},
		{
			name:       "valid with hyphen",
			branchName: "feature-test",
			wantErr:    false,
		},
		{
			name:       "valid with underscore",
			branchName: "feature_test",
			wantErr:    false,
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
