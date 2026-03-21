package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
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

			// TODO: Implement actual configuration logic in internal/install
			fmt.Printf("Configuring token: %s\n", token[:4]+"...")
			fmt.Printf("Configuring agents: %v\n", agents)

			return nil
		},
	}
}
