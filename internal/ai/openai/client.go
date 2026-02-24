package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	apiKey       string
	systemPrompt string
	httpClient   *http.Client
}

const openAIEndpoint = "https://api.openai.com/v1/responses"
const openAIModel = "gpt-4o"

func New(apiKey, systemPrompt string) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("openai API key is empty")
	}
	if strings.TrimSpace(systemPrompt) == "" {
		return nil, errors.New("openai system prompt is empty")
	}
	return &Client{
		apiKey:       apiKey,
		systemPrompt: systemPrompt,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (c *Client) OptimizeQuery(ctx context.Context, input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", errors.New("input is empty")
	}

	reqBody := openAIRequest{
		Model:           openAIModel,
		Input:           trimmed,
		Instructions:    c.systemPrompt,
		Temperature:     0.1,
		MaxOutputTokens: 256,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("openai marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("openai create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("openai read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai error: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed openAIResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("openai parse response: %w", err)
	}

	text := parsed.OutputText()
	if text == "" {
		return "", errors.New("openai empty response")
	}

	return strings.TrimSpace(text), nil
}

type openAIRequest struct {
	Model           string  `json:"model"`
	Input           string  `json:"input"`
	Instructions    string  `json:"instructions,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"max_output_tokens,omitempty"`
}

type openAIResponse struct {
	Output []openAIOutputItem `json:"output"`
	Error  *openAIError       `json:"error,omitempty"`
}

type openAIOutputItem struct {
	Type    string             `json:"type"`
	Role    string             `json:"role"`
	Content []openAIOutContent `json:"content"`
}

type openAIOutContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type openAIError struct {
	Message string `json:"message"`
}

func (r openAIResponse) OutputText() string {
	if r.Error != nil && strings.TrimSpace(r.Error.Message) != "" {
		return ""
	}
	for _, item := range r.Output {
		for _, part := range item.Content {
			if part.Type == "output_text" || part.Type == "text" {
				if strings.TrimSpace(part.Text) != "" {
					return part.Text
				}
			}
		}
	}
	return ""
}
