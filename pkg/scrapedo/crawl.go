package scrapedo

import (
	"context"
	"fmt"
)

// CrawlOptions configures a breadth-first crawl.
type CrawlOptions struct {
	MaxDepth int
	MaxPages int
	Format   string
}

// CrawlResult holds the output for a single crawled page.
type CrawlResult struct {
	URL     string
	Content string
	Links   []string
	Depth   int
	Size    int
	Error   error
}

type crawlEntry struct {
	url   string
	depth int
}

// Crawl performs a breadth-first crawl starting from the given URL.
// It calls the provided callback for each page scraped.
func (c *Client) Crawl(
	ctx context.Context, startURL string, opts CrawlOptions, callback func(CrawlResult),
) error {
	if startURL == "" {
		return ErrEmptyURL
	}

	if opts.MaxPages <= 0 {
		opts.MaxPages = 10 //nolint:mnd // sensible default
	}

	if opts.MaxDepth < 0 {
		opts.MaxDepth = 1
	}

	visited := make(map[string]struct{})
	queue := []crawlEntry{{url: startURL, depth: 0}}
	pages := 0

	for len(queue) > 0 && pages < opts.MaxPages {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("crawl cancelled: %w", err)
		}

		entry := queue[0]
		queue = queue[1:]

		if _, seen := visited[entry.url]; seen {
			continue
		}
		visited[entry.url] = struct{}{}

		result := c.scrapePage(ctx, entry)
		pages++
		callback(result)

		queue = enqueueLinks(result, entry, opts, visited, queue)
	}

	return nil
}

// scrapePage scrapes a single URL and returns the result.
func (c *Client) scrapePage(ctx context.Context, entry crawlEntry) CrawlResult {
	content, err := c.Scrape(ctx, ScrapeRequest{URL: entry.url})

	result := CrawlResult{
		URL:   entry.url,
		Depth: entry.depth,
		Error: err,
	}

	if err == nil {
		result.Content = content
		result.Size = len(content)
		result.Links = ExtractLinks(content, entry.url)
	}

	return result
}

// enqueueLinks adds new links from a crawl result to the queue.
func enqueueLinks(
	result CrawlResult, entry crawlEntry, opts CrawlOptions,
	visited map[string]struct{}, queue []crawlEntry,
) []crawlEntry {
	if result.Error != nil || entry.depth >= opts.MaxDepth {
		return queue
	}

	for _, link := range result.Links {
		if _, seen := visited[link]; !seen {
			queue = append(queue, crawlEntry{url: link, depth: entry.depth + 1})
		}
	}

	return queue
}
