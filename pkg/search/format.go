package search

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"text/tabwriter"
)

const (
	maxTitleLen  = 45
	noResultsMsg = "No results found.\n"
)

// FormatTable writes a tabular output to w.
func FormatTable(w io.Writer, resp *Response) error {
	if len(resp.Results) == 0 {
		_, err := fmt.Fprint(w, noResultsMsg)
		return err //nolint:wrapcheck // write-through error
	}

	tw := tabwriter.NewWriter(w, 2, 4, 2, ' ', 0)

	if _, err := fmt.Fprintln(tw, "#\tTitle\tURL"); err != nil {
		return fmt.Errorf("format table header: %w", err)
	}

	for _, r := range resp.Results {
		title := truncateTitle(r.Title, maxTitleLen)
		host := hostname(r.URL)

		if _, err := fmt.Fprintf(
			tw, "%d\t%s\t%s\n",
			r.Position, title, host,
		); err != nil {
			return fmt.Errorf("format table row: %w", err)
		}
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("flush table: %w", err)
	}

	return nil
}

// FormatJSON writes pretty-printed JSON to w.
func FormatJSON(w io.Writer, resp *Response) error {
	if len(resp.Results) == 0 {
		_, err := fmt.Fprint(w, noResultsMsg)
		return err //nolint:wrapcheck // write-through error
	}

	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	if _, err = w.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}

	return nil
}

// FormatMarkdown writes markdown formatted results to w.
func FormatMarkdown(w io.Writer, resp *Response) error {
	if len(resp.Results) == 0 {
		_, err := fmt.Fprint(w, noResultsMsg)
		return err //nolint:wrapcheck // write-through error
	}

	if _, err := fmt.Fprintf(
		w, "## Search: %q (%s via %s)\n\n",
		resp.Query, resp.Engine, resp.Provider,
	); err != nil {
		return fmt.Errorf("format markdown header: %w", err)
	}

	for _, r := range resp.Results {
		title := escapeMarkdown(r.Title)
		snippet := escapeMarkdown(r.Snippet)

		if _, err := fmt.Fprintf(
			w, "%d. **[%s](%s)**\n   %s\n\n",
			r.Position, title, r.URL, snippet,
		); err != nil {
			return fmt.Errorf("format markdown result: %w", err)
		}
	}

	return nil
}

// truncateTitle shortens s to maxLen characters, appending "..." if truncated.
func truncateTitle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}

// hostname extracts the host from a URL string, falling back to the raw string.
func hostname(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}

	return u.Host
}

// escapeMarkdown escapes characters that have special meaning in Markdown.
func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		"`", "\\`",
		"*", "\\*",
		"_", "\\_",
		"[", "\\[",
		"]", "\\]",
		"<", "\\<",
		">", "\\>",
	)

	return replacer.Replace(s)
}
