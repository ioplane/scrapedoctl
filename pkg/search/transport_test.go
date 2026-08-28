package search_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ioplane/scrapedoctl/pkg/search"
)

func TestProviderTransportPolicy(t *testing.T) {
	slow := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte(`{"organic_results":[]}`))
	}))
	t.Cleanup(slow.Close)

	var redirectTargetCalls atomic.Int32
	redirectTarget := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		redirectTargetCalls.Add(1)
	}))
	t.Cleanup(redirectTarget.Close)
	redirectSource := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirectTarget.URL, http.StatusFound)
	}))
	t.Cleanup(redirectSource.Close)

	type providerFactory func(string, ...search.HTTPOption) search.Provider
	providers := map[string]providerFactory{
		"brave": func(baseURL string, options ...search.HTTPOption) search.Provider {
			provider := search.NewBraveProvider("token", options...)
			provider.SetBaseURL(baseURL)
			return provider
		},
		"scrapedo": func(baseURL string, options ...search.HTTPOption) search.Provider {
			provider := search.NewScrapedoProvider("token", options...)
			provider.SetBaseURL(baseURL)
			return provider
		},
		"scraperapi": func(baseURL string, options ...search.HTTPOption) search.Provider {
			provider := search.NewScraperAPIProvider("token", options...)
			provider.SetBaseURL(baseURL)
			return provider
		},
		"serpapi": func(baseURL string, options ...search.HTTPOption) search.Provider {
			provider := search.NewSerpAPIProvider("token", options...)
			provider.SetBaseURL(baseURL)
			return provider
		},
	}

	for name, factory := range providers {
		t.Run(name+"/timeout", func(t *testing.T) {
			provider := factory(
				slow.URL,
				search.WithHTTPTransport(slow.Client().Transport),
				search.WithHTTPTimeout(30*time.Millisecond),
			)
			_, err := provider.Search(context.Background(), "query", providerOptions(name))
			require.ErrorIs(t, err, context.DeadlineExceeded)
		})

		t.Run(name+"/cancel", func(t *testing.T) {
			provider := factory(
				slow.URL,
				search.WithHTTPTransport(slow.Client().Transport),
				search.WithHTTPTimeout(time.Second),
			)
			ctx, cancel := context.WithCancel(context.Background())
			timer := time.AfterFunc(30*time.Millisecond, cancel)
			defer timer.Stop()
			defer cancel()
			_, err := provider.Search(ctx, "query", providerOptions(name))
			require.ErrorIs(t, err, context.Canceled)
		})

		t.Run(name+"/redirect", func(t *testing.T) {
			provider := factory(
				redirectSource.URL,
				search.WithHTTPTransport(redirectSource.Client().Transport),
				search.WithHTTPTimeout(time.Second),
			)
			_, err := provider.Search(context.Background(), "query", providerOptions(name))
			require.ErrorIs(t, err, search.ErrUnsafeRedirect)
		})
	}

	if redirectTargetCalls.Load() != 0 {
		t.Fatal("unsafe redirect reached target")
	}
}

func providerOptions(name string) search.Options {
	if name == "brave" {
		return search.Options{Engine: "brave"}
	}
	return search.Options{Engine: "google"}
}
