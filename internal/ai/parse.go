package ai

import (
	"encoding/json"
	"errors"
	"strings"
)

type RouterResponse struct {
	Site  string `json:"site"`
	Query string `json:"query"`
}

func ParseRouterResponse(text string) (RouterResponse, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return RouterResponse{}, errors.New("empty response")
	}

	var parsed RouterResponse
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return RouterResponse{}, err
	}

	parsed.Site = strings.TrimSpace(parsed.Site)
	parsed.Query = strings.TrimSpace(parsed.Query)
	return parsed, nil
}
