package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

//go:embed commands.template.yaml
var commandsTemplate string

// Step types for pipeline execution

// ExecStep executes a shell command
type ExecStep struct {
	Command string `yaml:"command"`         // Default command (used if no OS-specific command matches)
	Windows string `yaml:"windows"`         // Windows-specific command
	Darwin  string `yaml:"darwin"`          // macOS-specific command
	Linux   string `yaml:"linux"`           // Linux-specific command
	Confirm bool   `yaml:"confirm"`         // Prompt before execution
	Silent  bool   `yaml:"silent"`          // Don't print output (for intermediate steps)
	Summary string `yaml:"summary"`         // Optional: description shown before confirm
	Risk    string `yaml:"risk"`            // Optional: risk level shown before confirm (none/low/medium/high)
	Safer   string `yaml:"safer"`           // Optional: safer alternative shown for risky commands
}

// LLMStep makes a single LLM call
type LLMStep struct {
	System string `yaml:"system"`
	Prompt string `yaml:"prompt"`
	Silent bool   `yaml:"silent"` // Don't print output (for intermediate steps)
}

// AgenticStep runs a multi-turn agentic loop
type AgenticStep struct {
	System        string `yaml:"system"`
	Prompt        string `yaml:"prompt"`
	MaxIterations int    `yaml:"max_iterations"`
	AutoExecute   bool   `yaml:"auto_execute"` // Auto-execute shell commands without confirmation
}

// SubcommandStep calls another command
type SubcommandStep struct {
	Name   string   `yaml:"name"`   // Name of the command to call
	Args   []string `yaml:"args"`   // Arguments to pass (supports variable interpolation)
	Silent bool     `yaml:"silent"` // Don't print output
}

// Step represents a single step in a pipeline
type Step struct {
	ID         string          `yaml:"id,omitempty"` // Optional step identifier for referencing output
	Exec       *ExecStep       `yaml:"exec,omitempty"`
	LLM        *LLMStep        `yaml:"llm,omitempty"`
	Agentic    *AgenticStep    `yaml:"agentic,omitempty"`
	Subcommand *SubcommandStep `yaml:"subcommand,omitempty"`
}

// Arg represents a named argument for a command
type Arg struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Rest        bool   `yaml:"rest"` // Capture remaining args as one string
}

// Command represents a custom command configuration
type Command struct {
	Description string `yaml:"description"`
	Args        []Arg  `yaml:"args"`
	Steps       []Step `yaml:"steps"`
	Source      string `yaml:"-"` // Where this command was loaded from (not in YAML)
}

// CommandsConfig holds all commands and the default
type CommandsConfig struct {
	Default  string
	Commands map[string]Command
}

// getCommandsPath returns the global commands config file path
func getCommandsPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, CommandsFileName), nil
}

// findAllLocalCommandsFiles searches for xcommands.yaml files from root to cwd
// Returns paths ordered from root to current directory (so later ones override earlier)
func findAllLocalCommandsFiles() []string {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	// Collect all directories from cwd to root
	var dirs []string
	dir := cwd
	for {
		dirs = append(dirs, dir)
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Reverse to get root-to-cwd order
	for i, j := 0, len(dirs)-1; i < j; i, j = i+1, j-1 {
		dirs[i], dirs[j] = dirs[j], dirs[i]
	}

	// Find xcommands.yaml files
	var paths []string
	for _, d := range dirs {
		path := filepath.Join(d, LocalCommandsFileName)
		if _, err := os.Stat(path); err == nil {
			paths = append(paths, path)
		}
	}

	return paths
}

// LoadCommandsConfig reads and merges command configurations.
// Configs are merged with this precedence (later overrides earlier):
// 1. Built-in commands (shell, new) - always up-to-date
// 2. Global commands.yaml
// 3. Parent xcommands.yaml files (from root to current directory)
// 4. Local xcommands.yaml (in current directory)
func LoadCommandsConfig() (*CommandsConfig, error) {
	config := &CommandsConfig{
		Default:  "shell",
		Commands: make(map[string]Command),
	}

	// Load built-in commands first (always current, can be overridden)
	builtins, err := getBuiltinCommands()
	if err != nil {
		return nil, fmt.Errorf("failed to load built-in commands: %w", err)
	}
	for name, cmd := range builtins {
		config.Commands[name] = cmd
	}

	// Load global config (can override built-ins)
	globalPath, err := getCommandsPath()
	if err != nil {
		return nil, err
	}

	globalData, err := os.ReadFile(globalPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the default commands file automatically
			if err := os.MkdirAll(filepath.Dir(globalPath), DirPerms); err != nil {
				return nil, err
			}
			if err := os.WriteFile(globalPath, []byte(commandsTemplate), CommandsPerms); err != nil {
				return nil, err
			}
			globalData, err = os.ReadFile(globalPath)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Parse and merge global config
	if err := mergeConfig(config, globalData, "global"); err != nil {
		return nil, fmt.Errorf("failed to parse global config: %w", err)
	}

	// Load and merge all local xcommands.yaml files (root to cwd order)
	localPaths := findAllLocalCommandsFiles()
	for _, path := range localPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}
		// Use relative path or just filename for display
		source := filepath.Base(filepath.Dir(path))
		if source == "." || source == "" {
			source = "local"
		}
		if err := mergeConfig(config, data, source); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", path, err)
		}
	}

	return config, nil
}

