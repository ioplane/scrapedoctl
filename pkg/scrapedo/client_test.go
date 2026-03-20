package scrapedo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

// setBaseURL uses unsafe to set the unexported baseURL field for testing.
func setBaseURL(c *scrapedo.Client, url string) {
	// struct layout: token string, baseURL string, httpClient *http.Client
	// string is 16 bytes. baseURL starts at offset 16.
	ptr := unsafe.Pointer(c)
	baseURLPtr := (*string)(unsafe.Pointer(uintptr(ptr) + unsafe.Sizeof("")))
	*baseURLPtr = url
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("empty token", func(t *testing.T) {
		t.Parallel()
		client, err := scrapedo.NewClient("")
		require.ErrorIs(t, err, scrapedo.ErrEmptyToken)
		assert.Nil(t, client)
	})

	t.Run("valid token", func(t *testing.T) {
		t.Parallel()
		client, err := scrapedo.NewClient("test-token")
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestScrape_Success(t *testing.T) {
	t.Parallel()

	expectedResponse := "# Markdown Result\nHello world!"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assert method
		assert.Equal(t, http.MethodGet, r.Method)

		// Assert query params
		q := r.URL.Query()
		assert.Equal(t, "test-token", q.Get("token"))
		assert.Equal(t, "https://example.com", q.Get("url"))
		assert.Equal(t, "markdown", q.Get("output"))
		assert.Equal(t, "true", q.Get("render"))
		assert.Equal(t, "true", q.Get("super"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedResponse))
	}))
	defer server.Close()

	client, err := scrapedo.NewClient("test-token")
	require.NoError(t, err)

	setBaseURL(client, server.URL)

	ctx := context.Background()
	req := scrapedo.ScrapeRequest{
		URL:    "https://example.com",
		Render: true,
		Super:  true,
	}

	result, err := client.Scrape(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
}

func TestScrape_Failures(t *testing.T) {
	t.Parallel()

	t.Run("empty url", func(t *testing.T) {
		t.Parallel()
		client, _ := scrapedo.NewClient("token")
		_, err := client.Scrape(context.Background(), scrapedo.ScrapeRequest{URL: ""})
		require.ErrorIs(t, err, scrapedo.ErrEmptyURL)
	})

	t.Run("api error", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"success":false,"message":"Invalid token."}`))
		}))
		defer server.Close()

		client, _ := scrapedo.NewClient("invalid")
		setBaseURL(client, server.URL)

		_, err := client.Scrape(context.Background(), scrapedo.ScrapeRequest{URL: "https://example.com"})
		require.ErrorIs(t, err, scrapedo.ErrAPI)
		require.ErrorContains(t, err, "scrape.do API error: status 403")
		require.ErrorContains(t, err, "Invalid token.")
	})
}
