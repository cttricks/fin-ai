package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"fin/internal/ai"
	"fin/internal/config"
	"fin/internal/server"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var showVersion bool

	root := &cobra.Command{
		Use:           "fin",
		Short:         "Local AI-powered search intent optimizer",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if showVersion {
				fmt.Fprintln(cmd.OutOrStdout(), version)
				return nil
			}
			return cmd.Help()
		},
	}

	root.SetVersionTemplate("fin version " + version + "\n")
	root.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version")

	root.AddCommand(newVersionCmd())
	root.AddCommand(newRunCmd())
	root.AddCommand(newTestCmd())
	root.AddCommand(newProviderCmd(config.ProviderOpenAI))
	root.AddCommand(newProviderCmd(config.ProviderGemini))

	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version)
		},
	}
}

func newRunCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run local HTTP server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			provider, _, err := loadDefaultProvider()
			if err != nil {
				return err
			}

			srv := server.New(port, provider)
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			return srv.Start(ctx)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 2026, "Port to listen on")

	return cmd
}

type testResult struct {
	Provider string `json:"provider"`
	Input    string `json:"input"`
	Raw      string `json:"raw,omitempty"`
	Site     string `json:"site,omitempty"`
	Query    string `json:"query,omitempty"`
	Error    string `json:"error,omitempty"`
}

func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test [query]",
		Short: "Test AI provider response",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider, providerName, err := loadDefaultProvider()
			if err != nil {
				return err
			}

			input := args[0]
			output, callErr := provider.OptimizeQuery(cmd.Context(), input)

			result := testResult{
				Provider: providerName,
				Input:    input,
				Raw:      output,
			}
			if callErr != nil {
				result.Error = callErr.Error()
			}

			if output != "" {
				if parsed, err := ai.ParseRouterResponse(output); err == nil {
					result.Site = parsed.Site
					result.Query = parsed.Query
				}
			}

			if err := writeJSON(cmd, result); err != nil {
				return err
			}

			return callErr
		},
	}

	return cmd
}

func newProviderCmd(provider string) *cobra.Command {
	var apiKey string

	cmd := &cobra.Command{
		Use:   provider,
		Short: fmt.Sprintf("Configure %s API key", provider),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if apiKey == "" {
				return errors.New("missing API key; use -api \"<key>\"")
			}

			cfgMgr, err := config.NewManager()
			if err != nil {
				return err
			}

			cfg, err := cfgMgr.Load()
			if err != nil {
				return err
			}

			cfg.SetAPIKey(provider, apiKey)
			cfg.UpdatedAt = time.Now().UTC()

			if err := cfgMgr.Save(cfg); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Saved API key.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&apiKey, "api", "a", "", "API key")

	return cmd
}

func loadDefaultProvider() (ai.AIProvider, string, error) {
	cfgMgr, err := config.NewManager()
	if err != nil {
		return nil, "", err
	}

	cfg, err := cfgMgr.Load()
	if err != nil {
		return nil, "", err
	}

	if cfg.DefaultProvider == "" {
		return nil, "", errors.New("no default AI provider configured; run fin openai|gemini -api \"<key>\" first")
	}

	key := cfg.APIKeys[cfg.DefaultProvider]
	if key == "" {
		return nil, "", errors.New("no API key configured for provider; run fin openai|gemini -api \"<key>\" first")
	}

	provider, err := ai.NewProvider(cfg.DefaultProvider, key)
	if err != nil {
		return nil, "", err
	}
	return provider, cfg.DefaultProvider, nil
}

func writeJSON(cmd *cobra.Command, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(payload))
	return err
}
