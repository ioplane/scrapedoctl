package scrapedo

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const maxErrorBodySample = 1_024

var (
	// ErrInsecureBaseURL is returned for a plaintext non-loopback API endpoint.
	ErrInsecureBaseURL = errors.New("plaintext Scrape.do base URL is forbidden")
	// ErrUnsafeRedirect is returned before a credential-bearing request changes origin or downgrades to HTTP.
	ErrUnsafeRedirect = errors.New("unsafe credential redirect blocked")
	// ErrResponseTooLarge is returned when an upstream response exceeds the configured limit.
	ErrResponseTooLarge = errors.New("scrape.do response exceeds configured limit")
	// ErrInvalidTimeout is returned for a non-positive client timeout.
	ErrInvalidTimeout = errors.New("scrape.do timeout must be positive")
	// ErrInvalidResponseLimit is returned for a non-positive response-size limit.
	ErrInvalidResponseLimit = errors.New("scrape.do response limit must be positive")
	// ErrNilHTTPClient is returned when a nil HTTP client is supplied.
	ErrNilHTTPClient = errors.New("scrape.do HTTP client must not be nil")
	// ErrBaseURLMissingHost is returned when an API endpoint has no host.
	ErrBaseURLMissingHost = errors.New("scrape.do base URL host is required")
)

// APIError is a bounded, structured error returned for a non-success upstream response.
type APIError struct {
	StatusCode int
	RequestID  string
	RetryAfter time.Duration
	BodySample string
}

// Error implements error without exposing an unbounded response body.
func (e *APIError) Error() string {
	if e.BodySample == "" {
		return fmt.Sprintf("%s: status %d", ErrAPI, e.StatusCode)
	}

	return fmt.Sprintf("%s: status %d: %s", ErrAPI, e.StatusCode, e.BodySample)
}

// Unwrap makes APIError match ErrAPI through errors.Is.
func (e *APIError) Unwrap() error {
	return ErrAPI
}

func newAPIError(resp *http.Response, body []byte, token string) *APIError {
	sample := body
	if len(sample) > maxErrorBodySample {
		sample = sample[:maxErrorBodySample]
	}

	requestID := resp.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = resp.Header.Get("Scrape.do-Request-Id")
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		RequestID:  requestID,
		RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
		BodySample: redactText(strings.ToValidUTF8(string(sample), "�"), token),
	}
}

func parseRetryAfter(value string) time.Duration {
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if when, err := http.ParseTime(value); err == nil {
		delay := time.Until(when)
		if delay > 0 {
			return delay
		}
	}

	return 0
}
