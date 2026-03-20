// Package scrapedo provides an HTTP client for the Scrape.do API.
package scrapedo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// DefaultBaseURL is the standard API endpoint for Scrape.do.
const DefaultBaseURL = "http://api.scrape.do"

// ErrEmptyToken is returned when no token is provided.
var ErrEmptyToken = errors.New("scrape.do token is required")

// ErrEmptyURL is returned when the target URL is empty.
var ErrEmptyURL = errors.New("target URL cannot be empty")

// ErrAPI represents an error returned by the Scrape.do API.
var ErrAPI = errors.New("scrape.do API error")

// ScrapeRequest holds parameters for the Scrape.do API call.
type ScrapeRequest struct {
	// The target URL to scrape (Required).
	URL string
	// Set to true for JavaScript-heavy websites that need browser rendering.
	Render bool
	// Set to true to use residential and mobile proxies.
	Super bool
}

// Client is a bare-bones HTTP client for the Scrape.do API.
type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

// NewClient initializes a new Scrape.do API client.
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}

	return &Client{
		token:      token,
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{},
	}, nil
}

// Scrape performs a GET request to Scrape.do API with the given parameters
// and specifically requests markdown output.
func (c *Client) Scrape(ctx context.Context, req ScrapeRequest) (string, error) {
	if req.URL == "" {
		return "", ErrEmptyURL
	}

	reqURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}

	// Prepare query parameters
	q := reqURL.Query()
	q.Set("token", c.token)
	q.Set("url", req.URL)
	q.Set("output", "markdown")

	if req.Render {
		q.Set("render", "true")
	}
	if req.Super {
		q.Set("super", "true")
	}

	reqURL.RawQuery = q.Encode()

	// Create HTTP request with context
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	// Execute HTTP request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: status %d: %s", ErrAPI, resp.StatusCode, string(bodyBytes))
	}

	return string(bodyBytes), nil
}
