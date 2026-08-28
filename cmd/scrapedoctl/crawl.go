package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

var errUnsupportedCrawlFormat = errors.New("unsupported crawl format")

type crawlFlags struct {
	depth  int
	limit  int
	output string
	format string
}

func newCrawlCmd() *cobra.Command {
	cf := &crawlFlags{}
	cmd := &cobra.Command{
		Use:   "crawl <url>",
		Short: "Recursively crawl a site and save content",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error { return runCrawl(cmd, args, cf) },
	}

	cmd.Flags().IntVar(&cf.depth, "depth", 1, "Max crawl depth")
	//nolint:mnd // default crawl limit.
	cmd.Flags().IntVar(&cf.limit, "limit", 10, "Max pages to crawl")
	cmd.Flags().StringVar(&cf.output, "output", "./crawl-output", "Output directory")
	cmd.Flags().StringVar(&cf.format, "format", "markdown", "Output format: markdown, json")

	return cmd
}

func runCrawl(cmd *cobra.Command, args []string, cf *crawlFlags) error {
	if cf.format != "markdown" && cf.format != "json" {
		return fmt.Errorf("%w: %q", errUnsupportedCrawlFormat, cf.format)
	}

	client, err := buildClient(cfg, cacheStore)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(cf.output, 0o750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	pageNum := 0
	opts := scrapedo.CrawlOptions{
		MaxDepth: cf.depth,
		MaxPages: cf.limit,
	}
	crawlCtx, cancel := context.WithCancel(commandContext(cmd))
	defer cancel()
	var saveErr error

	err = client.Crawl(
		crawlCtx, args[0], opts,
		func(r scrapedo.CrawlResult) {
			pageNum++
			if saveErr = handleCrawlResult(cmd, r, cf, pageNum, opts.MaxPages); saveErr != nil {
				cancel()
			}
		},
	)
	if saveErr != nil {
		return saveErr
	}
	if err != nil {
		return fmt.Errorf("crawl failed: %w", err)
	}

	return nil
}

func handleCrawlResult(
	cmd *cobra.Command, r scrapedo.CrawlResult, cf *crawlFlags, pageNum, maxPages int,
) error {
	if r.Error != nil {
		cmd.PrintErrf("[%d/%d] %s → ERROR: %v\n", pageNum, maxPages, r.URL, r.Error)
		return fmt.Errorf("crawl page %q: %w", r.URL, r.Error)
	}

	cmd.PrintErrf("[%d/%d] %s → %s\n", pageNum, maxPages, r.URL, formatSize(r.Size))
	if err := saveCrawlPage(r, cf); err != nil {
		return err
	}
	recordCrawlUsage(cmd, r.URL)
	return nil
}

func saveCrawlPage(r scrapedo.CrawlResult, cf *crawlFlags) error {
	extension := ".md"
	content := []byte(r.Content)
	if cf.format == "json" {
		extension = ".json"
		var err error
		content, err = json.MarshalIndent(struct {
			URL     string   `json:"url"`
			Content string   `json:"content"`
			Links   []string `json:"links"`
			Depth   int      `json:"depth"`
			Size    int      `json:"size"`
		}{r.URL, r.Content, r.Links, r.Depth, r.Size}, "", "  ")
		if err != nil {
			return fmt.Errorf("encode crawl page: %w", err)
		}
	}

	filename := sanitizePath(r.URL) + extension
	path := filepath.Join(cf.output, filename)

	//nolint:gosec // output directory is user-specified
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("save crawl page %q: %w", path, err)
	}
	return nil
}

func recordCrawlUsage(cmd *cobra.Command, targetURL string) {
	if cacheStore != nil {
		//nolint:gosec // best-effort usage tracking
		_ = cacheStore.RecordUsage(
			commandContext(cmd), "scrapedo", "", "crawl", "", targetURL, 1,
		)
	}
}

func formatSize(bytes int) string {
	const kb = 1024

	if bytes < kb {
		return fmt.Sprintf("%dB", bytes)
	}

	return fmt.Sprintf("%dKB", bytes/kb)
}
