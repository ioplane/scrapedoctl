package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ioplane/scrapedoctl/internal/install"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

func TestSecretDoesNotLeakAcrossCLIAndGeneratedConfigs(t *testing.T) {
	const (
		fixtureToken  = "fixture-e2e-secret-token"
		authorization = "Bearer fixture-authorization-secret"
		cookie        = "session=fixture-cookie-secret"
		apiKey        = "fixture-header-api-secret"
	)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("SCRAPEDOCTL_E2E_TOKEN", fixtureToken)
	setupTestConfig(t)

	var logs bytes.Buffer
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(oldLogger) })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(server.Close)

	stdout, stderr := captureProcessOutput(t, func() error {
		setToken := newConfigSetCmd()
		setToken.SetArgs([]string{"global.token", "--from-env", "SCRAPEDOCTL_E2E_TOKEN"})
		if err := setToken.Execute(); err != nil {
			return err
		}

		addProvider := newProviderAddCmd()
		addProvider.SetArgs([]string{"brave", "--token-env", "SCRAPEDOCTL_E2E_TOKEN"})
		if err := addProvider.Execute(); err != nil {
			return err
		}
		if err := newConfigListCmd().Execute(); err != nil {
			return err
		}
		if err := newProviderListCmd().Execute(); err != nil {
			return err
		}

		client, err := scrapedo.NewClient(
			fixtureToken,
			scrapedo.WithBaseURL(server.URL),
			scrapedo.WithInsecureLoopback(),
		)
		if err != nil {
			return err
		}
		if _, err = client.Scrape(t.Context(), scrapedo.ScrapeRequest{
			URL: "https://example.com",
			Headers: map[string]string{
				"Authorization": authorization,
				"Cookie":        cookie,
				"X-API-Key":     apiKey,
			},
		}); err != nil {
			return err
		}
		return install.ConfigureAgents([]string{"claude", "codex"}, fixtureToken)
	})
	require.NotEmpty(t, stdout)
	require.NotEmpty(t, logs.String())

	claude, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	require.NoError(t, err)
	codex, err := os.ReadFile(filepath.Join(home, ".codex", "config.toml"))
	require.NoError(t, err)
	corpus := stdout + stderr + logs.String() + string(claude) + string(codex)
	for _, secret := range []string{fixtureToken, authorization, cookie, apiKey} {
		assert.NotContains(t, corpus, secret)
	}
}

func captureProcessOutput(t *testing.T, run func() error) (string, string) {
	t.Helper()
	dir := t.TempDir()
	stdoutFile, err := os.Create(filepath.Join(dir, "stdout"))
	require.NoError(t, err)
	stderrFile, err := os.Create(filepath.Join(dir, "stderr"))
	require.NoError(t, err)

	oldStdout, oldStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = stdoutFile, stderrFile
	runErr := run()
	os.Stdout, os.Stderr = oldStdout, oldStderr
	require.NoError(t, runErr)
	require.NoError(t, stdoutFile.Close())
	require.NoError(t, stderrFile.Close())

	stdout, err := os.ReadFile(stdoutFile.Name())
	require.NoError(t, err)
	stderr, err := os.ReadFile(stderrFile.Name())
	require.NoError(t, err)
	return string(stdout), string(stderr)
}
