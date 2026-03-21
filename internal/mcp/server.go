// Package mcp provides the Model Context Protocol server implementation for scrapedoctl.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ioplane/scrapedoctl/internal/version"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

// toolArgs defines the expected arguments from Claude for the scrape_url tool.
type toolArgs struct {
	URL     string            `json:"url"               jsonschema:"description=The target URL to scrape,required"`
	Render  bool              `json:"render,omitempty"  jsonschema:"description=Set to true to execute JavaScript"`
	Super   bool              `json:"super,omitempty"   jsonschema:"description=Set to true to utilize residential proxy"`
	GeoCode string            `json:"geoCode,omitempty" jsonschema:"description=2-letter country code (e.g. us, gb, de) to route requests through a specific location"`
	Session string            `json:"session,omitempty" jsonschema:"description=Unique string to maintain a sticky session (same proxy IP)"`
	Device  string            `json:"device,omitempty"  jsonschema:"description=Emulate a specific device,enum=desktop,enum=mobile,enum=tablet"`
	Method  string            `json:"method,omitempty"  jsonschema:"description=HTTP method to use (default: GET)"`
	Headers map[string]string `json:"headers,omitempty" jsonschema:"description=Custom HTTP headers to forward to the target"`
	Body    string            `json:"body,omitempty"    jsonschema:"description=Request body for POST/PUT requests"`
	Actions []any             `json:"actions,omitempty" jsonschema:"description=Browser actions like click, scroll, etc. (requires render=true)"`
}

// RunServer initializes and runs the standard stdio MCP server for Scrape.do.
func RunServer(ctx context.Context, apiToken string) error {
	client, err := scrapedo.NewClient(apiToken)
	if err != nil {
		return fmt.Errorf("failed to init scrape.do client: %w", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "scrapedoctl",
		Version: version.Version,
	}, nil)

	// Add Resource: CLI Documentation
	server.AddResource(&mcp.Resource{
		URI:         "resource://cli/help",
		Name:        "CLI Documentation",
		Description: "Detailed documentation of the scrapedoctl CLI commands and flags in JSON format.",
		MIMEType:    "application/json",
	}, func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		doc := map[string]any{
			"cli":         "scrapedoctl",
			"description": "Scrape.do CLI & MCP Server",
			"usage":       "scrapedoctl [command] [flags]",
			"commands": []map[string]any{
				{
					"name":        "scrape",
					"description": "Scrape a single URL to markdown",
					"flags":       []string{"--render", "--super", "--geoCode", "--session", "--device"},
				},
				{
					"name":        "repl",
					"description": "Start an interactive shell",
				},
				{
					"name":        "mcp",
					"description": "Run as an MCP stdio server",
				},
			},
		}
		data, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal documentation: %w", err)
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      "resource://cli/help",
					MIMEType: "application/json",
					Text:     string(data),
				},
			},
		}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "scrape_url",
		Description: "Scrape a web page and return optimized markdown. Supports JS rendering, proxy rotation, geo-targeting, sticky sessions, device emulation, and POST requests with custom headers and browser actions.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args toolArgs) (*mcp.CallToolResult, any, error) {
		scrapeReq := scrapedo.ScrapeRequest{
			URL:     args.URL,
			Render:  args.Render,
			Super:   args.Super,
			GeoCode: args.GeoCode,
			Session: args.Session,
			Device:  args.Device,
			Method:  args.Method,
			Headers: args.Headers,
			Body:    []byte(args.Body),
			Actions: args.Actions,
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
