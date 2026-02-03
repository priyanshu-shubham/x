package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	// Handle --help and -h flags
	if len(os.Args) >= 2 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		config, _ := LoadCommandsConfig()
		if config == nil {
			config = &CommandsConfig{
				Default:  "shell",
				Commands: make(map[string]Command),
			}
		}
		PrintHelp(config)
		return
	}

	// Handle built-in commands
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case CmdConfigure:
			if err := RunConfigure(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return

		case CmdCommands:
			if err := RunCommandsEditor(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return

		case CmdUsage:
			usage, err := LoadUsage()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading usage: %v\n", err)
				os.Exit(1)
			}
			usage.Display()
			return

		case CmdUpgrade:
			if err := RunUpgrade(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return

		case CmdVersion:
			fmt.Printf("x version %s\n", Version)
			return
		}
	}

	// Load commands configuration
	commandsConfig, err := LoadCommandsConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading commands: %v\n", err)
		os.Exit(1)
	}

	// No args provided - show help or use default
	if len(os.Args) < 2 {
		PrintHelp(commandsConfig)
		os.Exit(1)
	}

	// Check if first arg is --help for a command (e.g., "x cmd --help")
	firstArg := os.Args[1]

	// Check if first arg is a known command
	cmd, isCommand := commandsConfig.Commands[firstArg]

	if isCommand {
		// Check for command-specific help
		if len(os.Args) >= 3 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
			PrintCommandHelp(firstArg, cmd)
			return
		}
	}

	// Load configuration for API access
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'x configure' to set up authentication.\n")
		os.Exit(1)
	}

	// Create API client
	ctx := context.Background()
	client, err := CreateClient(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	// Route to appropriate handler
	if isCommand {
		// Run the matched command with remaining args
		if _, err := RunPipeline(client, config.AuthType, commandsConfig, cmd, os.Args[2:], false); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// No command matched - use default command with all args
	defaultCmd, hasDefault := commandsConfig.Commands[commandsConfig.Default]
	if hasDefault {
		// Use all args as input to the default command
		if _, err := RunPipeline(client, config.AuthType, commandsConfig, defaultCmd, os.Args[1:], false); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Fallback: no default command configured, show help
	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", firstArg)
	fmt.Fprintf(os.Stderr, "Run 'x --help' to see available commands.\n")
	os.Exit(1)
}
