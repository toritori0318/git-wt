package cli

import (
	"strings"
	"testing"

	"github.com/toritsuyo/wt/internal/gitx"
)

func TestNoWorktreesError(t *testing.T) {
	err := &NoWorktreesError{}
	errMsg := err.Error()

	if !strings.Contains(errMsg, "no worktrees found") {
		t.Errorf("NoWorktreesError message should contain 'no worktrees found', got: %s", errMsg)
	}
}

func TestIndexOutOfRangeError(t *testing.T) {
	tests := []struct {
		name  string
		index int
		max   int
	}{
		{
			name:  "index 5 max 3",
			index: 5,
			max:   3,
		},
		{
			name:  "index 0 max 10",
			index: 0,
			max:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &IndexOutOfRangeError{
				Index: tt.index,
				Max:   tt.max,
			}
			errMsg := err.Error()

			if !strings.Contains(errMsg, "out of range") {
				t.Errorf("IndexOutOfRangeError should contain 'out of range', got: %s", errMsg)
			}
		})
	}
}

func TestNoMatchError(t *testing.T) {
	query := "feature/test"
	err := &NoMatchError{Query: query}
	errMsg := err.Error()

	if !strings.Contains(errMsg, query) {
		t.Errorf("NoMatchError should contain query '%s', got: %s", query, errMsg)
	}
	if !strings.Contains(errMsg, "no match") {
		t.Errorf("NoMatchError should contain 'no match', got: %s", errMsg)
	}
}

func TestFormatBranch(t *testing.T) {
	tests := []struct {
		name     string
		worktree gitx.Worktree
		want     string
	}{
		{
			name: "normal branch",
			worktree: gitx.Worktree{
				Branch:     "main",
				IsDetached: false,
			},
			want: "main",
		},
		{
			name: "detached HEAD with short SHA",
			worktree: gitx.Worktree{
				HEAD:       "abc123",
				IsDetached: true,
			},
			want: "(detached: abc123)",
		},
		{
			name: "detached HEAD with long SHA",
			worktree: gitx.Worktree{
				HEAD:       "abc123def456789",
				IsDetached: true,
			},
			want: "(detached: abc123d)",
		},
		{
			name: "feature branch",
			worktree: gitx.Worktree{
				Branch:     "feature/new-ui",
				IsDetached: false,
			},
			want: "feature/new-ui",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBranch(tt.worktree)
			if got != tt.want {
				t.Errorf("formatBranch() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateDisplayItems(t *testing.T) {
	tests := []struct {
		name      string
		worktrees []gitx.Worktree
		wantCount int
		checkItem func(t *testing.T, items []string)
	}{
		{
			name: "single worktree",
			worktrees: []gitx.Worktree{
				{
					Branch: "main",
					Path:   "/path/to/repo",
				},
			},
			wantCount: 1,
			checkItem: func(t *testing.T, items []string) {
				if !strings.Contains(items[0], "main") {
					t.Errorf("item should contain branch name 'main', got: %s", items[0])
				}
				if !strings.Contains(items[0], "/path/to/repo") {
					t.Errorf("item should contain path '/path/to/repo', got: %s", items[0])
				}
			},
		},
		{
			name: "multiple worktrees",
			worktrees: []gitx.Worktree{
				{
					Branch: "main",
					Path:   "/path/to/repo",
				},
				{
					Branch: "feature/test",
					Path:   "/path/to/repo-feature-test",
				},
			},
			wantCount: 2,
			checkItem: func(t *testing.T, items []string) {
				if len(items) != 2 {
					t.Errorf("expected 2 items, got %d", len(items))
				}
			},
		},
		{
			name: "detached HEAD worktree",
			worktrees: []gitx.Worktree{
				{
					HEAD:       "abc123def456",
					IsDetached: true,
					Path:       "/path/to/detached",
				},
			},
			wantCount: 1,
			checkItem: func(t *testing.T, items []string) {
				if !strings.Contains(items[0], "detached") {
					t.Errorf("item should contain 'detached', got: %s", items[0])
				}
				if !strings.Contains(items[0], "abc123d") {
					t.Errorf("item should contain truncated SHA, got: %s", items[0])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createDisplayItems(tt.worktrees)
			if len(got) != tt.wantCount {
				t.Errorf("createDisplayItems() returned %d items, want %d", len(got), tt.wantCount)
			}
			if tt.checkItem != nil {
				tt.checkItem(t, got)
			}
		})
	}
}

