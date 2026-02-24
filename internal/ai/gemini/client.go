package gemini

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func New(apiKey string) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("gemini API key is empty")
	}
	return &Client{
		apiKey: apiKey,
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

	// TODO: Implement Gemini API call. Placeholder returns input for now.
	return trimmed, nil
}
