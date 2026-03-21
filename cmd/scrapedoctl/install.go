package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/ioplane/scrapedoctl/internal/install"
)

func newInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install and configure scrapedoctl",
		RunE: func(_ *cobra.Command, _ []string) error {
			var token string
			var agents []string

			// Start interactive setup
			err := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Scrape.do API Token").
						Description("Enter your API token (find it at https://scrape.do/dashboard)").
						Value(&token).
						Password(true),
				),
				huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("Configure AI Agents").
						Description("Select agents to configure for Scrape.do MCP server (Space to select)").
						Options(
							huh.NewOption("Claude Code", "claude"),
							huh.NewOption("JetBrains Junie", "junie"),
							huh.NewOption("Gemini CLI", "gemini"),
							huh.NewOption("Codex AI", "codex"),
							huh.NewOption("Kimi AI", "kimi"),
							huh.NewOption("OpenCode AI", "opencode"),
						).
						Value(&agents),
				),
			).Run()

			if err != nil {
				return fmt.Errorf("installation cancelled: %w", err)
			}

			// Ensure log directory exists
			logDir := "/var/log/scrapedoctl"
			if err := os.MkdirAll(logDir, 0755); err != nil {
				fmt.Printf("\nWarning: Could not create log directory %s: %v\n", logDir, err)
				fmt.Printf("To fix this, please run: sudo mkdir -p %s && sudo chown $USER %s\n", logDir, logDir)
			} else {
				fmt.Printf("Log directory created: %s\n", logDir)
			}

			// Execute configuration logic
			if err := install.ConfigureAgents(agents, token); err != nil {
				return fmt.Errorf("failed to configure agents: %w", err)
			}

			// Save token to local config
			cfg.Global.Token = token
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			fmt.Println("\nInstallation complete! You can now use Scrape.do tools in your selected AI agents.")
			return nil
		},
	}
}
