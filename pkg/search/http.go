package search

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	maxHTTPResponseBytes   = int64(8 << 20)
	maxHTTPErrorSampleSize = int64(4 << 10)
	maxHTTPResponseHeaders = int64(1 << 20)
	defaultHTTPTimeout     = 60 * time.Second
	maxConnectTimeout      = 10 * time.Second
	maxResponseHeadTimeout = 30 * time.Second
	idleConnectionTimeout  = 90 * time.Second
)

var (
	// ErrHTTPResponseTooLarge is returned when a provider response exceeds the shared limit.
	ErrHTTPResponseTooLarge = errors.New("provider response is too large")
	// ErrUnsafeRedirect is returned when a provider redirect changes scheme or authority.
	ErrUnsafeRedirect = errors.New("unsafe provider redirect")
)

type httpClientConfig struct {
	timeout   time.Duration
	transport http.RoundTripper
}

// HTTPOption configures the shared transport used by built-in search providers.
type HTTPOption func(*httpClientConfig)

// WithHTTPTimeout sets the total provider request deadline.
func WithHTTPTimeout(timeout time.Duration) HTTPOption {
	return func(config *httpClientConfig) {
		if timeout > 0 {
			config.timeout = timeout
		}
	}
}

func withHTTPTransport(transport http.RoundTripper) HTTPOption {
	return func(config *httpClientConfig) {
		if transport != nil {
			config.transport = transport
		}
	}
}

func newHTTPClient(options ...HTTPOption) *http.Client {
	config := httpClientConfig{timeout: defaultHTTPTimeout, transport: http.DefaultTransport}
	for _, option := range options {
		option(&config)
	}

	transport := config.transport
	if base, ok := transport.(*http.Transport); ok {
		transport = configuredTransport(base, config.timeout)
	}

	return &http.Client{
		Transport:     transport,
		CheckRedirect: sameOriginRedirect,
		Timeout:       config.timeout,
	}
}

func configuredTransport(base *http.Transport, timeout time.Duration) *http.Transport {
	transport := base.Clone()
	dialer := &net.Dialer{Timeout: min(timeout, maxConnectTimeout), KeepAlive: 30 * time.Second}
	transport.DialContext = dialer.DialContext
	transport.TLSHandshakeTimeout = min(timeout, maxConnectTimeout)
	transport.ResponseHeaderTimeout = min(timeout, maxResponseHeadTimeout)
	transport.IdleConnTimeout = idleConnectionTimeout
	transport.MaxIdleConns = 20
	transport.MaxIdleConnsPerHost = 4
	transport.MaxResponseHeaderBytes = maxHTTPResponseHeaders
	transport.ForceAttemptHTTP2 = true
	return transport
}

func sameOriginRedirect(request *http.Request, previous []*http.Request) error {
	if len(previous) == 0 {
		return nil
	}
	origin := previous[0].URL
	if !strings.EqualFold(request.URL.Scheme, origin.Scheme) ||
		!strings.EqualFold(request.URL.Host, origin.Host) {
		return ErrUnsafeRedirect
	}
	if len(previous) >= 10 {
		return ErrUnsafeRedirect
	}
	return nil
}

// HTTPError describes a non-successful provider response without exposing an unbounded body.
type HTTPError struct {
	StatusCode int
	RequestID  string
	RetryAfter time.Duration
	BodySample string
	cause      error
}

// Error returns a bounded diagnostic string.
func (e *HTTPError) Error() string {
	if e.BodySample == "" {
		return fmt.Sprintf("%v: status %d", e.cause, e.StatusCode)
	}
	return fmt.Sprintf("%v: status %d: %s", e.cause, e.StatusCode, e.BodySample)
}

// Unwrap exposes the provider-specific status sentinel.
func (e *HTTPError) Unwrap() error { return e.cause }

// httpGet performs an HTTP GET request and returns a bounded response body.
func httpGet(
	ctx context.Context, client *http.Client, url, prefix string, statusErr error, secret string,
	headers ...http.Header,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: create request: %w", prefix, err)
	}
	if len(headers) > 0 {
		req.Header = headers[0]
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: HTTP request: %w", prefix, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		sample, sampleErr := io.ReadAll(io.LimitReader(resp.Body, maxHTTPErrorSampleSize))
		if sampleErr != nil {
			return nil, fmt.Errorf("%s: read error response: %w", prefix, sampleErr)
		}
		bodySample := strings.ToValidUTF8(string(sample), "?")
		if secret != "" {
			bodySample = strings.ReplaceAll(bodySample, secret, "***")
		}
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			RequestID:  responseRequestID(resp.Header),
			RetryAfter: parseHTTPRetryAfter(resp.Header.Get("Retry-After")),
			BodySample: bodySample,
			cause:      statusErr,
		}
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxHTTPResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%s: read response: %w", prefix, err)
	}
	if int64(len(body)) > maxHTTPResponseBytes {
		return nil, fmt.Errorf("%s: %w", prefix, ErrHTTPResponseTooLarge)
	}
	return body, nil
}

func responseRequestID(header http.Header) string {
	if requestID := header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}
	return header.Get("Request-Id")
}

func parseHTTPRetryAfter(value string) time.Duration {
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	when, err := http.ParseTime(value)
	if err != nil {
		return 0
	}
	return max(time.Until(when), 0)
}
