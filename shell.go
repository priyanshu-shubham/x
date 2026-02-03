package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

// ANSI escape codes
const (
	ansiDim   = "\033[2m"
	ansiReset = "\033[0m"
)

// RunShellCommand executes a command in the appropriate shell for the OS
// with terminal connected (interactive mode)
func RunShellCommand(command string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == OSWindows {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// RunShellCommandWithOutput executes a command and returns the combined output
// Output is captured silently without streaming to terminal
func RunShellCommandWithOutput(command string) (string, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == OSWindows {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	return strings.TrimSpace(output.String()), err
}

// dimWriter wraps a writer to output dimmed text
type dimWriter struct {
	w       io.Writer
	started bool
}

func (d *dimWriter) Write(p []byte) (n int, err error) {
	if !d.started {
		d.w.Write([]byte(ansiDim))
		d.started = true
	}
	return d.w.Write(p)
}

// RunShellCommandStreaming executes a command, streams dimmed output to terminal,
// and returns the captured output. Handles Ctrl+C gracefully.
func RunShellCommandStreaming(command string) (string, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == OSWindows {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	var output bytes.Buffer
	dimOut := &dimWriter{w: os.Stdout}
	dimErr := &dimWriter{w: os.Stderr}

	// Write to both terminal (dimmed) and buffer simultaneously
	cmd.Stdout = io.MultiWriter(dimOut, &output)
	cmd.Stderr = io.MultiWriter(dimErr, &output)

	// Handle Ctrl+C gracefully to ensure we reset terminal formatting
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Reset terminal formatting when done (or interrupted)
	defer func() {
		if dimOut.started || dimErr.started {
			fmt.Print(ansiReset)
		}
	}()

	// Run command in goroutine so we can handle signals
	errChan := make(chan error, 1)
	go func() {
		errChan <- cmd.Run()
	}()

	// Wait for either completion or interrupt
	select {
	case err := <-errChan:
		return strings.TrimSpace(output.String()), err
	case <-sigChan:
		// Kill the process on interrupt
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return strings.TrimSpace(output.String()), fmt.Errorf("interrupted")
	}
}
