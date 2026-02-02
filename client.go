package main

import (
	"context"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/vertex"
)

// CreateClient creates an Anthropic client based on the configuration
func CreateClient(ctx context.Context, config *Config) (anthropic.Client, error) {
	switch config.AuthType {
	case AuthTypeAPIKey:
		return anthropic.NewClient(option.WithAPIKey(config.APIKey)), nil
	case AuthTypeVertex:
		return anthropic.NewClient(vertex.WithGoogleAuth(ctx, config.Region, config.ProjectID)), nil
	default:
		return anthropic.Client{}, ErrInvalidAuthType
	}
}

// GenerateResponse sends a prompt to the model and returns the response
func GenerateResponse(client anthropic.Client, systemPrompt, userQuery string) (string, error) {
	ctx := context.Background()

	stream := client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     DefaultModel,
		MaxTokens: DefaultMaxTokens,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userQuery)),
		},
	})

	var response strings.Builder

	for stream.Next() {
		event := stream.Current()
		if event.Type == "content_block_delta" {
			if event.Delta.Type == "text_delta" {
				response.WriteString(event.Delta.Text)
			}
		}
	}

	if err := stream.Err(); err != nil {
		return "", err
	}

	return strings.TrimSpace(response.String()), nil
}
