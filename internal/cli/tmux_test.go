package cli

import (
	"testing"

	"github.com/toritori0318/git-wt/internal/naming"
)

func TestSessionNameSanitization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "safe session name",
			input:    "my-session",
			expected: "my-session",
		},
		{
			name:     "session name with slashes",
			input:    "feature/auth",
			expected: "feature-auth",
		},
		{
			name:     "session name with special characters",
			input:    "evil; rm -rf /",
			expected: "evil-rm-rf",
		},
		{
			name:     "session name with spaces",
			input:    "my session name",
			expected: "my-session-name",
		},
		{
			name:     "session name with consecutive special chars",
			input:    "session!!!name",
			expected: "session-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := naming.Sanitize(tt.input)
			if result != tt.expected {
				t.Errorf("Sanitize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateLayout(t *testing.T) {
	tests := []struct {
		name    string
		layout  string
		wantErr bool
	}{
		{
			name:    "tiled layout",
			layout:  "tiled",
			wantErr: false,
		},
		{
			name:    "horizontal layout",
			layout:  "horizontal",
			wantErr: false,
		},
		{
			name:    "vertical layout",
			layout:  "vertical",
			wantErr: false,
		},
		{
			name:    "even-horizontal layout",
			layout:  "even-horizontal",
			wantErr: false,
		},
		{
			name:    "even-vertical layout",
			layout:  "even-vertical",
			wantErr: false,
		},
		{
			name:    "main-horizontal layout",
			layout:  "main-horizontal",
			wantErr: false,
		},
		{
			name:    "main-vertical layout",
			layout:  "main-vertical",
			wantErr: false,
		},
		{
			name:    "invalid layout",
			layout:  "invalid",
			wantErr: true,
		},
		{
			name:    "empty layout",
			layout:  "",
			wantErr: false, // Empty is allowed (uses tmux default)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLayout(tt.layout)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLayout(%q) error = %v, wantErr %v", tt.layout, err, tt.wantErr)
			}
		})
	}
}