// mergeConfig parses YAML data and merges it into the existing config
func mergeConfig(config *CommandsConfig, data []byte, source string) error {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Extract default if present
	if defaultVal, ok := raw["default"]; ok {
		if defaultStr, ok := defaultVal.(string); ok {
			config.Default = defaultStr
		}
		delete(raw, "default")
	}

	// Parse and merge commands
	for name, value := range raw {
		valueData, err := yaml.Marshal(value)
		if err != nil {
			continue
		}

		var cmd Command
		if err := yaml.Unmarshal(valueData, &cmd); err != nil {
			continue
		}

		cmd.Source = source
		config.Commands[name] = cmd
	}

	return nil
}

// LoadCommands reads and parses the commands configuration (legacy compatibility)
func LoadCommands() (map[string]Command, error) {
	config, err := LoadCommandsConfig()
	if err != nil {
		return nil, err
	}
	return config.Commands, nil
}

// EnsureCommandsFile creates the commands file if it doesn't exist
func EnsureCommandsFile() (string, error) {
	path, err := getCommandsPath()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), DirPerms); err != nil {
		return "", err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte(commandsTemplate), CommandsPerms); err != nil {
			return "", err
		}
	}

	return path, nil
}

// IsReservedCommand checks if a command name is reserved
func IsReservedCommand(name string) bool {
	switch name {
	case CmdConfigure, CmdCommands, CmdUsage, CmdUpgrade, CmdVersion:
		return true
	default:
		return false
	}
}

// PrintHelp prints the main help message with all available commands
func PrintHelp(config *CommandsConfig) {
	fmt.Println("Usage: x <command> [args]")
	fmt.Println()
	fmt.Println("Built-in commands:")
	fmt.Println("  configure   Configure authentication")
	fmt.Println("  commands    Edit custom commands")
	fmt.Println("  usage       Show token usage and cost")
	fmt.Println("  upgrade     Upgrade to latest version")
	fmt.Println("  version     Show current version")
	fmt.Println()

	if len(config.Commands) > 0 {
		// Group commands by source
		bySource := make(map[string][]string)
		for name, cmd := range config.Commands {
			bySource[cmd.Source] = append(bySource[cmd.Source], name)
		}

		// Sort sources: built-in first, then global, then others alphabetically
		var sources []string
		for source := range bySource {
			sources = append(sources, source)
		}
		sort.Slice(sources, func(i, j int) bool {
			// built-in comes first
			if sources[i] == "built-in" {
				return true
			}
			if sources[j] == "built-in" {
				return false
			}
			// global comes second
			if sources[i] == "global" {
				return true
			}
			if sources[j] == "global" {
				return false
			}
			return sources[i] < sources[j]
		})

		// Find max name length across all commands for alignment
		maxLen := 0
		for name := range config.Commands {
			if len(name) > maxLen {
				maxLen = len(name)
			}
		}

		for _, source := range sources {
			names := bySource[source]
			sort.Strings(names)

			// Print source header
			fmt.Printf("Commands (%s):\n", source)

			for _, name := range names {
				cmd := config.Commands[name]
				desc := cmd.Description
				if desc == "" {
					desc = "(no description)"
				}
				// Mark default command
				suffix := ""
				if name == config.Default {
					suffix = " (default)"
				}
				fmt.Printf("  %-*s  %s%s\n", maxLen, name, desc, suffix)
			}
			fmt.Println()
		}
	}

	fmt.Println("Run 'x <command> --help' for command-specific help.")
}

// PrintCommandHelp prints help for a specific command
func PrintCommandHelp(name string, cmd Command) {
	desc := cmd.Description
	if desc == "" {
		desc = "(no description)"
	}
	fmt.Printf("%s: %s\n", name, desc)

	if len(cmd.Args) > 0 {
		fmt.Println()
		fmt.Println("Arguments:")

		// Find max arg name length for alignment
		maxLen := 0
		for _, arg := range cmd.Args {
			if len(arg.Name) > maxLen {
				maxLen = len(arg.Name)
			}
		}

		for _, arg := range cmd.Args {
			argDesc := arg.Description
			if argDesc == "" {
				argDesc = "(no description)"
			}
			suffix := ""
			if arg.Rest {
				suffix = " (captures remaining args)"
			}
			fmt.Printf("  %-*s  %s%s\n", maxLen, arg.Name, argDesc, suffix)
		}
	}

	// Show usage example
	fmt.Println()
	fmt.Printf("Usage: x %s", name)
	for _, arg := range cmd.Args {
		if arg.Rest {
			fmt.Printf(" <%s>...", arg.Name)
		} else {
			fmt.Printf(" <%s>", arg.Name)
		}
	}
	fmt.Println()
}
