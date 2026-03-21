package cache

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ioplane/scrapedoctl/internal/config"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

func TestCache(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "cache.db")

	cfg := config.CacheConfig{
		Path:         dbPath,
		TTLDays:      7,
		KeepVersions: 2,
	}

	// We need to set the working directory so goose can find migrations if we use relative paths
	// But in tests, we can probably just use the absolute path to migrations.
	// For this environment, I'll try to find where internal/db/migrations is.
	
	store, err := NewStore(cfg)
	require.NoError(t, err)
	defer store.database.Close()

	ctx := context.Background()
	req := scrapedo.ScrapeRequest{
		URL:    "https://example.com",
		Method: "GET",
	}

	t.Run("empty cache", func(t *testing.T) {
		_, found, err := store.GetResult(ctx, req)
		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("save and retrieve", func(t *testing.T) {
		err := store.SaveResult(ctx, req, "content v1", map[string]any{"status": 200})
		require.NoError(t, err)

		content, found, err := store.GetResult(ctx, req)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "content v1", content)
	})

	t.Run("versioning and cleanup", func(t *testing.T) {
		// Save 2 more versions (total 3, but limit is 2)
		err := store.SaveResult(ctx, req, "content v2", map[string]any{"status": 200})
		require.NoError(t, err)
		err = store.SaveResult(ctx, req, "content v3", map[string]any{"status": 200})
		require.NoError(t, err)

		// Latest should be v3
		content, found, err := store.GetResult(ctx, req)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "content v3", content)

		// Check history count in DB directly
		var count int
		err = store.database.QueryRow("SELECT COUNT(*) FROM scrapes WHERE request_hash = ?", NormalizeAndHash(req)).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count, "Should keep only 2 versions")
	})
}

func TestNormalizeAndHash(t *testing.T) {
	req1 := scrapedo.ScrapeRequest{URL: "https://a.com", Render: true}
	req2 := scrapedo.ScrapeRequest{URL: "https://a.com", Render: true}
	req3 := scrapedo.ScrapeRequest{URL: "https://a.com", Render: false}

	assert.Equal(t, NormalizeAndHash(req1), NormalizeAndHash(req2))
	assert.NotEqual(t, NormalizeAndHash(req1), NormalizeAndHash(req3))

	// Test header sorting
	req4 := scrapedo.ScrapeRequest{URL: "https://a.com", Headers: map[string]string{"A": "1", "B": "2"}}
	req5 := scrapedo.ScrapeRequest{URL: "https://a.com", Headers: map[string]string{"B": "2", "A": "1"}}
	assert.Equal(t, NormalizeAndHash(req4), NormalizeAndHash(req5))
}
