package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Manager manages tmux sessions
type Manager struct {
	sessionName string
}

// Pane represents a tmux pane with worktree information
type Pane struct {
	WorktreePath string
	BranchName   string
}

// SessionConfig holds configuration for creating a tmux session
type SessionConfig struct {
	SessionName string
	Panes       []Pane
	Layout      string // "tiled", "horizontal", "vertical"
	SyncPanes   bool
	NoAttach    bool
}

// NewManager creates a new tmux manager
func NewManager(sessionName string) *Manager {
	return &Manager{
		sessionName: sessionName,
	}
}

// IsTmuxAvailable checks if tmux is installed
func IsTmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// CreateSession creates a new tmux session with split panes
func (m *Manager) CreateSession(cfg SessionConfig) error {
	if len(cfg.Panes) == 0 {
		return fmt.Errorf("no panes to create session for")
	}

	// Determine shell to use
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	// Create new detached session with shell in first pane
	firstPane := cfg.Panes[0]
	cmd := exec.Command("tmux", "new-session", "-d", "-s", m.sessionName,
		"-c", firstPane.WorktreePath, shell)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	// Verify session was created
	time.Sleep(100 * time.Millisecond)
	if !m.SessionExists() {
		return fmt.Errorf("tmux session was not created")
	}

	// Split window for remaining panes
	for i := 1; i < len(cfg.Panes); i++ {
		pane := cfg.Panes[i]
		cmd := exec.Command("tmux", "split-window", "-t", m.sessionName,
			"-c", pane.WorktreePath, shell)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to split window for pane %d: %w", i, err)
		}
	}

	// Apply layout
	if cfg.Layout != "" {
		cmd := exec.Command("tmux", "select-layout", "-t", m.sessionName, cfg.Layout)
		if err := cmd.Run(); err != nil {
			// Layout failure is not critical
			_ = err
		}
	}

	// Wait for panes to be fully initialized
	time.Sleep(200 * time.Millisecond)

	// Enable synchronize-panes if requested
	if cfg.SyncPanes {
		cmd := exec.Command("tmux", "set-window-option", "-t", m.sessionName, "synchronize-panes", "on")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to enable synchronize-panes: %w", err)
		}
	}

	return nil
}

// AttachSession attaches to the tmux session
func (m *Manager) AttachSession() error {
	cmd := exec.Command("tmux", "attach-session", "-t", m.sessionName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to attach to session: %w", err)
	}

	return nil
}

// KillSession kills the tmux session
func (m *Manager) KillSession() error {
	cmd := exec.Command("tmux", "kill-session", "-t", m.sessionName)
	if err := cmd.Run(); err != nil {
		// Session might not exist, which is fine
		return nil
	}
	return nil
}

// SessionExists checks if the tmux session exists
func (m *Manager) SessionExists() bool {
	cmd := exec.Command("tmux", "has-session", "-t", m.sessionName)
	return cmd.Run() == nil
}

// SendKeys sends keys to all panes in the session
func (m *Manager) SendKeys(keys string) error {
	// Get list of panes
	cmd := exec.Command("tmux", "list-panes", "-t", m.sessionName, "-F", "#{pane_id}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list panes: %w (output: %s)", err, string(output))
	}

	panes := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(panes) == 0 || (len(panes) == 1 && panes[0] == "") {
		return fmt.Errorf("no panes found in session %s", m.sessionName)
	}

	// Send keys to each pane
	for _, pane := range panes {
		if pane == "" {
			continue
		}
		cmd := exec.Command("tmux", "send-keys", "-t", pane, keys, "C-m")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to send keys to pane %s: %w", pane, err)
		}
	}

	return nil
}
