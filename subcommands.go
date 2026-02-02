package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const subcommandsTemplate = `# X CLI Subcommands Configuration
#
# Define custom subcommands with their own system prompts.
# Each subcommand can use template variables that get replaced at runtime:
#
#   {{time}}      - Current time (HH:MM:SS)
#   {{date}}      - Current date (YYYY-MM-DD)
#   {{datetime}}  - Current date and time
#   {{directory}} - Current working directory
#   {{os}}        - Operating system (darwin, linux, windows)
#   {{arch}}      - Architecture (amd64, arm64)
#   {{shell}}     - User's shell
#   {{user}}      - Current username
#
# Example:
#
# paraphrase:
#   prompt: |
#     You are a writing assistant. Paraphrase the given text to be more clear and professional.
#     Output ONLY the paraphrased text, nothing else.
#
# explain:
#   prompt: |
#     You are a helpful assistant. Explain the given concept in simple terms.
#     Current time: {{datetime}}
#
# git-commit:
#   prompt: |
#     You are a git expert. Generate a concise commit message for the given changes.
#     The user is in directory: {{directory}}
#     Output ONLY the commit message, nothing else.

`

// Subcommand represents a custom subcommand configuration
type Subcommand struct {
	Prompt string `yaml:"prompt"`
}

// SubcommandsConfig maps subcommand names to their configurations
type SubcommandsConfig map[string]Subcommand

// getSubcommandsPath returns the subcommands config file path
func getSubcommandsPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, SubcommandsFileName), nil
}

// LoadSubcommands reads and parses the subcommands configuration
func LoadSubcommands() (SubcommandsConfig, error) {
	path, err := getSubcommandsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return SubcommandsConfig{}, nil
		}
		return nil, err
	}

	var subcommands SubcommandsConfig
	if err := yaml.Unmarshal(data, &subcommands); err != nil {
		return nil, err
	}

	if subcommands == nil {
		subcommands = SubcommandsConfig{}
	}

	return subcommands, nil
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
	case CmdConfigure, CmdSubcommands:
		return true
	default:
		return false
	}
}
