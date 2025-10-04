package gitx

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	// Debug controls whether to log git commands to stderr
	Debug = false
)

// RunGit executes a git command with the given arguments
func RunGit(ctx context.Context, args ...string) (string, error) {
	return RunGitInDir(ctx, "", args...)
}

// RunGitInDir executes a git command in a specific directory
func RunGitInDir(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}

	if Debug {
		cmdStr := "git " + strings.Join(args, " ")
		if dir != "" {
			cmdStr = fmt.Sprintf("(cd %s && %s)", dir, cmdStr)
		}
		fmt.Fprintf(os.Stderr, "[debug] %s\n", cmdStr)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return "", fmt.Errorf("git %s failed: %w: %s", args[0], err, stderrStr)
		}
		return "", fmt.Errorf("git %s failed: %w", args[0], err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// CheckGitInstalled verifies that git is available
func CheckGitInstalled() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git command not found: please install git")
	}
	return nil
}
