package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// AIClient handles interactions with OpenAI and Grok APIs
type AIClient struct {
	openaiClient *openai.Client
	openaiAPIKey string
	xaiAPIKey    string
	httpClient   *http.Client
	logger       *slog.Logger
}

// NewAIClient creates a new AI client with proper timeouts
func NewAIClient(openaiAPIKey, xaiAPIKey string, logger *slog.Logger) *AIClient {
	var oaClient *openai.Client
	if openaiAPIKey != "" {
		oaClient = openai.NewClient(openaiAPIKey)
	}

	return &AIClient{
		openaiClient: oaClient,
		openaiAPIKey: openaiAPIKey,
		xaiAPIKey:    xaiAPIKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger,
	}
}

// AskClient sends a prompt to OpenAI or Grok with a system message and returns the response
func (c *AIClient) AskClient(ctx context.Context, prompt, systemMessage, model, provider string, maxTokens int) (string, error) {
	c.logger.InfoContext(ctx, "sending AI request",
		"provider", provider,
		"model", model,
		"max_tokens", maxTokens,
		"prompt_length", len(prompt))

	switch provider {
	case ProviderOpenAI:
		if c.openaiClient == nil {
			return "", NewValidationError("OPENAI_API_KEY", "OpenAI support is deprecated; set OPENAI_API_KEY to enable overrides")
		}
		return c.askOpenAI(ctx, prompt, systemMessage, model, maxTokens)
	default:
		return c.askGrok(ctx, prompt, systemMessage, model, maxTokens)
	}
}

// askOpenAI sends a request to OpenAI API
func (c *AIClient) askOpenAI(ctx context.Context, prompt, systemMessage, model string, maxTokens int) (string, error) {
	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp, err := c.openaiClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:     model,
			MaxTokens: maxTokens,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemMessage,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		c.logger.ErrorContext(ctx, "OpenAI API error", "error", err)
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		c.logger.ErrorContext(ctx, "no response from OpenAI")
		return "", NewAPIError("OpenAI", 0, "no response from OpenAI", nil)
	}

	c.logger.InfoContext(ctx, "received OpenAI response",
		"response_length", len(resp.Choices[0].Message.Content),
		"finish_reason", resp.Choices[0].FinishReason)

	return resp.Choices[0].Message.Content, nil
}

// askGrok sends a request to Grok API
func (c *AIClient) askGrok(ctx context.Context, prompt, systemMessage, model string, maxTokens int) (string, error) {
	if c.xaiAPIKey == "" {
		return "", NewValidationError("XAI_API_KEY", "environment variable not set")
	}

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	grokModel := model
	if grokModel == "" {
		grokModel = DefaultGrokModel
	}

	if maxTokens == 0 {
		maxTokens = DefaultMaxTokens
	}

	requestBody := map[string]interface{}{
		"model": grokModel,
		"messages": []map[string]string{
			{"role": "system", "content": systemMessage},
			{"role": "user", "content": prompt},
		},
		"max_tokens": maxTokens,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.x.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.xaiAPIKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "Grok API request failed", "error", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "Grok API error",
			"status_code", resp.StatusCode,
			"response_body", string(body))
		return "", NewAPIError("Grok", resp.StatusCode, string(body), nil)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		c.logger.ErrorContext(ctx, "no response from Grok")
		return "", NewAPIError("Grok", resp.StatusCode, "no response from Grok", nil)
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content, ok := message["content"].(string)
	if !ok {
		return "", NewAPIError("Grok", resp.StatusCode, "invalid response format from Grok", nil)
	}

	c.logger.InfoContext(ctx, "received Grok response",
		"response_length", len(content))

	return content, nil
}

// ImageOpinionOpenAI sends an image to OpenAI's vision endpoint
func (c *AIClient) ImageOpinionOpenAI(ctx context.Context, imageURL, systemMessage, model string, maxTokens int, customPrompt *string) (string, error) {
	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	c.logger.InfoContext(ctx, "processing image with OpenAI", "image_url", imageURL)

	// Download and encode image
	base64Image, err := c.downloadAndEncodeImage(ctx, imageURL)
	if err != nil {
		return "", fmt.Errorf("error downloading or encoding image: %w", err)
	}

	promptText := "Form an opinion on this image. Try to be controversial or humorous."
	if customPrompt != nil && *customPrompt != "" {
		promptText = *customPrompt
	}

	requestBody := map[string]interface{}{
		"model":      model,
		"max_tokens": maxTokens,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": systemMessage,
			},
			{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": promptText},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.openaiAPIKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "OpenAI vision API request failed", "error", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "OpenAI vision API error",
			"status_code", resp.StatusCode,
			"response_body", string(body))
		return "", NewAPIError("OpenAI", resp.StatusCode, string(body), nil)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", NewAPIError("OpenAI", resp.StatusCode, "no response from OpenAI", nil)
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content, ok := message["content"].(string)
	if !ok {
		return "", NewAPIError("OpenAI", resp.StatusCode, "invalid response format from OpenAI", nil)
	}

	c.logger.InfoContext(ctx, "received OpenAI vision response",
		"response_length", len(content))

	return content, nil
}

