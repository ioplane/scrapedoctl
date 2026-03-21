// Package search provides a pluggable multi-provider web search system.
package search

import "context"

// Result represents a single search result normalized across providers.
type Result struct {
	Position     int    `json:"position"`
	Title        string `json:"title"`
	URL          string `json:"url"`
	Snippet      string `json:"snippet"`
	DisplayedURL string `json:"displayed_url,omitempty"`
}

// Response is the unified response from any search provider.
type Response struct {
	Query    string         `json:"query"`
	Engine   string         `json:"engine"`
	Provider string         `json:"provider"`
	Results  []Result       `json:"results"`
	Raw      any            `json:"raw,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Options configures a search request.
type Options struct {
	Engine  string
	Lang    string
	Country string
	Limit   int
	Page    int
	Raw     bool
}

// Provider is the interface that all search providers must implement.
type Provider interface {
	// Name returns the provider's unique identifier.
	Name() string
	// Engines returns the list of search engines this provider supports.
	Engines() []string
	// Search performs a web search and returns normalized results.
	Search(ctx context.Context, query string, opts Options) (*Response, error)
}
