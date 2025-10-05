package editor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// FindEditor finds the best available editor
func FindEditor(preferredEditor string) (string, error) {
	// Search for editor in priority order
	candidates := []string{
		preferredEditor,
		os.Getenv("WT_EDITOR"),
		os.Getenv("VISUAL"),
		os.Getenv("EDITOR"),
		"code",   // VS Code
		"idea",   // IntelliJ IDEA
		"subl",   // Sublime Text
		"vim",    // Vim
		"vi",     // Vi
	}

	// Add platform-specific fallbacks: "open" for macOS, "xdg-open" for Linux
	if runtime.GOOS == "darwin" {
		candidates = append(candidates, "open")
	} else if runtime.GOOS == "linux" {
		candidates = append(candidates, "xdg-open")
	}

	for _, editor := range candidates {
		if editor == "" {
			continue
		}

		// Check if command exists
		if path, err := exec.LookPath(editor); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no editor found. Please set WT_EDITOR, VISUAL, or EDITOR environment variable")
}

// Open opens the specified path with an editor
func Open(path, editor string) error {
	editorPath, err := FindEditor(editor)
	if err != nil {
		return err
	}
	return OpenWithPath(path, editorPath)
}

// OpenWithPath opens the specified path with a resolved editor path
func OpenWithPath(path, editorPath string) error {
	cmd := exec.Command(editorPath, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch editor: %w", err)
	}

	return nil
}
