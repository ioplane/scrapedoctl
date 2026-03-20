// Package mcp provides the Model Context Protocol server implementation for scrapedoctl.
package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

// toolArgs defines the expected arguments from Claude for the scrape_url tool.
type toolArgs struct {
	URL    string `json:"url"              jsonschema:"description=The target URL to scrape,required"`
	Render bool   `json:"render,omitempty" jsonschema:"description=Set to true to execute JavaScript"`
	Super  bool   `json:"super,omitempty"  jsonschema:"description=Set to true to utilize residential proxy"`
}

// RunServer initializes and runs the standard stdio MCP server for Scrape.do.
func RunServer(ctx context.Context, apiToken string) error {
	client, err := scrapedo.NewClient(apiToken)
	if err != nil {
		return fmt.Errorf("failed to init scrape.do client: %w", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "scrapedoctl",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "scrape_url",
		Description: "Scrape a web page and return optimized markdown. Supports JS rendering and proxy rotation.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args toolArgs) (*mcp.CallToolResult, any, error) {
		scrapeReq := scrapedo.ScrapeRequest{
			URL:    args.URL,
			Render: args.Render,
			Super:  args.Super,
		}

		result, err := client.Scrape(ctx, scrapeReq)
		if err != nil {
			// Instead of returning a raw Go error which might crash the tool,
			// we return it as text content to let Claude see the error.
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Scrape failed: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, nil, nil
	})

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return fmt.Errorf("mcp server failed: %w", err)
	}

	return nil
}
