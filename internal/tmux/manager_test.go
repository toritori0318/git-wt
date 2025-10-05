package tmux

import (
	"errors"
	"strings"
	"testing"
)

// mockExecutor is a mock implementation of CommandExecutor for testing
type mockExecutor struct {
	runCalls    [][]string
	outputCalls [][]string
	runErr      error
	outputData  []byte
	outputErr   error
	// runFunc allows dynamic behavior for testing
	runFunc func(name string, args ...string) error
}

func (m *mockExecutor) Run(name string, args ...string) error {
	m.runCalls = append(m.runCalls, append([]string{name}, args...))
	if m.runFunc != nil {
		return m.runFunc(name, args...)
	}
	return m.runErr
}

func (m *mockExecutor) Output(name string, args ...string) ([]byte, error) {
	m.outputCalls = append(m.outputCalls, append([]string{name}, args...))
	return m.outputData, m.outputErr
}

func TestNewManager(t *testing.T) {
	m := NewManager("test-session")
	if m.sessionName != "test-session" {
		t.Errorf("expected session name 'test-session', got %q", m.sessionName)
	}
	if m.executor == nil {
		t.Error("expected executor to be initialized")
	}
}

func TestNewManagerWithExecutor(t *testing.T) {
	mockExec := &mockExecutor{}
	m := NewManagerWithExecutor("test-session", mockExec)

	if m.sessionName != "test-session" {
		t.Errorf("expected session name 'test-session', got %q", m.sessionName)
	}
	if m.executor != mockExec {
		t.Error("expected custom executor to be set")
	}
}

func TestSessionExists(t *testing.T) {
	tests := []struct {
		name     string
		runErr   error
		expected bool
	}{
		{
			name:     "session exists",
			runErr:   nil,
			expected: true,
		},
		{
			name:     "session does not exist",
			runErr:   errors.New("session not found"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &mockExecutor{runErr: tt.runErr}
			m := NewManagerWithExecutor("test-session", mockExec)

			result := m.SessionExists()
			if result != tt.expected {
				t.Errorf("SessionExists() = %v, want %v", result, tt.expected)
			}

			if len(mockExec.runCalls) != 1 {
				t.Fatalf("expected 1 run call, got %d", len(mockExec.runCalls))
			}

			expectedCmd := []string{"tmux", "has-session", "-t", "test-session"}
			if !equalSlices(mockExec.runCalls[0], expectedCmd) {
				t.Errorf("expected command %v, got %v", expectedCmd, mockExec.runCalls[0])
			}
		})
	}
}

func TestKillSession(t *testing.T) {
	tests := []struct {
		name   string
		runErr error
	}{
		{
			name:   "kill successful",
			runErr: nil,
		},
		{
			name:   "kill fails (session doesn't exist)",
			runErr: errors.New("session not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &mockExecutor{runErr: tt.runErr}
			m := NewManagerWithExecutor("test-session", mockExec)

			err := m.KillSession()
			// KillSession should always return nil (ignores errors)
			if err != nil {
				t.Errorf("KillSession() should return nil, got %v", err)
			}

			if len(mockExec.runCalls) != 1 {
				t.Fatalf("expected 1 run call, got %d", len(mockExec.runCalls))
			}

			expectedCmd := []string{"tmux", "kill-session", "-t", "test-session"}
			if !equalSlices(mockExec.runCalls[0], expectedCmd) {
				t.Errorf("expected command %v, got %v", expectedCmd, mockExec.runCalls[0])
			}
		})
	}
}

func TestSendKeys(t *testing.T) {
	tests := []struct {
		name        string
		outputData  []byte
		outputErr   error
		runErr      error
		expectError bool
	}{
		{
			name:        "send keys to single pane",
			outputData:  []byte("%0"),
			outputErr:   nil,
			runErr:      nil,
			expectError: false,
		},
		{
			name:        "send keys to multiple panes",
			outputData:  []byte("%0\n%1\n%2"),
			outputErr:   nil,
			runErr:      nil,
			expectError: false,
		},
		{
			name:        "list panes fails",
			outputData:  nil,
			outputErr:   errors.New("tmux not running"),
			runErr:      nil,
			expectError: true,
		},
		{
			name:        "no panes found",
			outputData:  []byte(""),
			outputErr:   nil,
			runErr:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &mockExecutor{
				outputData: tt.outputData,
				outputErr:  tt.outputErr,
				runErr:     tt.runErr,
			}
			m := NewManagerWithExecutor("test-session", mockExec)

			err := m.SendKeys("echo hello")
			if (err != nil) != tt.expectError {
				t.Errorf("SendKeys() error = %v, expectError %v", err, tt.expectError)
			}

			if !tt.expectError && tt.outputErr == nil {
				// Verify list-panes command was called
				if len(mockExec.outputCalls) != 1 {
					t.Fatalf("expected 1 output call, got %d", len(mockExec.outputCalls))
				}

				expectedListCmd := []string{"tmux", "list-panes", "-t", "test-session", "-F", "#{pane_id}"}
				if !equalSlices(mockExec.outputCalls[0], expectedListCmd) {
					t.Errorf("expected command %v, got %v", expectedListCmd, mockExec.outputCalls[0])
				}

				// Verify send-keys was called for each pane
				panes := strings.Split(strings.TrimSpace(string(tt.outputData)), "\n")
				expectedRunCalls := len(panes)
				if len(mockExec.runCalls) != expectedRunCalls {
					t.Errorf("expected %d run calls, got %d", expectedRunCalls, len(mockExec.runCalls))
				}
			}
		})
	}
}

