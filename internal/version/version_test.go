package version_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ioplane/scrapedoctl/internal/version"
)

func TestInfo(t *testing.T) {
	t.Parallel()
	info := version.Info()
	assert.Contains(t, info, "scrapedoctl")
	assert.Contains(t, info, version.Version)
}

func TestCheckLatest_Newer(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"tag_name": "v99.0.0",
			"html_url": "https://github.com/ioplane/scrapedoctl/releases/tag/v99.0.0",
		})
	}))
	defer srv.Close()

	tag, url, newer, err := checkWithURL(t, srv.URL)
	require.NoError(t, err)
	assert.Equal(t, "v99.0.0", tag)
	assert.Contains(t, url, "v99.0.0")
	assert.True(t, newer)
}

func TestCheckLatest_SameVersion(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"tag_name": "v" + version.Version,
			"html_url": "https://github.com/ioplane/scrapedoctl/releases",
		})
	}))
	defer srv.Close()

	_, _, newer, err := checkWithURL(t, srv.URL)
	require.NoError(t, err)
	assert.False(t, newer)
}

func TestCheckLatest_APIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, _, _, err := checkWithURL(t, srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestConstants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "ioplane", version.RepoOwner)
	assert.Equal(t, "scrapedoctl", version.RepoName)
	assert.Contains(t, version.RepoURL, "github.com/ioplane/scrapedoctl")
	assert.Contains(t, version.ReleasesURL, "/releases")
}

// checkWithURL simulates CheckLatest against a test server.
func checkWithURL(t *testing.T, baseURL string) (string, string, bool, error) {
	t.Helper()

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return "", "", false, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", false, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", false, fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var rel struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", "", false, fmt.Errorf("decode: %w", err)
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	current := strings.TrimPrefix(version.Version, "v")
	currentBase := strings.SplitN(current, "-", 2)[0]
	newer := latest != currentBase && latest > currentBase

	return rel.TagName, rel.HTMLURL, newer, nil
}
