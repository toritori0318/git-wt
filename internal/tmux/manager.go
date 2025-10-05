package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CommandExecutor defines the interface for executing commands
type CommandExecutor interface {
	Run(name string, args ...string) error
	Output(name string, args ...string) ([]byte, error)
}

// defaultExecutor implements CommandExecutor using exec.Command
type defaultExecutor struct{}

func (e *defaultExecutor) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

func (e *defaultExecutor) Output(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

// Manager manages tmux sessions
type Manager struct {
	sessionName string
	executor    CommandExecutor
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
	Debug       bool // Enable debug logging
}

// NewManager creates a new tmux manager with default executor
func NewManager(sessionName string) *Manager {
	return NewManagerWithExecutor(sessionName, &defaultExecutor{})
}

// NewManagerWithExecutor creates a new tmux manager with custom executor
func NewManagerWithExecutor(sessionName string, executor CommandExecutor) *Manager {
	return &Manager{
		sessionName: sessionName,
		executor:    executor,
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
	if err := m.executor.Run("tmux", "new-session", "-d", "-s", m.sessionName,
		"-c", firstPane.WorktreePath, shell); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	// Verify session was created with retry
	maxRetries := 10
	retryDelay := 50 * time.Millisecond
	sessionCreated := false
	for i := 0; i < maxRetries; i++ {
		if m.SessionExists() {
			sessionCreated = true
			break
		}
		time.Sleep(retryDelay)
	}
	if !sessionCreated {
		return fmt.Errorf("tmux session was not created after %d retries", maxRetries)
	}

	// Split window for remaining panes
	for i := 1; i < len(cfg.Panes); i++ {
		pane := cfg.Panes[i]
		if err := m.executor.Run("tmux", "split-window", "-t", m.sessionName,
			"-c", pane.WorktreePath, shell); err != nil {
			return fmt.Errorf("failed to split window for pane %d: %w", i, err)
		}
	}

	// Apply layout
	if cfg.Layout != "" {
		if err := m.executor.Run("tmux", "select-layout", "-t", m.sessionName, cfg.Layout); err != nil {
			// Layout failure is not critical, but log in debug mode
			if cfg.Debug {
				fmt.Fprintf(os.Stderr, "Warning: failed to set layout '%s': %v\n", cfg.Layout, err)
			}
		}
	}

	// Enable synchronize-panes if requested
	if cfg.SyncPanes {
		if err := m.executor.Run("tmux", "set-window-option", "-t", m.sessionName, "synchronize-panes", "on"); err != nil {
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
	if err := m.executor.Run("tmux", "kill-session", "-t", m.sessionName); err != nil {
		// Session might not exist, which is fine
		return nil
	}
	return nil
}

// SessionExists checks if the tmux session exists
func (m *Manager) SessionExists() bool {
	return m.executor.Run("tmux", "has-session", "-t", m.sessionName) == nil
}

// SendKeys sends keys to all panes in the session
func (m *Manager) SendKeys(keys string) error {
	// Get list of panes
	output, err := m.executor.Output("tmux", "list-panes", "-t", m.sessionName, "-F", "#{pane_id}")
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
		if err := m.executor.Run("tmux", "send-keys", "-t", pane, keys, "C-m"); err != nil {
			return fmt.Errorf("failed to send keys to pane %s: %w", pane, err)
		}
	}

	return nil
}
