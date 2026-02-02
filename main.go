package main

import (
	"context"
	"fmt"
	"os"
	"strings"
)

func printUsage() {
	fmt.Println("Usage: x <command>")
	fmt.Println("       x configure     - Configure authentication")
	fmt.Println("       x subcommands   - Edit custom subcommands")
	fmt.Println("       x tokens        - Show token usage")
	fmt.Println("       x upgrade       - Upgrade to latest version")
	fmt.Println("       x version       - Show current version")
	fmt.Println("       x <query>       - Generate and run a shell command")
	fmt.Println("       x <subcommand>  - Run a custom subcommand")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Handle built-in commands
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

	case CmdTokens:
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

	// Load configuration
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

	// Load subcommands
	subcommands, err := LoadSubcommands()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading subcommands: %v\n", err)
		os.Exit(1)
	}

	// Check if first arg is a custom subcommand
	if subcmd, exists := subcommands[os.Args[1]]; exists {
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: x %s <input>\n", os.Args[1])
			os.Exit(1)
		}

		if err := RunSubcommand(client, subcmd, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Default: generate shell command
	query := strings.Join(os.Args[1:], " ")
	if err := RunShellGeneration(client, query); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
