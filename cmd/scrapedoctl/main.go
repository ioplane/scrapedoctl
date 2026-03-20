// Package main provides the entrypoint for the scrapedoctl CLI.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ioplane/scrapedoctl/internal/mcp"
	"github.com/ioplane/scrapedoctl/internal/repl"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

// errMissingToken is returned when the SCRAPEDO_TOKEN environment variable is missing.
var errMissingToken = errors.New("SCRAPEDO_TOKEN environment variable is required")

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
	}

	cmd.AddCommand(newMCPCmd())
	cmd.AddCommand(newREPLCmd())
	cmd.AddCommand(newScrapeCmd())

	return cmd
}

func newScrapeCmd() *cobra.Command {
	var render, super bool
	cmd := &cobra.Command{
		Use:   "scrape <url>",
		Short: "Scrape a single URL and output markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			token := os.Getenv("SCRAPEDO_TOKEN")
			if token == "" {
				return errMissingToken
			}

			client, err := scrapedo.NewClient(token)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			result, err := client.Scrape(context.Background(), scrapedo.ScrapeRequest{
				URL:    args[0],
				Render: render,
				Super:  super,
			})
			if err != nil {
				return err
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
			token := os.Getenv("SCRAPEDO_TOKEN")
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
			token := os.Getenv("SCRAPEDO_TOKEN")
			if token == "" {
				return errMissingToken
			}

			// We use context.Background() since the MCP server handles its own lifecycle
			// over stdio and will exit when the stream closes.
			return mcp.RunServer(context.Background(), token)
		},
	}
}
