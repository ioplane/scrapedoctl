package scrapedo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the standard API endpoint for Scrape.do.
const DefaultBaseURL = "https://api.scrape.do"

const (
	defaultTimeout          = 60 * time.Second
	defaultMaxResponseBytes = int64(32 << 20)
)

// ErrEmptyToken is returned when no token is provided.
var ErrEmptyToken = errors.New("scrape.do token is required")

// ErrEmptyURL is returned when the target URL is missing.
var ErrEmptyURL = errors.New("target URL is required")

// ErrAPI is a generic error wrapper for API-level failures.
var ErrAPI = errors.New("scrape.do API error")

// Cacher defines the interface for persistent caching.
type Cacher interface {
	// GetResult checks the cache for a matching request.
	GetResult(ctx context.Context, req ScrapeRequest) (string, bool, error)
	// SaveResult stores a new scrape result and performs cleanup.
	SaveResult(ctx context.Context, req ScrapeRequest, content string, metadata map[string]any) error
}

// ScrapeRequest holds parameters for the Scrape.do API call.
type ScrapeRequest struct {
	// URL is the target URL to scrape (Required).
	URL string
	// Render set to true for JavaScript-heavy websites that need browser rendering.
	Render bool
	// Super set to true to use residential and mobile proxies.
	Super bool
	// GeoCode is a 2-letter country code (e.g., "us", "gb", "de") to route requests through a specific location.
	GeoCode string
	// Session is a unique string to maintain a sticky session (same proxy IP).
	Session string
	// Device emulates a specific device: "desktop" (default), "mobile", or "tablet".
	Device string
	// Method is the HTTP method: "GET" (default), "POST", "PUT", etc.
	Method string
	// Headers are custom HTTP headers to be forwarded.
	Headers map[string]string
	// Body is the data to be sent for POST/PUT requests.
	Body []byte
	// Actions are the browser actions to perform (for render=true).
	Actions []any

	// NoCache bypasses the local SQLite cache and forces a new API call without saving.
	NoCache bool
	// Refresh forces a new API call and stores the result as a new version in history.
	Refresh bool
}

// Client is a bare-bones HTTP client for the Scrape.do API.
type Client struct {
	token            string
	baseURL          string
	baseURLErr       error
	httpClient       *http.Client
	cache            Cacher
	maxResponseBytes int64
}

type clientOptions struct {
	baseURL               string
	httpClient            *http.Client
	timeout               time.Duration
	maxResponseBytes      int64
	cache                 Cacher
	allowInsecureLoopback bool
}

// ClientOption configures a Client before it becomes visible to callers.
type ClientOption func(*clientOptions) error

// WithBaseURL configures the Scrape.do API endpoint.
func WithBaseURL(rawURL string) ClientOption {
	return func(options *clientOptions) error {
		options.baseURL = rawURL
		return nil
	}
}

// WithHTTPClient configures the HTTP client while preserving secure redirect policy.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(options *clientOptions) error {
		if client == nil {
			return ErrNilHTTPClient
		}
		options.httpClient = client
		return nil
	}
}

// WithTimeout configures the overall HTTP request timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(options *clientOptions) error {
		options.timeout = timeout
		return nil
	}
}

// WithMaxResponseBytes configures the maximum upstream response body size.
func WithMaxResponseBytes(limit int64) ClientOption {
	return func(options *clientOptions) error {
		options.maxResponseBytes = limit
		return nil
	}
}

// WithCache configures the optional cache.
func WithCache(cache Cacher) ClientOption {
	return func(options *clientOptions) error {
		options.cache = cache
		return nil
	}
}

// WithInsecureLoopback permits an HTTP endpoint only when it resolves to loopback.
func WithInsecureLoopback() ClientOption {
	return func(options *clientOptions) error {
		options.allowInsecureLoopback = true
		return nil
	}
}

// NewClient creates a new Scrape.do client with the provided token and immutable options.
func NewClient(token string, optionList ...ClientOption) (*Client, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}

	options := clientOptions{
		baseURL:          DefaultBaseURL,
		httpClient:       &http.Client{},
		timeout:          defaultTimeout,
		maxResponseBytes: defaultMaxResponseBytes,
	}
	for _, option := range optionList {
		if option == nil {
			continue
		}
		if err := option(&options); err != nil {
			return nil, fmt.Errorf("configure Scrape.do client: %w", err)
		}
	}

	baseURL, err := validateBaseURL(options.baseURL, options.allowInsecureLoopback)
	if err != nil {
		return nil, err
	}
	if options.timeout <= 0 {
		return nil, ErrInvalidTimeout
	}
	if options.maxResponseBytes <= 0 {
		return nil, ErrInvalidResponseLimit
	}

	httpClient := *options.httpClient
	httpClient.Timeout = options.timeout
	httpClient.CheckRedirect = secureRedirectPolicy(httpClient.CheckRedirect)

	return &Client{
		token:            token,
		baseURL:          baseURL.String(),
		httpClient:       &httpClient,
		cache:            options.cache,
		maxResponseBytes: options.maxResponseBytes,
	}, nil
}

