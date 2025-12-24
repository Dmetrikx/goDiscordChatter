package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

// AIClient handles interactions with OpenAI and Grok APIs
type AIClient struct {
	openaiClient *openai.Client
	openaiAPIKey string
	xaiAPIKey    string
}

// NewAIClient creates a new AI client
func NewAIClient(openaiAPIKey, xaiAPIKey string) *AIClient {
	return &AIClient{
		openaiClient: openai.NewClient(openaiAPIKey),
		openaiAPIKey: openaiAPIKey,
		xaiAPIKey:    xaiAPIKey,
	}
}

// AskClient sends a prompt to OpenAI or Grok with a system message and returns the response
func (c *AIClient) AskClient(prompt, systemMessage, model, provider string, maxTokens int) (string, error) {
	if provider == "grok" {
		return c.askGrok(prompt, systemMessage)
	}
	return c.askOpenAI(prompt, systemMessage, model, maxTokens)
}

// askOpenAI sends a request to OpenAI API
func (c *AIClient) askOpenAI(prompt, systemMessage, model string, maxTokens int) (string, error) {
	ctx := context.Background()

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
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// askGrok sends a request to Grok API
func (c *AIClient) askGrok(prompt, systemMessage string) (string, error) {
	if c.xaiAPIKey == "" {
		return "", fmt.Errorf("XAI_API_KEY environment variable not set")
	}

	requestBody := map[string]interface{}{
		"model": "grok-3",
		"messages": []map[string]string{
			{"role": "system", "content": systemMessage},
			{"role": "user", "content": prompt},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.x.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.xaiAPIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Grok API error: %d - %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no response from Grok")
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format from Grok")
	}

	return content, nil
}

// ImageOpinionOpenAI sends an image to OpenAI's vision endpoint
func (c *AIClient) ImageOpinionOpenAI(imageURL, systemMessage, model string, maxTokens int, customPrompt *string) (string, error) {
	// Download and encode image
	base64Image, err := downloadAndEncodeImage(imageURL)
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

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.openaiAPIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format from OpenAI")
	}

	return content, nil
}

// ImageOpinionGrok sends an image to Grok API
func (c *AIClient) ImageOpinionGrok(imageURL, systemMessage string, customPrompt *string) (string, error) {
	if c.xaiAPIKey == "" {
		return "", fmt.Errorf("XAI_API_KEY environment variable not set")
	}

	base64Image, err := downloadAndEncodeImage(imageURL)
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

	req, err := http.NewRequest("POST", "https://api.x.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.xaiAPIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Grok API error: %d - %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no response from Grok")
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format from Grok")
	}

	return content, nil
}

// downloadAndEncodeImage downloads an image from URL and returns base64 encoded string
func downloadAndEncodeImage(imageURL string) (string, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(imageData), nil
}
