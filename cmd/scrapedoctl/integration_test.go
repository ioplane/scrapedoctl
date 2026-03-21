package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "conf.toml")
	
	// Create a default config
	require.NoError(t, os.WriteFile(configPath, []byte(`
[global]
token = "test-token"
[cache]
enabled = true
path = "`+filepath.Join(tmpDir, "cache.db")+`"
`), 0o644))

	// Helper to run command and capture REAL stdout
	runCmd := func(args ...string) (string, string, error) {
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		rOut, wOut, _ := os.Pipe()
		rErr, wErr, _ := os.Pipe()
		os.Stdout = wOut
		os.Stderr = wErr

		outChan := make(chan string)
		errChan := make(chan string)
		go func() {
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, rOut)
			outChan <- buf.String()
		}()
		go func() {
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, rErr)
			errChan <- buf.String()
		}()

		root := newRootCmd()
		fullArgs := append([]string{"--config", configPath}, args...)
		root.SetArgs(fullArgs)
		
		err := root.Execute()
		
		_ = wOut.Close()
		_ = wErr.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		
		return <-outChan, <-errChan, err
	}

	t.Run("metadata command", func(t *testing.T) {
		stdout, _, err := runCmd("metadata")
		require.NoError(t, err)
		assert.Contains(t, stdout, `"name": "scrapedoctl"`)
		assert.Contains(t, stdout, `"version": "0.1.0"`)
	})

	t.Run("config list command", func(t *testing.T) {
		stdout, _, err := runCmd("config", "list")
		require.NoError(t, err)
		assert.Contains(t, stdout, "Global Token: test-token")
	})

	t.Run("config set command", func(t *testing.T) {
		_, _, err := runCmd("config", "set", "global.base_url=https://new-api.com")
		require.NoError(t, err)
		
		// Verify change
		stdout, _, _ := runCmd("config", "list")
		assert.Contains(t, stdout, "Global BaseURL: https://new-api.com")
	})

	t.Run("cache stats command", func(t *testing.T) {
		stdout, _, err := runCmd("cache", "stats")
		require.NoError(t, err)
		assert.Contains(t, stdout, "Total entries:")
	})

	t.Run("cache clear command", func(t *testing.T) {
		stdout, _, err := runCmd("cache", "clear")
		require.NoError(t, err)
		assert.Contains(t, stdout, "Cache cleared successfully")
	})

	t.Run("history command empty", func(t *testing.T) {
		stdout, _, err := runCmd("history", "https://example.com")
		require.NoError(t, err)
		assert.Contains(t, stdout, "No history found")
	})
}

func TestCLI_Integration_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	
	t.Run("missing token error", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "no_token.toml")
		_ = os.WriteFile(configPath, []byte(`[global]`), 0o644)
		
		root := newRootCmd()
		root.SetArgs([]string{"--config", configPath, "scrape", "http://example.com"})
		
		// We can't easily capture the error from RunE when it's wrapped by Cobra in this simple test,
		// but Execute should return it.
		err := root.Execute()
		require.Error(t, err)
	})

	t.Run("invalid config path", func(t *testing.T) {
		root := newRootCmd()
		root.SetArgs([]string{"--config", "/nonexistent/path/conf.toml", "metadata"})
		err := root.Execute()
		require.NoError(t, err)
	})
}
