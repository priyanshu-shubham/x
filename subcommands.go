package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

const subcommandsTemplate = `# X CLI Subcommands Configuration
#
# Define custom subcommands with step-based pipelines.
# Each step can be: exec (shell), llm (single call), or agentic (multi-turn loop).
#
# Template variables available in prompts:
#   {{time}}      - Current time (HH:MM:SS)
#   {{date}}      - Current date (YYYY-MM-DD)
#   {{datetime}}  - Current date and time
#   {{directory}} - Current working directory
#   {{os}}        - Operating system (darwin, linux, windows)
#   {{arch}}      - Architecture (amd64, arm64)
#   {{shell}}     - User's shell
#   {{user}}      - Current username
#
# Step-specific variables:
#   {{args.<name>}}         - Named argument value
#   {{output}}              - Output from previous step
#   {{steps.<id>.output}}   - Output from a specific named step
#
# OS-specific exec commands:
#   exec:
#     command: "cat file"    # Default (used if no OS match)
#     windows: "type file"   # Windows-specific
#     darwin: "cat file"     # macOS-specific
#     linux: "cat file"      # Linux-specific
#
# Default subcommand (runs when no command matches):
default: shell

# Shell command generation (the default behavior)
shell:
  description: Generate and run shell commands from natural language
  args:
    - name: query
      description: What you want to do
      rest: true
  steps:
    - llm:
        system: |
          You are a command-line assistant. Generate the appropriate shell command.
          Environment: {{os}}, {{arch}}, {{directory}}, {{shell}}
          Rules: Output ONLY the command, no explanations, no markdown.
        prompt: "{{args.query}}"
        silent: true
    - exec:
        command: "{{output}}"
        confirm: true

# Create new subcommands using AI
new:
  description: Create a new subcommand with AI assistance
  args:
    - name: description
      description: What the new subcommand should do
      rest: true
  steps:
    - id: current
      exec:
        command: "cat ~/.config/x/subcommands.yaml"
        darwin: "cat ~/Library/Application\\ Support/x/subcommands.yaml"
        windows: "type %LOCALAPPDATA%\\x\\subcommands.yaml"
        silent: true
    - agentic:
        system: |
          You are a helper that creates new subcommands for the x CLI tool.

          YAML SCHEMA FOR SUBCOMMANDS:

          subcommand-name:
            description: Short description shown in help
            args:                          # Optional: define named arguments
              - name: argname              # Referenced as args.argname in double braces
                description: What this arg is for
                rest: true                 # Optional: captures all remaining args
            steps:                         # List of steps to execute in order
              - llm:                       # Single LLM call
                  system: "System prompt"
                  prompt: "User prompt"
              - exec:                      # Shell command
                  command: "default cmd"   # Required: default command
                  windows: "win cmd"       # Optional: Windows-specific
                  darwin: "mac cmd"        # Optional: macOS-specific
                  linux: "linux cmd"       # Optional: Linux-specific
                  confirm: true            # Optional: ask before running (default: false)
              - agentic:                   # Multi-turn with shell access
                  system: "System prompt"
                  prompt: "User prompt"
                  max_iterations: 10       # Optional (default: 10)
                  auto_execute: false      # Optional: auto-run commands (default: false)

          AVAILABLE TEMPLATE VARIABLES (use double curly braces):
          - args.<name> - Named argument value
          - output - Output from previous step
          - steps.<id>.output - Output from named step (give step an id: field)
          - directory, os, arch, shell, user - Environment info
          - time, date, datetime - Current time/date

          STEP TYPES:
          1. llm: Single AI call. Good for text generation, explanations, summaries.
          2. exec: Run shell command. Use for reading files, running tools. Add OS variants for cross-platform.
          3. agentic: Multi-turn AI loop with shell access. Good for complex tasks needing multiple commands.

          COMMON PATTERNS:
          - Read file then analyze: exec (cat file) -> llm (analyze the output variable)
          - Generate and run: llm (generate command) -> exec with confirm: true
          - Complex task: agentic with auto_execute: false for safety

          RULES:
          1. Create valid YAML that can be appended to the subcommands file
          2. Use descriptive names (kebab-case)
          3. Write clear descriptions for help text
          4. Use OS-specific commands when needed (cat vs type, grep vs findstr)
          5. For dangerous operations, use confirm: true or auto_execute: false
          6. Keep system prompts short and concise - no fluff, just clear instructions

          Current OS: {{os}}
          Config file location:
          - Linux: ~/.config/x/subcommands.yaml
          - macOS: ~/Library/Application Support/x/subcommands.yaml
          - Windows: %LOCALAPPDATA%\x\subcommands.yaml

          After creating the YAML, append it to the config file using echo/cat or appropriate command.
        prompt: |
          Create a new subcommand for: {{args.description}}

          Current subcommands.yaml content:
          {{steps.current.output}}
        max_iterations: 10
        auto_execute: false

# Example: Simple LLM subcommand
# paraphrase:
#   description: Paraphrase text to be clearer and more professional
#   args:
#     - name: text
#       description: Text to paraphrase
#       rest: true
#   steps:
#     - llm:
#         system: |
#           You are a writing assistant. Paraphrase the text clearly and professionally.
#           Output ONLY the paraphrased text.
#         prompt: "{{args.text}}"

# Example: Multi-step with file reading (cross-platform)
# analyze:
#   description: Analyze a code file
#   args:
#     - name: file
#       description: Path to the file to analyze
#   steps:
#     - id: read
#       exec:
#         command: "cat {{args.file}}"
#         windows: "type {{args.file}}"
#     - llm:
#         system: "You are a code analyst. Analyze the code and provide insights."
#         prompt: "{{output}}"

# Example: Agentic workflow
# fix:
#   description: Automatically fix issues in code
#   args:
#     - name: task
#       description: What to fix
#       rest: true
#   steps:
#     - agentic:
#         system: |
#           You are a code fixer. Analyze and fix issues in the codebase.
#           Use shell commands to read/write files as needed.
#           Current directory: {{directory}}
#         prompt: "{{args.task}}"
#         max_iterations: 10
#         auto_execute: false
`

