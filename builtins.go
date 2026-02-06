package main

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed builtins.yaml
var builtinsYAML []byte

// getBuiltinCommands returns the built-in commands that are always available.
// These are kept up-to-date with each release and can be overridden by user config.
func getBuiltinCommands() (map[string]Command, error) {
	var raw map[string]any
	if err := yaml.Unmarshal(builtinsYAML, &raw); err != nil {
		return nil, err
	}

	commands := make(map[string]Command)
	for name, value := range raw {
		valueData, err := yaml.Marshal(value)
		if err != nil {
			continue
		}

		var cmd Command
		if err := yaml.Unmarshal(valueData, &cmd); err != nil {
			continue
		}

		cmd.Source = "built-in"
		commands[name] = cmd
	}

	return commands, nil
}
