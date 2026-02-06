package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"vex-backend/config"
)

type openAiChatter struct {
	model string
}

func newOpenAIChatter() chatter {
	return &openAiChatter{
		model: "gpt-4o",
	}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func (oac openAiChatter) GetResponse(ctx context.Context, query string) (string, error) {
	if query == "" {
		return "", errors.New("query cannot be empty")
	}

	reqBody := ChatCompletionRequest{
		Model: oac.model,
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: query,
			},
		},
	}

	return oac.makeRequest(ctx, reqBody)
}

func (oac openAiChatter) GetResponseWithSystemPrompt(ctx context.Context, query string, systemprompt string) (string, error) {
	if query == "" {
		return "", errors.New("query cannot be empty")
	}
	if systemprompt == "" {
		return "", errors.New("system prompt cannot be empty")
	}

	reqBody := ChatCompletionRequest{
		Model: oac.model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: systemprompt,
			},
			{
				Role:    "user",
				Content: query,
			},
		},
	}

	return oac.makeRequest(ctx, reqBody)
}

// makeRequest is a helper function to make the HTTP request
func (oac openAiChatter) makeRequest(ctx context.Context, reqBody ChatCompletionRequest) (string, error) {
	// Marshal request to JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Config.OpenAiAPIKey))

	httpClient := http.Client{}
	// Make the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var completion ChatCompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if completion.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s (type: %s, code: %s)",
			completion.Error.Message,
			completion.Error.Type,
			completion.Error.Code)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Check if we got a response
	if len(completion.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}

	return completion.Choices[0].Message.Content, nil
}