func TestCreateSession_NoPanes(t *testing.T) {
	mockExec := &mockExecutor{}
	m := NewManagerWithExecutor("test-session", mockExec)

	cfg := SessionConfig{
		Panes: []Pane{},
	}

	err := m.CreateSession(cfg)
	if err == nil {
		t.Error("expected error when creating session with no panes")
	}
	if !strings.Contains(err.Error(), "no panes") {
		t.Errorf("expected error message to contain 'no panes', got %q", err.Error())
	}
}

func TestCreateSession_SessionCreationFails(t *testing.T) {
	mockExec := &mockExecutor{
		runErr: errors.New("tmux command failed"),
	}
	m := NewManagerWithExecutor("test-session", mockExec)

	cfg := SessionConfig{
		Panes: []Pane{
			{WorktreePath: "/tmp/test", BranchName: "main"},
		},
	}

	err := m.CreateSession(cfg)
	if err == nil {
		t.Error("expected error when session creation fails")
	}
	if !strings.Contains(err.Error(), "failed to create tmux session") {
		t.Errorf("expected error message to contain 'failed to create tmux session', got %q", err.Error())
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestDefaultExecutor(t *testing.T) {
	// Test that defaultExecutor implements CommandExecutor interface
	var _ CommandExecutor = &defaultExecutor{}

	// Basic smoke test for Run (will fail if tmux is not installed)
	exec := &defaultExecutor{}
	err := exec.Run("echo", "test")
	if err != nil {
		t.Logf("echo command failed: %v (this is OK if in restricted environment)", err)
	}

	// Basic smoke test for Output
	output, err := exec.Output("echo", "test")
	if err != nil {
		t.Logf("echo command failed: %v (this is OK if in restricted environment)", err)
	} else {
		expected := "test\n"
		if string(output) != expected {
			t.Errorf("expected output %q, got %q", expected, string(output))
		}
	}
}

func TestIsTmuxAvailable(t *testing.T) {
	// This test depends on the environment
	// We can only verify that it doesn't panic
	available := IsTmuxAvailable()
	t.Logf("tmux available: %v", available)
}

// TestCreateSession_WithRetry tests that session creation waits for session to exist
func TestCreateSession_WithRetry(t *testing.T) {
	callCount := 0
	mockExec := &mockExecutor{
		runFunc: func(name string, args ...string) error {
			// Simulate that has-session fails first 2 times, then succeeds
			if name == "tmux" && len(args) > 0 && args[0] == "has-session" {
				callCount++
				if callCount <= 2 {
					return errors.New("session not found yet")
				}
				return nil // Session exists on 3rd check
			}
			// All other commands succeed
			return nil
		},
	}

	m := NewManagerWithExecutor("test-session", mockExec)

	cfg := SessionConfig{
		SessionName: "test-session",
		Panes: []Pane{
			{WorktreePath: "/tmp/wt1", BranchName: "main"},
		},
		Layout:    "tiled",
		SyncPanes: false,
		NoAttach:  false,
	}

	err := m.CreateSession(cfg)
	if err != nil {
		t.Errorf("CreateSession should succeed with retry, got error: %v", err)
	}

	// Verify has-session was called multiple times
	hasSessionCount := 0
	for _, call := range mockExec.runCalls {
		if len(call) > 1 && call[1] == "has-session" {
			hasSessionCount++
		}
	}

	if hasSessionCount < 2 {
		t.Errorf("expected at least 2 has-session calls for retry, got %d", hasSessionCount)
	}
}

// TestCreateSession_MockE2E tests the complete flow with mocked executor
func TestCreateSession_MockE2E(t *testing.T) {
	// Create a mock that simulates successful tmux operations
	mockExec := &mockExecutor{
		runErr: nil, // All commands succeed
	}

	m := NewManagerWithExecutor("test-session", mockExec)

	cfg := SessionConfig{
		SessionName: "test-session",
		Panes: []Pane{
			{WorktreePath: "/tmp/wt1", BranchName: "main"},
			{WorktreePath: "/tmp/wt2", BranchName: "feature"},
		},
		Layout:    "tiled",
		SyncPanes: true,
		NoAttach:  false,
	}

	err := m.CreateSession(cfg)
	if err != nil {
		t.Errorf("CreateSession should succeed, got error: %v", err)
	}

	// Verify at least the new-session command was called
	if len(mockExec.runCalls) < 1 {
		t.Error("expected at least one run call for new-session")
	} else {
		firstCmd := mockExec.runCalls[0]
		if firstCmd[0] != "tmux" || firstCmd[1] != "new-session" {
			t.Errorf("expected first command to be 'tmux new-session', got %v", firstCmd)
		}
	}
}
