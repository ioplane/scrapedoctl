// Package main provides the entrypoint for the scrapedoctl CLI.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ioplane/scrapedoctl/internal/mcp"
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

	return cmd
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