// ImageOpinionGrok sends an image to Grok API
func (c *AIClient) ImageOpinionGrok(ctx context.Context, imageURL, systemMessage string, customPrompt *string) (string, error) {
	if c.xaiAPIKey == "" {
		return "", NewValidationError("XAI_API_KEY", "environment variable not set")
	}

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	c.logger.InfoContext(ctx, "processing image with Grok", "image_url", imageURL)

	base64Image, err := c.downloadAndEncodeImage(ctx, imageURL)
	if err != nil {
		return "", fmt.Errorf("error downloading or encoding image: %w", err)
	}

	promptText := "Form an opinion on this image. Try to be controversial or humorous."
	if customPrompt != nil && *customPrompt != "" {
		promptText = *customPrompt
	}

	requestBody := map[string]interface{}{
		"model": "grok-vision-beta",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": systemMessage,
			},
			{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": promptText},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url":    fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
							"detail": "high",
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.x.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.xaiAPIKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "Grok vision API request failed", "error", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "Grok vision API error",
			"status_code", resp.StatusCode,
			"response_body", string(body))
		return "", NewAPIError("Grok", resp.StatusCode, string(body), nil)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", NewAPIError("Grok", resp.StatusCode, "no response from Grok", nil)
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content, ok := message["content"].(string)
	if !ok {
		return "", NewAPIError("Grok", resp.StatusCode, "invalid response format from Grok", nil)
	}

	c.logger.InfoContext(ctx, "received Grok vision response",
		"response_length", len(content))

	return content, nil
}

// downloadAndEncodeImage downloads an image from URL and returns base64 encoded string
func (c *AIClient) downloadAndEncodeImage(ctx context.Context, imageURL string) (string, error) {
	// Add timeout to context if not already set
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(imageData), nil
}

// SuggestMessageBreaks uses AI to intelligently break a message into natural chunks
// that feel more human and conversational, like following up thoughts with additional messages
func (c *AIClient) SuggestMessageBreaks(ctx context.Context, message string) ([]string, error) {
	// If message is short enough, return as-is
	if len(message) <= 500 {
		return []string{message}, nil
	}

	c.logger.InfoContext(ctx, "requesting message break suggestions",
		"message_length", len(message))

	systemPrompt := `You are a message chunking assistant. Your job is to break up messages into natural, conversational chunks 
that feel like how humans text - following up one message with more messages as they flesh out their thoughts.

Rules:
1. Split at natural thought boundaries (paragraphs, topic shifts, etc.)
2. Each chunk should be a complete thought or idea
3. Aim for 3-5 chunks for longer messages
4. Preserve the exact original text - no changes to content
5. Respond ONLY with the chunks separated by the delimiter: <<<BREAK>>>
6. Do not add any explanations or commentary

Example input: "I think pizza is great. It has cheese and sauce. But honestly, the best part is the crust when it's done right. Brooklyn style is my favorite."

Example output: "I think pizza is great. It has cheese and sauce.<<<BREAK>>>But honestly, the best part is the crust when it's done right.<<<BREAK>>>Brooklyn style is my favorite."`

	userPrompt := fmt.Sprintf("Break this message into natural conversational chunks:\n\n%s", message)

	// Use Grok as default for this utility function (faster and cheaper)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	response, err := c.askGrok(ctx, userPrompt, systemPrompt, DefaultGrokModel, 1000)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get message breaks, falling back to simple chunking", "error", err)
		// Fallback to simple paragraph-based chunking
		return c.fallbackMessageBreaks(message), nil
	}

	// Parse the response
	chunks := c.parseMessageBreaks(response, message)

	c.logger.InfoContext(ctx, "message broken into chunks",
		"original_length", len(message),
		"chunk_count", len(chunks))

	return chunks, nil
}

// parseMessageBreaks parses the AI response and validates chunks
func (c *AIClient) parseMessageBreaks(response, originalMessage string) []string {
	chunks := []string{}
	parts := []string{}

	// Split by the delimiter
	for _, part := range []string{response} {
		parts = append(parts, strings.Split(part, "<<<BREAK>>>")...)
	}

	// Clean and validate chunks
	for _, chunk := range parts {
		chunk = strings.TrimSpace(chunk)
		if chunk != "" && len(chunk) <= 2000 {
			chunks = append(chunks, chunk)
		}
	}

	// If parsing failed or resulted in just one chunk, fallback
	if len(chunks) <= 1 {
		return c.fallbackMessageBreaks(originalMessage)
	}

	return chunks
}

// fallbackMessageBreaks provides a simple fallback chunking strategy
func (c *AIClient) fallbackMessageBreaks(message string) []string {
	chunks := []string{}

	// Try splitting by double newlines (paragraphs)
	paragraphs := strings.Split(message, "\n\n")

	currentChunk := ""
	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// If adding this paragraph would exceed a reasonable chunk size, start new chunk
		if len(currentChunk) > 0 && len(currentChunk)+len(para)+2 > 800 {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = para
		} else {
			if len(currentChunk) > 0 {
				currentChunk += "\n\n" + para
			} else {
				currentChunk = para
			}
		}
	}

	// Add the last chunk
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	// If we still have just one chunk or none, return the original message
	if len(chunks) <= 1 {
		return []string{message}
	}

	return chunks
}
