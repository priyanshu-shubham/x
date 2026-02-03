package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
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

// getMarkdownStyle returns a custom style for markdown rendering
func getMarkdownStyle() ansi.StyleConfig {
	// Start with Tokyo Night style as base
	style := styles.DarkStyleConfig

	// Small left indent for readability
	style.Document.Margin = uintPtr(1)

	// Remove header numbering by setting Format to empty
	style.H1.Format = ""
	style.H2.Format = ""
	style.H3.Format = ""
	style.H4.Format = ""
	style.H5.Format = ""
	style.H6.Format = ""

	// Add margin to code blocks
	style.CodeBlock.StyleBlock.Margin = uintPtr(2)

	return style
}

func uintPtr(v uint) *uint {
	return &v
}

// addCodeBlockBorder adds a vertical border to code blocks in rendered output
func addCodeBlockBorder(rendered string) string {
	lines := strings.Split(rendered, "\n")
	var result []string
	inCodeBlock := false

	for _, line := range lines {
		stripped := stripAnsi(line)
		trimmed := strings.TrimSpace(stripped)

		// Heuristic: code blocks have 4+ spaces of indentation (margin of 2 = 4 spaces)
		hasCodeIndent := strings.HasPrefix(stripped, "    ") && len(trimmed) > 0

		if hasCodeIndent && !inCodeBlock {
			inCodeBlock = true
		} else if !hasCodeIndent && inCodeBlock && len(trimmed) > 0 {
			inCodeBlock = false
		}

		if inCodeBlock && len(trimmed) > 0 {
			// Find position after leading ANSI codes and initial spaces to insert the bar
			line = insertBarAfterIndent(line, 2)
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// insertBarAfterIndent inserts "│ " after the specified number of visual spaces
func insertBarAfterIndent(line string, afterSpaces int) string {
	ansiPattern := regexp.MustCompile(`\x1b\[[0-9;]*m`)

	var result strings.Builder
	visualSpaces := 0
	inserted := false
	i := 0

	for i < len(line) {
		// Check for ANSI escape sequence
		if loc := ansiPattern.FindStringIndex(line[i:]); loc != nil && loc[0] == 0 {
			result.WriteString(line[i : i+loc[1]])
			i += loc[1]
			continue
		}

		// Check if we've passed enough spaces to insert the bar
		if !inserted && line[i] == ' ' {
			visualSpaces++
			result.WriteByte(line[i])
			i++
			if visualSpaces >= afterSpaces {
				result.WriteString("│ ")
				inserted = true
			}
			continue
		}

		// If we hit a non-space before inserting, insert now
		if !inserted {
			result.WriteString("│ ")
			inserted = true
		}

		result.WriteByte(line[i])
		i++
	}

	return result.String()
}

// stripAnsi removes ANSI escape codes from a string
func stripAnsi(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}
