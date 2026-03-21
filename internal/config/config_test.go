package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ioplane/scrapedoctl/internal/config"
)

func TestLoad(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "conf.toml")

	configContent := `
[global]
token = "file-token"
timeout = 30000

[profiles.stealth]
render = true
super = true
geo_code = "us"
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	t.Run("default and file loading", func(t *testing.T) {
		cfg, err := config.Load(configPath, "")
		require.NoError(t, err)

		assert.Equal(t, "file-token", cfg.Global.Token)
		assert.Equal(t, 30000, cfg.Global.Timeout)
		assert.Equal(t, "https://api.scrape.do", cfg.Global.BaseURL) // from default
		assert.False(t, cfg.Resolved.Render)
	})

	t.Run("environment variable override", func(t *testing.T) {
		t.Setenv("SCRAPEDO_GLOBAL_TOKEN", "env-token")

		cfg, err := config.Load(configPath, "")
		require.NoError(t, err)

		assert.Equal(t, "env-token", cfg.Global.Token)
	})

	t.Run("profile resolution", func(t *testing.T) {
		cfg, err := config.Load(configPath, "stealth")
		require.NoError(t, err)

		assert.Equal(t, "stealth", cfg.ActiveProfile)
		assert.True(t, cfg.Resolved.Render)
		assert.True(t, cfg.Resolved.Super)
		assert.Equal(t, "us", cfg.Resolved.GeoCode)
	})

	t.Run("profile not found", func(t *testing.T) {
		_, err := config.Load(configPath, "non-existent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "profile not found")
	})
}
