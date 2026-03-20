// Package scrapedo provides an HTTP client for the Scrape.do API.
package scrapedo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
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
	// 2-letter country code (e.g., "us", "gb", "de") to route requests through a specific location.
	GeoCode string
	// Unique string to maintain a sticky session (same proxy IP).
	Session string
	// Emulate a specific device: "desktop" (default), "mobile", or "tablet".
	Device string
	// HTTP method: "GET" (default), "POST", "PUT", etc.
	Method string
	// Custom HTTP headers to be forwarded.
	Headers map[string]string
	// Data to be sent for POST/PUT requests.
	Body []byte
	// Actions to perform in the browser (for render=true).
	Actions []any
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

// Scrape performs a GET/POST request to Scrape.do API with the given parameters.
func (c *Client) Scrape(ctx context.Context, req ScrapeRequest) (string, error) {
	if req.URL == "" {
		return "", ErrEmptyURL
	}

	reqURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}

	q, err := c.prepareQueryParams(req)
	if err != nil {
		return "", err
	}
	reqURL.RawQuery = q.Encode()

	// For POST/PUT requests or when custom headers are provided, we need to signal Scrape.do.
	method := req.Method
	if method == "" {
		method = http.MethodGet
	}

	// Create HTTP request with context
	var bodyReader io.Reader
	if req.Body != nil {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, reqURL.String(), bodyReader)
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	// Add custom headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Execute HTTP request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	logMetadata(resp.Header)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: status %d: %s", ErrAPI, resp.StatusCode, string(bodyBytes))
	}

	return string(bodyBytes), nil
}

func (c *Client) prepareQueryParams(req ScrapeRequest) (url.Values, error) {
	q := url.Values{}
	q.Set("token", c.token)
	q.Set("url", req.URL)
	q.Set("output", "markdown")

	if req.Render {
		q.Set("render", "true")
	}
	if req.Super {
		q.Set("super", "true")
	}
	if req.GeoCode != "" {
		q.Set("geoCode", req.GeoCode)
	}
	if req.Session != "" {
		q.Set("session", req.Session)
	}
	if req.Device != "" {
		q.Set("device", req.Device)
	}

	if len(req.Actions) > 0 {
		actionsJSON, err := json.Marshal(req.Actions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal browser actions: %w", err)
		}
		q.Set("playWithBrowser", string(actionsJSON))
	}

	if len(req.Headers) > 0 {
		q.Set("customHeaders", "true")
	}

	return q, nil
}

// logMetadata extracts Scrape.do custom headers and logs them to stderr.
func logMetadata(headers http.Header) {
	// Custom headers from Scrape.do:
	// Scrape.do-Remaining-Credits
	// Scrape.do-Initial-Status-Code
	// Scrape.do-Request-Cost
	remaining := headers.Get("Scrape.do-Remaining-Credits")
	targetStatus := headers.Get("Scrape.do-Initial-Status-Code")
	cost := headers.Get("Scrape.do-Request-Cost")

	if remaining != "" || targetStatus != "" || cost != "" {
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
		logger.Info("Scrape.do metadata",
			slog.String("remaining_credits", remaining),
			slog.String("target_status", targetStatus),
			slog.String("cost", cost),
		)
	}
}
