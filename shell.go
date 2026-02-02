package main

import (
	"os"
	"os/exec"
	"runtime"
)

// RunShellCommand executes a command in the appropriate shell for the OS
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
