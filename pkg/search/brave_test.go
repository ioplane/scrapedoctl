package search_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ioplane/scrapedoctl/pkg/search"
)

func TestBraveProviderSearchContract(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/res/v1/web/search" {
			t.Errorf("path = %q, want /res/v1/web/search", r.URL.Path)
		}
		if got := r.Header.Get("X-Subscription-Token"); got != "brave-token" {
			t.Errorf("X-Subscription-Token = %q", got)
		}
		want := map[string]string{
			"q": "finance bachelor", "country": "ES", "search_lang": "en",
			"count": "7", "offset": "2",
		}
		for key, value := range want {
			if got := r.URL.Query().Get(key); got != value {
				t.Errorf("%s = %q, want %q", key, got, value)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(
			`{"web":{"results":[{"title":"Finance","url":"https://example.com/finance",` +
				`"description":"Bachelor programme","profile":{"long_name":"example.com"}}]}}`,
		))
	}))
	t.Cleanup(server.Close)

	provider := search.NewBraveProvider("brave-token")
	provider.SetBaseURL(server.URL)
	response, err := provider.Search(context.Background(), "finance bachelor", search.Options{
		Engine: "brave", Lang: "en", Country: "ES", Limit: 7, Page: 3, Raw: true,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if response.Provider != "brave" || response.Engine != "brave" || len(response.Results) != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}
	result := response.Results[0]
	if result.Position != 1 || result.Title != "Finance" || result.URL != "https://example.com/finance" ||
		result.Snippet != "Bachelor programme" || result.DisplayedURL != "example.com" {
		t.Errorf("unexpected result: %+v", result)
	}
	if response.Raw == nil {
		t.Error("Raw is nil")
	}
	_, err = search.NewBraveProvider("").Search(context.Background(), "query", search.Options{})
	if !errors.Is(err, search.ErrBraveEmptyToken) {
		t.Errorf("empty token error = %v", err)
	}
}
