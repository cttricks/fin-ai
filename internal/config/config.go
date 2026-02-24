package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const (
	ProviderOpenAI = "openai"
	ProviderGemini = "gemini"
)

type Config struct {
	DefaultProvider string            `json:"default_provider"`
	APIKeys         map[string]string `json:"api_keys"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

func (c *Config) SetAPIKey(provider, key string) {
	if c.APIKeys == nil {
		c.APIKeys = make(map[string]string)
	}
	c.APIKeys[provider] = key
	c.DefaultProvider = provider
}

type Manager struct {
	path string
}

func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".fin", "config.json")
	return &Manager{path: path}, nil
}

func (m *Manager) Load() (*Config, error) {
	data, err := os.ReadFile(m.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{APIKeys: map[string]string{}}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.APIKeys == nil {
		cfg.APIKeys = map[string]string{}
	}
	return &cfg, nil
}

func (m *Manager) Save(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}

	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.path, payload, 0o600)
}
