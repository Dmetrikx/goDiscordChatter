package ai

import "context"

// Client defines the interface for AI client operations
type Client interface {
	// AskClient sends a prompt to an AI provider and returns the response
	AskClient(ctx context.Context, prompt, systemMessage, model, provider string, maxTokens int) (string, error)

	// ImageOpinionOpenAI sends an image to OpenAI's vision endpoint
	ImageOpinionOpenAI(ctx context.Context, imageURL, systemMessage, model string, maxTokens int, customPrompt *string) (string, error)

	// ImageOpinionGrok sends an image to Grok's vision endpoint
	ImageOpinionGrok(ctx context.Context, imageURL, systemMessage string, customPrompt *string) (string, error)
}
