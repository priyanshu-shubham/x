package main

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"
)

// TemplateValues holds runtime values for template substitution
type TemplateValues struct {
	Time      string
	Date      string
	DateTime  string
	Directory string
	OS        string
	Arch      string
	Shell     string
	User      string
}

// GetTemplateValues returns current runtime values
func GetTemplateValues() TemplateValues {
	now := time.Now()
	cwd, _ := os.Getwd()

	shell := os.Getenv("SHELL")
	if shell == "" {
		if runtime.GOOS == OSWindows {
			shell = DefaultShellWindows
		} else {
			shell = DefaultShellUnix
		}
	}

	username := ""
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	return TemplateValues{
		Time:      now.Format(TimeFormat),
		Date:      now.Format(DateFormat),
		DateTime:  now.Format(DateTimeFormat),
		Directory: cwd,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Shell:     shell,
		User:      username,
	}
}

// ToMap converts template values to a map for substitution
func (tv TemplateValues) ToMap() map[string]string {
	return map[string]string{
		PlaceholderTime:      tv.Time,
		PlaceholderDate:      tv.Date,
		PlaceholderDateTime:  tv.DateTime,
		PlaceholderDirectory: tv.Directory,
		PlaceholderOS:        tv.OS,
		PlaceholderArch:      tv.Arch,
		PlaceholderShell:     tv.Shell,
		PlaceholderUser:      tv.User,
	}
}

// ApplyTemplate substitutes all template placeholders in a string
func ApplyTemplate(text string) string {
	values := GetTemplateValues().ToMap()
	for placeholder, value := range values {
		text = strings.ReplaceAll(text, placeholder, value)
	}
	return text
}

// GetDefaultSystemPrompt returns the default system prompt for shell command generation
func GetDefaultSystemPrompt() string {
	tv := GetTemplateValues()

	shellType := "bash/zsh"
	if runtime.GOOS == OSWindows {
		shellType = "PowerShell/cmd"
	}

	return fmt.Sprintf(`You are a command-line assistant. Given a natural language description of what the user wants to do, generate the appropriate command.

Environment:
- OS: %s
- Architecture: %s
- Current directory: %s
- Shell: %s

Rules:
- Output ONLY the command, nothing else
- No explanations, no markdown, no code blocks
- The command should be safe and correct
- Generate commands appropriate for the user's OS (%s commands)
- Use pipes (|) to chain commands when processing output
- Use && to run sequential commands that depend on each other
- Use ; when commands are independent but should run in sequence
- Prefer simple, readable commands over complex one-liners when possible`,
		tv.OS, tv.Arch, tv.Directory, tv.Shell, shellType)
}
