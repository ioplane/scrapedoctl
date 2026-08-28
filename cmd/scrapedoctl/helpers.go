package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ioplane/scrapedoctl/internal/cache"
	"github.com/ioplane/scrapedoctl/internal/config"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

// buildClient creates a scrapedo.Client from config/env and attaches the cache.
func buildClient(
	c *config.Config, store *cache.Store, extraOptions ...scrapedo.ClientOption,
) (*scrapedo.Client, error) {
	token := c.Global.Token
	if token == "" {
		token = os.Getenv("SCRAPEDO_TOKEN")
	}

	if token == "" {
		return nil, errMissingToken
	}

	options := make([]scrapedo.ClientOption, 0, 3+len(extraOptions))
	if c.Global.BaseURL != "" {
		options = append(options, scrapedo.WithBaseURL(c.Global.BaseURL))
	}
	if c.Global.Timeout > 0 {
		options = append(options, scrapedo.WithTimeout(time.Duration(c.Global.Timeout)*time.Millisecond))
	}
	if store != nil {
		options = append(options, scrapedo.WithCache(store))
	}
	options = append(options, extraOptions...)

	client, err := scrapedo.NewClient(token, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}

// commandContext supports direct RunE unit calls while production execution uses ExecuteContext.
func commandContext(cmd interface{ Context() context.Context }) context.Context {
	if ctx := cmd.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}

// extractHost returns the hostname from a URL string.
func extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	return u.Host
}

// sanitizePath converts a URL path to a safe filename.
var unsafePathChars = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

func sanitizePath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "page"
	}

	p := strings.Trim(u.Path, "/")
	if p == "" {
		return "index"
	}

	return unsafePathChars.ReplaceAllString(p, "_")
}
