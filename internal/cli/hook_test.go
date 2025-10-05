package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestValidateShell(t *testing.T) {
	tests := []struct {
		name    string
		shell   string
		wantErr bool
	}{
		{
			name:    "bash is supported",
			shell:   "bash",
			wantErr: false,
		},
		{
			name:    "zsh is supported",
			shell:   "zsh",
			wantErr: false,
		},
		{
			name:    "fish is supported",
			shell:   "fish",
			wantErr: false,
		},
		{
			name:    "uppercase bash is normalized",
			shell:   "BASH",
			wantErr: false,
		},
		{
			name:    "mixed case zsh is normalized",
			shell:   "Zsh",
			wantErr: false,
		},
		{
			name:    "shell with whitespace is normalized",
			shell:   "  bash  ",
			wantErr: false,
		},
		{
			name:    "unsupported shell returns error",
			shell:   "powershell",
			wantErr: true,
		},
		{
			name:    "empty string returns error",
			shell:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateShell(tt.shell)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateShell() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check error type for unsupported shells
			if tt.wantErr && err != nil {
				if _, ok := err.(*UnsupportedShellError); !ok {
					t.Errorf("validateShell() error type = %T, want *UnsupportedShellError", err)
				}
			}
		})
	}
}

func TestGetShellScript(t *testing.T) {
	tests := []struct {
		name    string
		shell   string
		wantErr bool
		wantLen bool // true if we expect non-empty script
	}{
		{
			name:    "bash returns script",
			shell:   "bash",
			wantErr: false,
			wantLen: true,
		},
		{
			name:    "zsh returns script",
			shell:   "zsh",
			wantErr: false,
			wantLen: true,
		},
		{
			name:    "fish returns script",
			shell:   "fish",
			wantErr: false,
			wantLen: true,
		},
		{
			name:    "uppercase is normalized",
			shell:   "BASH",
			wantErr: false,
			wantLen: true,
		},
		{
			name:    "unsupported shell returns error",
			shell:   "cmd",
			wantErr: true,
			wantLen: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := getShellScript(tt.shell)
			if (err != nil) != tt.wantErr {
				t.Errorf("getShellScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantLen && len(script) == 0 {
				t.Errorf("getShellScript() returned empty script for %s", tt.shell)
			}

			if !tt.wantErr && len(script) == 0 {
				t.Errorf("getShellScript() returned empty script for valid shell %s", tt.shell)
			}
		})
	}
}

func TestPrintHookScript(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			name:   "prints script to writer",
			script: "#!/bin/bash\necho 'test'",
		},
		{
			name:   "prints empty script",
			script: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printHookScript(&buf, tt.script)

			got := buf.String()
			if got != tt.script {
				t.Errorf("printHookScript() = %q, want %q", got, tt.script)
			}
		})
	}
}

func TestRunHookWithConfig(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name:    "bash outputs script",
			args:    []string{"bash"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "function wt") && !strings.Contains(output, "wt()") {
					t.Errorf("output doesn't contain shell function definition")
				}
			},
		},
		{
			name:    "zsh outputs script",
			args:    []string{"zsh"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "function wt") && !strings.Contains(output, "wt()") {
					t.Errorf("output doesn't contain shell function definition")
				}
			},
		},
		{
			name:    "fish outputs script",
			args:    []string{"fish"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "function wt") {
					t.Errorf("output doesn't contain fish function definition")
				}
			},
		},
		{
			name:    "unsupported shell returns error",
			args:    []string{"powershell"},
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := newHookCmd()
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := runHookWithConfig(cmd, tt.args, &hookCmdConfig{})
			if (err != nil) != tt.wantErr {
				t.Errorf("runHookWithConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}

func TestUnsupportedShellError(t *testing.T) {
	err := &UnsupportedShellError{
		Shell:           "powershell",
		SupportedShells: []string{"bash", "zsh", "fish"},
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "powershell") {
		t.Errorf("error message should contain shell name, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "bash") {
		t.Errorf("error message should contain supported shells, got: %s", errMsg)
	}
}
