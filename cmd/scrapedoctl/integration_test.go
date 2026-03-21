package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "conf.toml")
	cachePath := filepath.Join(tmpDir, "cache.db")
	
	// Create a default config
	require.NoError(t, os.WriteFile(configPath, []byte(`
[global]
token = "test-token"
base_url = "https://api.scrape.do"
[cache]
enabled = true
path = "`+cachePath+`"
`), 0o644))

	// Helper to run command and capture REAL stdout/stderr
	runCmd := func(args ...string) (string, string, error) {
		// Save old stdout/stderr
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		
		rOut, wOut, _ := os.Pipe()
		rErr, wErr, _ := os.Pipe()
		
		os.Stdout = wOut
		os.Stderr = wErr

		// Use a channel to capture output in a goroutine
		outChan := make(chan string)
		errChan := make(chan string)
		
		go func() {
			var buf bytes.Buffer
			io.Copy(&buf, rOut)
			outChan <- buf.String()
		}()
		
		go func() {
			var buf bytes.Buffer
			io.Copy(&buf, rErr)
			errChan <- buf.String()
		}()

		root := newRootCmd()
		fullArgs := append([]string{"--config", configPath}, args...)
		root.SetArgs(fullArgs)
		
		// Set cobra output to our pipes as well just in case
		root.SetOut(wOut)
		root.SetErr(wErr)
		
		err := root.Execute()
		
		// Close writers and restore stdout/stderr
		wOut.Close()
		wErr.Close()
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

	t.Run("scrape command", func(t *testing.T) {
		// Mock server for Scrape.do API
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("mocked result"))
		}))
		defer server.Close()

		// Temporarily update config with mock server URL
		require.NoError(t, os.WriteFile(configPath, []byte(`
[global]
token = "test-token"
base_url = "`+server.URL+`"
[cache]
enabled = false
`), 0o644))

		stdout, _, err := runCmd("scrape", "https://example.com")
		require.NoError(t, err)
		assert.Contains(t, stdout, "mocked result")
	})

	t.Run("install command mocked", func(t *testing.T) {
		// We need a separate way to run it since we are replacing the command
		root := newRootCmd()
		
		installCmd := &cobra.Command{
			Use:   "install",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("Mocked install execution")
				return nil
			},
		}
		
		// Replace the real install command
		for i, c := range root.Commands() {
			if c.Name() == "install" {
				root.RemoveCommand(c)
				root.AddCommand(installCmd)
				break
			}
		}
		
		outBuf := new(bytes.Buffer)
		root.SetOut(outBuf)
		root.SetArgs([]string{"--config", configPath, "install"})
		
		err := root.Execute()
		require.NoError(t, err)
		assert.Contains(t, outBuf.String(), "Mocked install execution")
	})
}

func TestCLI_Integration_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	
	t.Run("missing token error", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "no_token.toml")
		_ = os.WriteFile(configPath, []byte(`[global]`), 0o644)
		
		root := newRootCmd()
		root.SetArgs([]string{"--config", configPath, "scrape", "http://example.com"})
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		
		err := root.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "token")
	})

	t.Run("invalid config path", func(t *testing.T) {
		root := newRootCmd()
		root.SetArgs([]string{"--config", "/nonexistent/path/conf.toml", "metadata"})
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		
		// Metadata should work even without config (it triggers auto-setup bypass logic)
		err := root.Execute()
		require.NoError(t, err)
	})
}

func TestRunFunction(t *testing.T) {
	// Save and restore os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "conf.toml")
	_ = os.WriteFile(configPath, []byte(`[global]
token = "test"`), 0o644)

	os.Args = []string{"scrapedoctl", "--config", configPath, "metadata"}
	
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	err := run()
	
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = oldStdout
	
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "scrapedoctl")
}
