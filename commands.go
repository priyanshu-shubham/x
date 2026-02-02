package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// RunConfigure handles the configure command
func RunConfigure() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Choose authentication type:")
	fmt.Println("1. Anthropic API Key")
	fmt.Println("2. Google Cloud Vertex AI")
	fmt.Print("Enter choice (1 or 2): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	config := &Config{}

	switch choice {
	case "1":
		config.AuthType = AuthTypeAPIKey
		fmt.Print("Enter your Anthropic API key: ")
		apiKey, _ := reader.ReadString('\n')
		config.APIKey = strings.TrimSpace(apiKey)

	case "2":
		config.AuthType = AuthTypeVertex
		fmt.Print("Enter your Google Cloud Project ID: ")
		projectID, _ := reader.ReadString('\n')
		config.ProjectID = strings.TrimSpace(projectID)

		fmt.Print("Enter your Vertex AI region (e.g., us-east5): ")
		region, _ := reader.ReadString('\n')
		config.Region = strings.TrimSpace(region)

	default:
		return fmt.Errorf("%w: %s", ErrInvalidChoice, choice)
	}

	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Configuration saved successfully!")
	return nil
}

// RunSubcommandsEditor opens the subcommands configuration file in an editor
func RunSubcommandsEditor() error {
	path, err := EnsureSubcommandsFile()
	if err != nil {
		return fmt.Errorf("failed to create subcommands file: %w", err)
	}

	fmt.Printf("Opening %s\n", path)
	if err := OpenEditor(path); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	return nil
}

// RunSubcommand executes a custom subcommand
func RunSubcommand(client anthropic.Client, subcmd Subcommand, args []string) error {
	query := strings.Join(args, " ")
	systemPrompt := ApplyTemplate(subcmd.Prompt)

	response, err := GenerateResponse(client, systemPrompt, query)
	if err != nil {
		return err
	}

	fmt.Println(response)
	return nil
}

// RunShellGeneration generates and optionally runs a shell command
func RunShellGeneration(client anthropic.Client, query string) error {
	systemPrompt := GetDefaultSystemPrompt()

	command, err := GenerateResponse(client, systemPrompt, query)
	if err != nil {
		return fmt.Errorf("error generating command: %w", err)
	}

	fmt.Printf("\n  %s\n\n", command)
	fmt.Print("Run this command? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "" || response == "y" || response == "yes" {
		fmt.Println()
		err := RunShellCommand(command)
		fmt.Println()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			return err
		}
	} else {
		fmt.Println("Cancelled.")
	}

	return nil
}
