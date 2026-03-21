package scrapedo

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaskTokenInURL(t *testing.T) {
	c := &Client{token: "test-token"}
	u, _ := url.Parse("http://api.scrape.do?token=test-token&url=http://example.com")
	masked := c.maskTokenInURL(u)
	assert.Contains(t, masked, "token=%2A%2A%2A")

	u2, _ := url.Parse("http://api.scrape.do?url=http://example.com")
	masked2 := c.maskTokenInURL(u2)
	assert.NotContains(t, masked2, "token=")
}

func TestScrape_DebugLogging(t *testing.T) {
	// Setup a custom logger with Debug level
	// We use a handler that does nothing but enables Debug level
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	oldLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(oldLogger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client, _ := NewClient("test-token")
	client.SetBaseURL(server.URL)

	_, err := client.Scrape(context.Background(), ScrapeRequest{URL: "https://example.com"})
	require.NoError(t, err)
}

func TestPrepareQueryParams_Error(t *testing.T) {
	c := &Client{token: "test-token"}
	_, err := c.prepareQueryParams(ScrapeRequest{
		URL: "http://example.com",
		Actions: []any{
			func() {}, // Cannot marshal func to JSON
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal browser actions")
}
