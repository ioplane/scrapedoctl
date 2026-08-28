package scrapedo_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

const fixtureToken = "fixture-secret-token"

func TestClient_HTTPSPolicy(t *testing.T) {
	t.Parallel()

	t.Run("default endpoint uses HTTPS", func(t *testing.T) {
		t.Parallel()

		var scheme string
		httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			scheme = req.URL.Scheme
			return response(http.StatusOK, "ok"), nil
		})}

		client, err := scrapedo.NewClient(fixtureToken, scrapedo.WithHTTPClient(httpClient))
		require.NoError(t, err)
		_, err = client.Scrape(t.Context(), scrapedo.ScrapeRequest{URL: "https://example.com"})
		require.NoError(t, err)
		assert.Equal(t, "https", scheme)
	})

	t.Run("remote plaintext endpoint is rejected", func(t *testing.T) {
		t.Parallel()

		client, err := scrapedo.NewClient(fixtureToken, scrapedo.WithBaseURL("http://api.example.com"))

		require.ErrorIs(t, err, scrapedo.ErrInsecureBaseURL)
		assert.Nil(t, client)
	})
}

func TestClient_BlocksCredentialRedirect(t *testing.T) {
	t.Parallel()

	var destinationCalls atomic.Int32
	destination := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		destinationCalls.Add(1)
	}))
	defer destination.Close()

	source := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", destination.URL)
		w.WriteHeader(http.StatusFound)
	}))
	defer source.Close()

	client, err := scrapedo.NewClient(
		fixtureToken,
		scrapedo.WithBaseURL(source.URL),
		scrapedo.WithHTTPClient(source.Client()),
	)
	require.NoError(t, err)

	_, err = client.Scrape(t.Context(), scrapedo.ScrapeRequest{URL: "https://example.com"})

	require.ErrorIs(t, err, scrapedo.ErrUnsafeRedirect)
	assert.Zero(t, destinationCalls.Load())
}

func TestClient_Limits(t *testing.T) {
	t.Parallel()

	t.Run("response body is bounded", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("123456789"))
		}))
		defer server.Close()

		client, err := scrapedo.NewClient(
			fixtureToken,
			scrapedo.WithBaseURL(server.URL),
			scrapedo.WithInsecureLoopback(),
			scrapedo.WithMaxResponseBytes(8),
		)
		require.NoError(t, err)

		_, err = client.Scrape(t.Context(), scrapedo.ScrapeRequest{URL: "https://example.com"})

		require.ErrorIs(t, err, scrapedo.ErrResponseTooLarge)
	})

	t.Run("client timeout preserves deadline cause", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			time.Sleep(250 * time.Millisecond)
		}))
		defer server.Close()

		client, err := scrapedo.NewClient(
			fixtureToken,
			scrapedo.WithBaseURL(server.URL),
			scrapedo.WithInsecureLoopback(),
			scrapedo.WithTimeout(20*time.Millisecond),
		)
		require.NoError(t, err)

		_, err = client.Scrape(t.Context(), scrapedo.ScrapeRequest{URL: "https://example.com"})

		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestClient_APIErrorIsTypedBoundedAndRedacted(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Request-ID", "request-42")
		w.Header().Set("Retry-After", "3")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(fixtureToken + strings.Repeat("x", 2_000)))
	}))
	defer server.Close()

	client, err := scrapedo.NewClient(
		fixtureToken,
		scrapedo.WithBaseURL(server.URL),
		scrapedo.WithInsecureLoopback(),
	)
	require.NoError(t, err)

	_, err = client.Scrape(t.Context(), scrapedo.ScrapeRequest{URL: "https://example.com"})

	var apiErr *scrapedo.APIError
	require.ErrorAs(t, err, &apiErr)
	require.ErrorIs(t, err, scrapedo.ErrAPI)
	assert.Equal(t, http.StatusTooManyRequests, apiErr.StatusCode)
	assert.Equal(t, "request-42", apiErr.RequestID)
	assert.Equal(t, 3*time.Second, apiErr.RetryAfter)
	assert.LessOrEqual(t, len(apiErr.BodySample), 1_024)
	assert.NotContains(t, err.Error(), fixtureToken)
}

func TestClient_RedactsDebugMetadata(t *testing.T) {
	var logs bytes.Buffer
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(oldLogger) })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client, err := scrapedo.NewClient(
		fixtureToken,
		scrapedo.WithBaseURL(server.URL),
		scrapedo.WithInsecureLoopback(),
	)
	require.NoError(t, err)

	_, err = client.Scrape(t.Context(), scrapedo.ScrapeRequest{
		URL: "https://example.com",
		Headers: map[string]string{
			"Authorization": "Bearer authorization-secret",
			"Cookie":        "session=cookie-secret",
			"X-API-Key":     "header-api-secret",
		},
	})
	require.NoError(t, err)

	output := logs.String()
	assert.NotContains(t, output, fixtureToken)
	assert.NotContains(t, output, "authorization-secret")
	assert.NotContains(t, output, "cookie-secret")
	assert.NotContains(t, output, "header-api-secret")
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
