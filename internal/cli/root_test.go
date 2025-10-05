package cli

import (
	"reflect"
	"strings"
	"testing"
)

func TestExitCodeError(t *testing.T) {
	innerErr := &NoWorktreesError{}
	err := &ExitCodeError{
		Code: 42,
		Err:  innerErr,
	}

	// Test Error() method
	if err.Error() != innerErr.Error() {
		t.Errorf("ExitCodeError.Error() = %q, want %q", err.Error(), innerErr.Error())
	}

	// Test Unwrap() method
	if err.Unwrap() != innerErr {
		t.Errorf("ExitCodeError.Unwrap() returned wrong error")
	}

	// Test Code field
	if err.Code != 42 {
		t.Errorf("ExitCodeError.Code = %d, want 42", err.Code)
	}
}

func TestShouldPassthrough(t *testing.T) {
	tests := []struct {
		name    string
		errMsg  string
		want    bool
	}{
		{
			name:    "unknown command should passthrough",
			errMsg:  "unknown command \"list\" for \"wt\"",
			want:    true,
		},
		{
			name:    "unknown flag should passthrough",
			errMsg:  "unknown flag: --porcelain",
			want:    true,
		},
		{
			name:    "unknown shorthand flag should passthrough",
			errMsg:  "unknown shorthand flag: 'v' in -v",
			want:    true,
		},
		{
			name:    "other errors should not passthrough",
			errMsg:  "failed to execute command",
			want:    false,
		},
		{
			name:    "nil error should not passthrough",
			errMsg:  "",
			want:    false,
		},
		{
			name:    "uppercase unknown command",
			errMsg:  "Unknown Command \"test\"",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.errMsg != "" {
				err = &mockError{msg: tt.errMsg}
			}

			got := shouldPassthrough(err)
			if got != tt.want {
				t.Errorf("shouldPassthrough(%v) = %v, want %v", tt.errMsg, got, tt.want)
			}
		})
	}
}

// mockError is a simple error implementation for testing
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestFilterPassthroughArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "remove debug flag",
			args: []string{"list", "--debug"},
			want: []string{"list"},
		},
		{
			name: "remove quiet flag",
			args: []string{"--quiet", "list"},
			want: []string{"list"},
		},
		{
			name: "remove repo flag with value",
			args: []string{"list", "--repo", "/path/to/repo"},
			want: []string{"list"},
		},
		{
			name: "remove repo flag with equals",
			args: []string{"list", "--repo=/path/to/repo"},
			want: []string{"list"},
		},
		{
			name: "keep other flags",
			args: []string{"list", "--porcelain", "-v"},
			want: []string{"list", "--porcelain", "-v"},
		},
		{
			name: "handle double dash",
			args: []string{"list", "--debug", "--", "--repo"},
			want: []string{"list", "--repo"},
		},
		{
			name: "empty args",
			args: []string{},
			want: []string{},
		},
		{
			name: "mixed internal and external flags",
			args: []string{"add", "/path", "--debug", "-b", "branch", "--quiet"},
			want: []string{"add", "/path", "-b", "branch"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterPassthroughArgs(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterPassthroughArgs(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestSetVersionInfo(t *testing.T) {
	// Save original values
	origVersion := versionInfo
	origCommit := commitInfo
	origDate := dateInfo
	origRootVersion := rootCmd.Version

	// Restore after test
	defer func() {
		versionInfo = origVersion
		commitInfo = origCommit
		dateInfo = origDate
		rootCmd.Version = origRootVersion
	}()

	// Test SetVersionInfo
	version := "1.2.3"
	commit := "abc123"
	date := "2024-01-01"

	SetVersionInfo(version, commit, date)

	if versionInfo != version {
		t.Errorf("versionInfo = %q, want %q", versionInfo, version)
	}
	if commitInfo != commit {
		t.Errorf("commitInfo = %q, want %q", commitInfo, commit)
	}
	if dateInfo != date {
		t.Errorf("dateInfo = %q, want %q", dateInfo, date)
	}

	// Check rootCmd.Version contains all info
	if !strings.Contains(rootCmd.Version, version) {
		t.Errorf("rootCmd.Version should contain version %q, got %q", version, rootCmd.Version)
	}
	if !strings.Contains(rootCmd.Version, commit) {
		t.Errorf("rootCmd.Version should contain commit %q, got %q", commit, rootCmd.Version)
	}
	if !strings.Contains(rootCmd.Version, date) {
		t.Errorf("rootCmd.Version should contain date %q, got %q", date, rootCmd.Version)
	}
}