// Step types for pipeline execution

// ExecStep executes a shell command
type ExecStep struct {
	Command string `yaml:"command"`         // Default command (used if no OS-specific command matches)
	Windows string `yaml:"windows"`         // Windows-specific command
	Darwin  string `yaml:"darwin"`          // macOS-specific command
	Linux   string `yaml:"linux"`           // Linux-specific command
	Confirm bool   `yaml:"confirm"`         // Prompt before execution
	Silent  bool   `yaml:"silent"`          // Don't print output (for intermediate steps)
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

// SubcommandStep calls another subcommand
type SubcommandStep struct {
	Name   string   `yaml:"name"`   // Name of the subcommand to call
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

// Arg represents a named argument for a subcommand
type Arg struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Rest        bool   `yaml:"rest"` // Capture remaining args as one string
}

// Subcommand represents a custom subcommand configuration
type Subcommand struct {
	Description string `yaml:"description"`
	Args        []Arg  `yaml:"args"`
	Steps       []Step `yaml:"steps"`
	Source      string `yaml:"-"` // Where this subcommand was loaded from (not in YAML)
}

// SubcommandsConfig holds all subcommands and the default
type SubcommandsConfig struct {
	Default     string
	Subcommands map[string]Subcommand
}

// getSubcommandsPath returns the global subcommands config file path
func getSubcommandsPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, SubcommandsFileName), nil
}

// findAllLocalSubcommandsFiles searches for xcommands.yaml files from root to cwd
// Returns paths ordered from root to current directory (so later ones override earlier)
func findAllLocalSubcommandsFiles() []string {
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
		path := filepath.Join(d, LocalSubcommandsFileName)
		if _, err := os.Stat(path); err == nil {
			paths = append(paths, path)
		}
	}

	return paths
}

