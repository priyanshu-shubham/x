package main

import (
	"os"
	"os/exec"
	"runtime"
)

// Common editor names to try on Linux/Unix
var fallbackEditors = []string{"nano", "vim", "vi"}

// OpenEditor opens a file in the system's default text editor
func OpenEditor(filePath string) error {
	switch runtime.GOOS {
	case OSDarwin:
		return openEditorDarwin(filePath)
	case OSWindows:
		return openEditorWindows(filePath)
	default:
		return openEditorUnix(filePath)
	}
}

func openEditorDarwin(filePath string) error {
	cmd := exec.Command("open", "-t", filePath)
	return cmd.Start()
}

func openEditorWindows(filePath string) error {
	cmd := exec.Command("notepad", filePath)
	return cmd.Start()
}

func openEditorUnix(filePath string) error {
	editor := findEditor()
	if editor == "" {
		return ErrNoEditorFound
	}

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func findEditor() string {
	// Check environment variables first
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}

	// Try common editors
	for _, editor := range fallbackEditors {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}

	return ""
}
