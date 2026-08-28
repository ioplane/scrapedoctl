package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

func TestCrawlOutputContracts(t *testing.T) {
	result := scrapedo.CrawlResult{
		URL:     "https://example.com/docs/start",
		Content: "# Start\n",
		Links:   []string{"https://example.com/docs/next"},
		Depth:   1,
		Size:    8,
	}
	output := t.TempDir()
	require.NoError(t, saveCrawlPage(result, &crawlFlags{output: output, format: "json"}))

	content, err := os.ReadFile(filepath.Join(output, "docs_start.json"))
	require.NoError(t, err)
	var decoded struct {
		URL     string `json:"url"`
		Content string `json:"content"`
	}
	require.NoError(t, json.Unmarshal(content, &decoded))
	require.Equal(t, result.URL, decoded.URL)
	require.Equal(t, result.Content, decoded.Content)

	file := filepath.Join(t.TempDir(), "not-a-directory")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o600))
	require.Error(t, saveCrawlPage(result, &crawlFlags{output: file, format: "markdown"}))

	err = runCrawl(&cobra.Command{}, []string{result.URL}, &crawlFlags{format: "xml"})
	require.ErrorIs(t, err, errUnsupportedCrawlFormat)
}
