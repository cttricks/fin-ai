package main

import (
	"context"
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
			cfgMgr, err := config.NewManager()
			if err != nil {
				return err
			}

			cfg, err := cfgMgr.Load()
			if err != nil {
				return err
			}

			if cfg.DefaultProvider == "" {
				return errors.New("no default AI provider configured; run fin openai|gemini -api \"<key>\" first")
			}

			key := cfg.APIKeys[cfg.DefaultProvider]
			if key == "" {
				return errors.New("no API key configured for provider; run fin openai|gemini -api \"<key>\" first")
			}

			provider, err := ai.NewProvider(cfg.DefaultProvider, key)
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
