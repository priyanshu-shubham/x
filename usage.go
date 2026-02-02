package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const UsageFileName = "usage.json"

// Usage tracks cumulative token usage
type Usage struct {
	InputTokens    int64 `json:"input_tokens"`
	OutputTokens   int64 `json:"output_tokens"`
	ThinkingTokens int64 `json:"thinking_tokens"`
	RequestCount   int64 `json:"request_count"`
}

// getUsagePath returns the usage file path
func getUsagePath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, UsageFileName), nil
}

// LoadUsage reads the usage data from disk
func LoadUsage() (*Usage, error) {
	path, err := getUsagePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Usage{}, nil
		}
		return nil, err
	}

	var usage Usage
	if err := json.Unmarshal(data, &usage); err != nil {
		return nil, err
	}

	return &usage, nil
}

// Save writes the usage data to disk
func (u *Usage) Save() error {
	path, err := getUsagePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), DirPerms); err != nil {
		return err
	}

	data, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, ConfigFilePerms)
}

// Add adds token counts to the usage
func (u *Usage) Add(input, output, thinking int64) {
	u.InputTokens += input
	u.OutputTokens += output
	u.ThinkingTokens += thinking
	u.RequestCount++
}

// Reset clears all usage data
func (u *Usage) Reset() {
	u.InputTokens = 0
	u.OutputTokens = 0
	u.ThinkingTokens = 0
	u.RequestCount = 0
}

// Display prints the usage summary
func (u *Usage) Display() {
	fmt.Println("Token Usage")
	fmt.Println("-----------")
	fmt.Printf("Input tokens:    %d\n", u.InputTokens)
	fmt.Printf("Output tokens:   %d\n", u.OutputTokens)
	if u.ThinkingTokens > 0 {
		fmt.Printf("Thinking tokens: %d\n", u.ThinkingTokens)
	}
	fmt.Printf("Total tokens:    %d\n", u.InputTokens+u.OutputTokens+u.ThinkingTokens)
	fmt.Printf("Requests:        %d\n", u.RequestCount)
}
