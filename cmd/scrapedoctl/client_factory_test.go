package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ioplane/scrapedoctl/internal/config"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

func TestBuildClientAppliesConfiguredEndpointAndTimeout(t *testing.T) {
	t.Parallel()

	requestReceived := make(chan struct{}, 1)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestReceived <- struct{}{}
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte("too late"))
	}))
	defer server.Close()

	client, err := buildClient(&config.Config{Global: config.GlobalConfig{
		Token:   "fixture-token",
		BaseURL: server.URL,
		Timeout: 20,
	}}, nil, scrapedo.WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	_, err = client.Scrape(context.Background(), scrapedo.ScrapeRequest{URL: "https://example.com"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Scrape() error = %v, want context deadline exceeded", err)
	}
	select {
	case <-requestReceived:
	default:
		t.Fatal("configured endpoint did not receive the request")
	}
}