// LoadSubcommandsConfig reads and merges subcommand configurations.
// Configs are merged with this precedence (later overrides earlier):
// 1. Global subcommands.yaml
// 2. Parent xcommands.yaml files (from root to current directory)
// 3. Local xcommands.yaml (in current directory)
func LoadSubcommandsConfig() (*SubcommandsConfig, error) {
	config := &SubcommandsConfig{
		Default:     "shell",
		Subcommands: make(map[string]Subcommand),
	}

	// Load global config first
	globalPath, err := getSubcommandsPath()
	if err != nil {
		return nil, err
	}

	globalData, err := os.ReadFile(globalPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the default subcommands file automatically
			if err := os.MkdirAll(filepath.Dir(globalPath), DirPerms); err != nil {
				return nil, err
			}
			if err := os.WriteFile(globalPath, []byte(subcommandsTemplate), SubcommandsPerms); err != nil {
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
	localPaths := findAllLocalSubcommandsFiles()
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
func mergeConfig(config *SubcommandsConfig, data []byte, source string) error {
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

	// Parse and merge subcommands
	for name, value := range raw {
		valueData, err := yaml.Marshal(value)
		if err != nil {
			continue
		}

		var subcmd Subcommand
		if err := yaml.Unmarshal(valueData, &subcmd); err != nil {
			continue
		}

		subcmd.Source = source
		config.Subcommands[name] = subcmd
	}

	return nil
}

// LoadSubcommands reads and parses the subcommands configuration (legacy compatibility)
func LoadSubcommands() (map[string]Subcommand, error) {
	config, err := LoadSubcommandsConfig()
	if err != nil {
		return nil, err
	}
	return config.Subcommands, nil
}

// EnsureSubcommandsFile creates the subcommands file if it doesn't exist
func EnsureSubcommandsFile() (string, error) {
	path, err := getSubcommandsPath()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), DirPerms); err != nil {
		return "", err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte(subcommandsTemplate), SubcommandsPerms); err != nil {
			return "", err
		}
	}

	return path, nil
}

// IsReservedCommand checks if a command name is reserved
func IsReservedCommand(name string) bool {
	switch name {
	case CmdConfigure, CmdSubcommands, CmdUsage, CmdUpgrade, CmdVersion:
		return true
	default:
		return false
	}
}

// PrintHelp prints the main help message with all available subcommands
func PrintHelp(config *SubcommandsConfig) {
	fmt.Println("Usage: x <command> [args]")
	fmt.Println()
	fmt.Println("Built-in commands:")
	fmt.Println("  configure     Configure authentication")
	fmt.Println("  subcommands   Edit custom subcommands")
	fmt.Println("  usage         Show token usage and cost")
	fmt.Println("  upgrade       Upgrade to latest version")
	fmt.Println("  version       Show current version")
	fmt.Println()

	if len(config.Subcommands) > 0 {
		// Group subcommands by source
		bySource := make(map[string][]string)
		for name, subcmd := range config.Subcommands {
			bySource[subcmd.Source] = append(bySource[subcmd.Source], name)
		}

		// Sort sources: global first, then others alphabetically
		var sources []string
		for source := range bySource {
			sources = append(sources, source)
		}
		sort.Slice(sources, func(i, j int) bool {
			if sources[i] == "global" {
				return true
			}
			if sources[j] == "global" {
				return false
			}
			return sources[i] < sources[j]
		})

		// Find max name length across all subcommands for alignment
		maxLen := 0
		for name := range config.Subcommands {
			if len(name) > maxLen {
				maxLen = len(name)
			}
		}

		for _, source := range sources {
			names := bySource[source]
			sort.Strings(names)

			// Print source header
			fmt.Printf("Subcommands (%s):\n", source)

			for _, name := range names {
				subcmd := config.Subcommands[name]
				desc := subcmd.Description
				if desc == "" {
					desc = "(no description)"
				}
				// Mark default subcommand
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

// PrintSubcommandHelp prints help for a specific subcommand
func PrintSubcommandHelp(name string, subcmd Subcommand) {
	desc := subcmd.Description
	if desc == "" {
		desc = "(no description)"
	}
	fmt.Printf("%s: %s\n", name, desc)

	if len(subcmd.Args) > 0 {
		fmt.Println()
		fmt.Println("Arguments:")

		// Find max arg name length for alignment
		maxLen := 0
		for _, arg := range subcmd.Args {
			if len(arg.Name) > maxLen {
				maxLen = len(arg.Name)
			}
		}

		for _, arg := range subcmd.Args {
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
	for _, arg := range subcmd.Args {
		if arg.Rest {
			fmt.Printf(" <%s>...", arg.Name)
		} else {
			fmt.Printf(" <%s>", arg.Name)
		}
	}
	fmt.Println()
}
