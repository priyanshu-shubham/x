package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const UsageFileName = "usage.json"

const PricingURL = "https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json"

// ModelPricing holds pricing info for a model
type ModelPricing struct {
	InputCostPerToken         float64 `json:"input_cost_per_token"`
	OutputCostPerToken        float64 `json:"output_cost_per_token"`
	CacheCreationCostPerToken float64 `json:"cache_creation_input_token_cost"`
	CacheReadCostPerToken     float64 `json:"cache_read_input_token_cost"`
}

// FetchModelPricing fetches pricing for the current model from LiteLLM's pricing data
func FetchModelPricing() (*ModelPricing, error) {
	resp, err := http.Get(PricingURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var prices map[string]ModelPricing
	if err := json.NewDecoder(resp.Body).Decode(&prices); err != nil {
		return nil, err
	}

	// Generate possible key formats from DefaultModel
	// DefaultModel format: "claude-sonnet-4-5@20250929"
	keysToTry := []string{DefaultModel}

	// Without @version suffix
	if idx := strings.Index(DefaultModel, "@"); idx != -1 {
		base := DefaultModel[:idx]
		version := DefaultModel[idx+1:]
		keysToTry = append(keysToTry, base)
		// With dash instead of @
		keysToTry = append(keysToTry, base+"-"+version)
		// With anthropic prefix
		keysToTry = append(keysToTry, "anthropic."+base+"-"+version+"-v1:0")
	}

	for _, key := range keysToTry {
		if pricing, ok := prices[key]; ok {
			return &pricing, nil
		}
	}

	return nil, fmt.Errorf("pricing not found for model %s", DefaultModel)
}

// Usage tracks cumulative token usage
type Usage struct {
	InputTokens         int64 `json:"input_tokens"`
	OutputTokens        int64 `json:"output_tokens"`
	ThinkingTokens      int64 `json:"thinking_tokens"`
	CacheCreationTokens int64 `json:"cache_creation_tokens"`
	CacheReadTokens     int64 `json:"cache_read_tokens"`
	RequestCount        int64 `json:"request_count"`
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
func (u *Usage) Add(input, output, thinking, cacheCreation, cacheRead int64) {
	u.InputTokens += input
	u.OutputTokens += output
	u.ThinkingTokens += thinking
	u.CacheCreationTokens += cacheCreation
	u.CacheReadTokens += cacheRead
	u.RequestCount++
}

// Reset clears all usage data
func (u *Usage) Reset() {
	u.InputTokens = 0
	u.OutputTokens = 0
	u.ThinkingTokens = 0
	u.CacheCreationTokens = 0
	u.CacheReadTokens = 0
	u.RequestCount = 0
}

// Display prints the usage summary
func (u *Usage) Display() {
	fmt.Println("Token Usage")
	fmt.Println("-----------")
	fmt.Printf("Input tokens:          %d\n", u.InputTokens)
	fmt.Printf("Output tokens:         %d\n", u.OutputTokens)
	if u.ThinkingTokens > 0 {
		fmt.Printf("Thinking tokens:       %d\n", u.ThinkingTokens)
	}
	if u.CacheCreationTokens > 0 {
		fmt.Printf("Cache creation tokens: %d\n", u.CacheCreationTokens)
	}
	if u.CacheReadTokens > 0 {
		fmt.Printf("Cache read tokens:     %d\n", u.CacheReadTokens)
	}
	fmt.Printf("Total tokens:          %d\n", u.InputTokens+u.OutputTokens+u.ThinkingTokens)
	fmt.Printf("Requests:              %d\n", u.RequestCount)

	// Fetch and display cost
	pricing, err := FetchModelPricing()
	if err != nil {
		fmt.Printf("\nCost:                  (unable to fetch pricing)\n")
		return
	}

	inputCost := float64(u.InputTokens) * pricing.InputCostPerToken
	outputCost := float64(u.OutputTokens) * pricing.OutputCostPerToken
	cacheCreationCost := float64(u.CacheCreationTokens) * pricing.CacheCreationCostPerToken
	cacheReadCost := float64(u.CacheReadTokens) * pricing.CacheReadCostPerToken
	totalCost := inputCost + outputCost + cacheCreationCost + cacheReadCost

	fmt.Printf("\nEstimated Cost\n")
	fmt.Printf("--------------\n")
	fmt.Printf("Input cost:            $%.4f\n", inputCost)
	fmt.Printf("Output cost:           $%.4f\n", outputCost)
	if u.CacheCreationTokens > 0 {
		fmt.Printf("Cache creation cost:   $%.4f\n", cacheCreationCost)
	}
	if u.CacheReadTokens > 0 {
		fmt.Printf("Cache read cost:       $%.4f\n", cacheReadCost)
	}
	fmt.Printf("Total cost:            $%.4f\n", totalCost)
}
