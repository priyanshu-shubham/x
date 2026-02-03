package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	// Handle --help and -h flags
	if len(os.Args) >= 2 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		config, _ := LoadSubcommandsConfig()
		if config == nil {
			config = &SubcommandsConfig{
				Default:     "shell",
				Subcommands: make(map[string]Subcommand),
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

		case CmdSubcommands:
			if err := RunSubcommandsEditor(); err != nil {
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

	// Load subcommands configuration
	subcommandsConfig, err := LoadSubcommandsConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading subcommands: %v\n", err)
		os.Exit(1)
	}

	// No args provided - show help or use default
	if len(os.Args) < 2 {
		PrintHelp(subcommandsConfig)
		os.Exit(1)
	}

	// Check if first arg is --help for a subcommand (e.g., "x subcmd --help")
	firstArg := os.Args[1]

	// Check if first arg is a known subcommand
	subcmd, isSubcommand := subcommandsConfig.Subcommands[firstArg]

	if isSubcommand {
		// Check for subcommand-specific help
		if len(os.Args) >= 3 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
			PrintSubcommandHelp(firstArg, subcmd)
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
	if isSubcommand {
		// Run the matched subcommand with remaining args
		if _, err := RunPipeline(client, config.AuthType, subcommandsConfig, subcmd, os.Args[2:], false); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// No subcommand matched - use default subcommand with all args
	defaultSubcmd, hasDefault := subcommandsConfig.Subcommands[subcommandsConfig.Default]
	if hasDefault {
		// Use all args as input to the default subcommand
		if _, err := RunPipeline(client, config.AuthType, subcommandsConfig, defaultSubcmd, os.Args[1:], false); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Fallback: no default subcommand configured, show help
	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", firstArg)
	fmt.Fprintf(os.Stderr, "Run 'x --help' to see available commands.\n")
	os.Exit(1)
}
