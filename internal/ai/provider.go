package ai

import (
	"context"
	"errors"

	"fin/internal/ai/gemini"
	"fin/internal/ai/openai"
)

const (
	ProviderOpenAI = "openai"
	ProviderGemini = "gemini"
)

type AIProvider interface {
	OptimizeQuery(ctx context.Context, input string) (string, error)
}

func NewProvider(provider, apiKey string) (AIProvider, error) {
	switch provider {
	case ProviderOpenAI:
		return openai.New(apiKey, SystemPrompt())
	case ProviderGemini:
		return gemini.New(apiKey, SystemPrompt())
	default:
		return nil, errors.New("unknown AI provider")
	}
}
