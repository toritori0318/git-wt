package selectx

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// IsFzfAvailable checks if fzf is installed
func IsFzfAvailable() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

// SelectWithFzf uses fzf to select from a list of items
func SelectWithFzf(items []string, prompt string) (int, error) {
	if len(items) == 0 {
		return -1, fmt.Errorf("no items to select from")
	}

	// Build fzf command
	cmd := exec.Command("fzf",
		"--height=40%",
		"--reverse",
		"--prompt="+prompt+"> ",
		"--select-1", // Auto-select if only one item
	)

	// Pass items to stdin
	cmd.Stdin = bytes.NewBufferString(strings.Join(items, "\n"))

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run fzf
	err := cmd.Run()
	if err != nil {
		// User cancelled (exit code 130)
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 {
				return -1, fmt.Errorf("selection cancelled")
			}
		}
		return -1, fmt.Errorf("fzf failed: %w: %s", err, stderr.String())
	}

	// Get selected item
	selected := strings.TrimSpace(stdout.String())
	if selected == "" {
		return -1, fmt.Errorf("no selection made")
	}

	// Find index of selected item
	for i, item := range items {
		if item == selected {
			return i, nil
		}
	}

	return -1, fmt.Errorf("selected item not found in list")
}
