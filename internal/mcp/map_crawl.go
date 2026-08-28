package mcp

import (
	"context"
	"fmt"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

const (
	defaultMapURLs    = 100
	defaultCrawlDepth = 1
	defaultCrawlPages = 10
	maxMapURLs        = 1000
	maxCrawlDepth     = 5
	maxCrawlPages     = 100
)

// mapToolArgs defines arguments for the map_urls tool.
type mapToolArgs struct {
	URL    string `json:"url"              jsonschema:"The target URL"`
	Search string `json:"search,omitempty" jsonschema:"Filter URLs by keyword"`
	Limit  int    `json:"limit,omitempty"  jsonschema:"Max URLs (default 100, maximum 1000)"`
}

// crawlToolArgs defines arguments for the crawl_site tool.
type crawlToolArgs struct {
	URL      string `json:"url"                jsonschema:"Start URL"`
	MaxDepth int    `json:"maxDepth,omitempty" jsonschema:"Max depth (default 1, maximum 5)"`
	MaxPages int    `json:"maxPages,omitempty" jsonschema:"Max pages (default 10, maximum 100)"`
}

func addMapTool(server *mcpsdk.Server, client *scrapedo.Client, recorder UsageRecorder) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "map_urls",
		Description: "Discover all same-domain URLs on a web page by scraping it and extracting links.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, args mapToolArgs) (*mcpsdk.CallToolResult, any, error) {
		return handleMapTool(ctx, client, args, recorder)
	})
}

func handleMapTool(
	ctx context.Context, client *scrapedo.Client, args mapToolArgs, recorder UsageRecorder,
) (*mcpsdk.CallToolResult, any, error) {
	if args.URL == "" {
		return toolErr("url is required"), nil, nil
	}
	if args.Limit > maxMapURLs {
		return toolErr(fmt.Sprintf("limit must not exceed %d", maxMapURLs)), nil, nil
	}

	content, err := client.Scrape(ctx, scrapedo.ScrapeRequest{URL: args.URL})
	if err != nil {
		return toolErr(fmt.Sprintf("scrape failed: %v", err)), nil, nil
	}

	recordUsage(ctx, recorder, "map", args.URL)

	links := scrapedo.ExtractLinks(content, args.URL)
	links = applyMapFilters(links, args)

	text := fmt.Sprintf("Discovered %d URLs:\n\n%s", len(links), strings.Join(links, "\n"))
	if len(text) > maxMCPOutputBytes {
		return toolErr(outputLimitMessage()), nil, nil
	}

	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: text}},
	}, nil, nil
}

func applyMapFilters(links []string, args mapToolArgs) []string {
	if args.Search != "" {
		needle := strings.ToLower(args.Search)
		filtered := links[:0]

		for _, l := range links {
			if strings.Contains(strings.ToLower(l), needle) {
				filtered = append(filtered, l)
			}
		}

		links = filtered
	}

	limit := args.Limit
	if limit <= 0 {
		limit = defaultMapURLs
	}

	if len(links) > limit {
		links = links[:limit]
	}

	return links
}

func addCrawlTool(server *mcpsdk.Server, client *scrapedo.Client, recorder UsageRecorder) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "crawl_site",
		Description: "Crawl a website breadth-first, scraping multiple pages and returning all content as markdown.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, args crawlToolArgs) (*mcpsdk.CallToolResult, any, error) {
		return handleCrawlTool(ctx, client, args, recorder)
	})
}

func handleCrawlTool(
	ctx context.Context, client *scrapedo.Client, args crawlToolArgs, recorder UsageRecorder,
) (*mcpsdk.CallToolResult, any, error) {
	if args.URL == "" {
		return toolErr("url is required"), nil, nil
	}
	if args.MaxDepth > maxCrawlDepth {
		return toolErr(fmt.Sprintf("maxDepth must not exceed %d", maxCrawlDepth)), nil, nil
	}
	if args.MaxPages > maxCrawlPages {
		return toolErr(fmt.Sprintf("maxPages must not exceed %d", maxCrawlPages)), nil, nil
	}

	opts := buildCrawlOpts(args)
	var buf strings.Builder
	pageNum := 0
	outputTooLarge := false
	crawlCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := client.Crawl(crawlCtx, args.URL, opts, func(r scrapedo.CrawlResult) {
		pageNum++
		if !appendCrawlResult(&buf, r, pageNum) {
			outputTooLarge = true
			cancel()
			return
		}
		recordUsage(ctx, recorder, "crawl", r.URL)
	})
	if outputTooLarge {
		return toolErr(outputLimitMessage()), nil, nil
	}
	if err != nil {
		return toolErr(fmt.Sprintf("crawl failed: %v", err)), nil, nil
	}

	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: buf.String()}},
	}, nil, nil
}

func buildCrawlOpts(args crawlToolArgs) scrapedo.CrawlOptions {
	depth := args.MaxDepth
	if depth <= 0 {
		depth = defaultCrawlDepth
	}

	pages := args.MaxPages
	if pages <= 0 {
		pages = defaultCrawlPages
	}

	return scrapedo.CrawlOptions{MaxDepth: depth, MaxPages: pages}
}

func appendCrawlResult(buf *strings.Builder, r scrapedo.CrawlResult, pageNum int) bool {
	var page strings.Builder
	if r.Error != nil {
		fmt.Fprintf(&page, "## Page %d: %s\n\nError: %v\n\n", pageNum, r.URL, r.Error)
	} else {
		fmt.Fprintf(&page, "## Page %d: %s\n\n%s\n\n---\n\n", pageNum, r.URL, r.Content)
	}

	if buf.Len()+page.Len() > maxMCPOutputBytes {
		return false
	}

	_, _ = buf.WriteString(page.String())
	return true
}

func recordUsage(ctx context.Context, recorder UsageRecorder, action, targetURL string) {
	if recorder != nil {
		//nolint:gosec // best-effort usage tracking
		_ = recorder.RecordUsage(ctx, "scrapedo", "", action, "", targetURL, 1)
	}
}

func toolErr(msg string) *mcpsdk.CallToolResult {
	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: msg}},
		IsError: true,
	}
}
