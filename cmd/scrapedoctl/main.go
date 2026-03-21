package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ioplane/scrapedoctl/internal/config"
	"github.com/ioplane/scrapedoctl/internal/logger"
	"github.com/ioplane/scrapedoctl/internal/mcp"
	"github.com/ioplane/scrapedoctl/internal/repl"
	"github.com/ioplane/scrapedoctl/internal/ui"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

var (
	// cfg is the loaded application configuration.
	cfg *config.Config
	// configPath is the path to the configuration file.
	configPath string
	// profileName is the name of the profile to use.
	profileName string
)

// errMissingToken is returned when the SCRAPEDO_TOKEN environment variable is missing.
var errMissingToken = errors.New("SCRAPEDO_TOKEN environment variable is required (or set it in config file)")

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scrapedoctl",
		Short: "scrapedoctl is a CLI and MCP server for Scrape.do",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			cfg, err = config.Load(configPath, profileName)
			if err != nil {
				// If config file is missing, we check if we should trigger install
				if errors.Is(err, config.ErrConfigNotFound) {
					if cmd.Name() != "help" && cmd.Name() != "metadata" && cmd.Name() != "install" && cmd.Name() != "completion" {
						fmt.Println("No configuration file found. Starting initial setup...")
						// We need to pass an empty config so we don't panic in commands
						cfg = &config.Config{}
						logger.Init(cfg.Logging)

						// Find the install command and execute it
						installCmd, _, err := cmd.Root().Find([]string{"install"})
						if err == nil && installCmd != nil && installCmd.RunE != nil {
							return installCmd.RunE(installCmd, nil)
						}
						return fmt.Errorf("failed to find or execute install command")
					}
					// For help/metadata, we ignore missing config and use empty/defaults
					cfg = &config.Config{}
					logger.Init(cfg.Logging)
					return nil
				}
				// Otherwise, just fail if it was a different error
				if cmd.Name() != "install" {
					return fmt.Errorf("failed to load config: %w", err)
				}
			}

			if cfg != nil {
				logger.Init(cfg.Logging)
			}
			return nil
		},
	}

	ui.SetCustomHelp(cmd)

	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", config.DefaultConfigPath, "path to config file")
	cmd.PersistentFlags().StringVarP(&profileName, "profile", "p", "", "use a specific configuration profile")

	cmd.AddCommand(newMCPCmd())
	cmd.AddCommand(newREPLCmd())
	cmd.AddCommand(newScrapeCmd())
	cmd.AddCommand(newMetadataCmd())
	cmd.AddCommand(newInstallCmd())
	cmd.AddCommand(newConfigCmd())

	return cmd
}

func findCmdIndex(root *cobra.Command, name string) int {
	for i, c := range root.Commands() {
		if c.Name() == name {
			return i
		}
	}
	return 0
}

func newScrapeCmd() *cobra.Command {
	var render, super bool
	cmd := &cobra.Command{
		Use:   "scrape <url>",
		Short: "Scrape a single URL and output markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token := cfg.Global.Token
			if token == "" {
				token = os.Getenv("SCRAPEDO_TOKEN")
			}
			if token == "" {
				return errMissingToken
			}

			client, err := scrapedo.NewClient(token)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			req := scrapedo.ScrapeRequest{
				URL:     args[0],
				Render:  cfg.Resolved.Render,
				Super:   cfg.Resolved.Super,
				GeoCode: cfg.Resolved.GeoCode,
				Session: cfg.Resolved.Session,
				Device:  cfg.Resolved.Device,
			}

			// CLI flags override config/profile
			if cmd.Flags().Changed("render") {
				req.Render = render
			}
			if cmd.Flags().Changed("super") {
				req.Super = super
			}

			result, err := client.Scrape(context.Background(), req)
			if err != nil {
				return fmt.Errorf("scrape failed: %w", err)
			}

			fmt.Println(result)
			return nil
		},
	}

	cmd.Flags().BoolVar(&render, "render", false, "Execute JavaScript")
	cmd.Flags().BoolVar(&super, "super", false, "Use residential proxy")

	return cmd
}

func newREPLCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "repl",
		Short: "Start an interactive Scrape.do shell",
		RunE: func(_ *cobra.Command, _ []string) error {
			token := cfg.Global.Token
			if token == "" {
				token = os.Getenv("SCRAPEDO_TOKEN")
			}
			if token == "" {
				return errMissingToken
			}

			client, err := scrapedo.NewClient(token)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			shell := repl.NewShell(client)
			return shell.Run(context.Background())
		},
	}
}

func newMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP server over stdio",
		RunE: func(_ *cobra.Command, _ []string) error {
			token := cfg.Global.Token
			if token == "" {
				token = os.Getenv("SCRAPEDO_TOKEN")
			}
			if token == "" {
				return errMissingToken
			}

			// We use context.Background() since the MCP server handles its own lifecycle
			// over stdio and will exit when the stream closes.
			return mcp.RunServer(context.Background(), token)
		},
	}
}