// SetBaseURL overrides the default API endpoint (useful for testing).
func (c *Client) SetBaseURL(u string) {
	baseURL, err := validateBaseURL(u, true)
	c.baseURLErr = err
	if err == nil {
		c.baseURL = baseURL.String()
	}
}

// SetCache sets the optional caching layer for the client.
func (c *Client) SetCache(cache Cacher) {
	c.cache = cache
}

// Scrape performs a GET/POST request to Scrape.do API with the given parameters.
func (c *Client) Scrape(ctx context.Context, req ScrapeRequest) (string, error) {
	if req.URL == "" {
		return "", ErrEmptyURL
	}
	if c.baseURLErr != nil {
		return "", c.baseURLErr
	}

	// 1. Check Cache
	if content, found := c.checkCache(ctx, req); found {
		return content, nil
	}

	// 2. Prepare Request
	httpReq, err := c.prepareHTTPRequest(ctx, req)
	if err != nil {
		return "", err
	}

	// 3. Execute Request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	c.logMetadata(resp.Header)

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, c.maxResponseBytes+1))
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	if int64(len(bodyBytes)) > c.maxResponseBytes {
		return "", fmt.Errorf("%w: limit %d bytes", ErrResponseTooLarge, c.maxResponseBytes)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", newAPIError(resp, bodyBytes, c.token)
	}

	// 4. Save to Cache
	c.saveToCache(ctx, req, resp, bodyBytes)

	return string(bodyBytes), nil
}

func (c *Client) checkCache(ctx context.Context, req ScrapeRequest) (string, bool) {
	if c.cache != nil && !req.NoCache && !req.Refresh {
		if content, found, err := c.cache.GetResult(ctx, req); err == nil && found {
			slog.Info("Cache hit", slog.String("url", req.URL))
			return content, true
		}
	}
	return "", false
}

func (c *Client) prepareHTTPRequest(ctx context.Context, req ScrapeRequest) (*http.Request, error) {
	reqURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	q, err := c.prepareQueryParams(req)
	if err != nil {
		return nil, err
	}
	reqURL.RawQuery = q.Encode()

	method := req.Method
	if method == "" {
		method = http.MethodGet
	}

	var bodyReader io.Reader
	if req.Body != nil {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, reqURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	if slog.Default().Enabled(ctx, slog.LevelDebug) {
		slog.Debug("Sending request to Scrape.do",
			slog.String("method", method),
			slog.String("url", redactURL(reqURL)),
			slog.Any("headers", loggedHeaders(httpReq.Header)),
		)
	}

	return httpReq, nil
}

func (c *Client) saveToCache(ctx context.Context, req ScrapeRequest, resp *http.Response, bodyBytes []byte) {
	if c.cache != nil && !req.NoCache {
		metadata := map[string]any{
			"status":            resp.StatusCode,
			"remaining_credits": resp.Header.Get("Scrape.do-Remaining-Credits"),
			"cost":              resp.Header.Get("Scrape.do-Request-Cost"),
		}
		if err := c.cache.SaveResult(ctx, req, string(bodyBytes), metadata); err != nil {
			slog.Warn("Failed to save result to cache", slog.Any("error", err))
		}
	}
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
func (c *Client) logMetadata(headers http.Header) {
	// Custom headers from Scrape.do:
	remaining := headers.Get("Scrape.do-Remaining-Credits")
	targetStatus := headers.Get("Scrape.do-Initial-Status-Code")
	cost := headers.Get("Scrape.do-Request-Cost")

	if remaining != "" || targetStatus != "" || cost != "" {
		slog.Info("Scrape.do metadata",
			slog.String("remaining_credits", remaining),
			slog.String("target_status", targetStatus),
			slog.String("cost", cost),
		)
	}
}

func (c *Client) maskTokenInURL(u *url.URL) string {
	return redactURL(u)
}

func validateBaseURL(rawURL string, allowInsecureLoopback bool) (*url.URL, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("failed to parse base URL: host is required")
	}
	if parsed.Scheme == "https" {
		return parsed, nil
	}
	if parsed.Scheme == "http" && allowInsecureLoopback && isLoopbackHost(parsed.Hostname()) {
		return parsed, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrInsecureBaseURL, parsed.Redacted())
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func secureRedirectPolicy(previous func(*http.Request, []*http.Request) error) func(*http.Request, []*http.Request) error {
	return func(next *http.Request, via []*http.Request) error {
		if len(via) == 0 {
			return nil
		}
		origin := via[0].URL
		if next.URL.Scheme != "https" || !strings.EqualFold(next.URL.Host, origin.Host) {
			return ErrUnsafeRedirect
		}
		if previous != nil {
			return previous(next, via)
		}
		if len(via) >= 10 {
			return http.ErrUseLastResponse
		}

		return nil
	}
}
