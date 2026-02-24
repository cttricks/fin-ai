package gemini

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

const geminiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"

func New(apiKey, systemPrompt string) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("gemini API key is empty")
	}
	if strings.TrimSpace(systemPrompt) == "" {
		return nil, errors.New("gemini system prompt is empty")
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

	reqBody := geminiRequest{
		SystemInstruction: &geminiSystemInstruction{
			Parts: []geminiPart{{Text: c.systemPrompt}},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: trimmed}},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:     0.1,
			MaxOutputTokens: 256,
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("gemini marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, geminiEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("gemini create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gemini read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini error: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed geminiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("gemini parse response: %w", err)
	}

	text := parsed.OutputText()
	if text == "" {
		return "", errors.New("gemini empty response")
	}

	return strings.TrimSpace(text), nil
}

type geminiRequest struct {
	SystemInstruction *geminiSystemInstruction `json:"system_instruction,omitempty"`
	Contents          []geminiContent          `json:"contents"`
	GenerationConfig  geminiGenerationConfig   `json:"generationConfig"`
}

type geminiSystemInstruction struct {
	Parts []geminiPart `json:"parts"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text,omitempty"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

func (r geminiResponse) OutputText() string {
	for _, candidate := range r.Candidates {
		for _, part := range candidate.Content.Parts {
			if strings.TrimSpace(part.Text) != "" {
				return part.Text
			}
		}
	}
	return ""
}
