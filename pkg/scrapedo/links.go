// Package scrapedo provides an HTTP client for the Scrape.do API.
package scrapedo

import (
	"net/url"
	"regexp"
	"slices"
	"strings"
)

// linkPatterns matches href="..." and markdown [text](url) patterns.
var linkPatterns = []*regexp.Regexp{
	regexp.MustCompile(`href=["']([^"']+)["']`),
	regexp.MustCompile(`\[(?:[^\]]*)\]\(([^)]+)\)`),
}

// ExtractLinks parses HTML/markdown content and returns unique absolute URLs
// from the same domain as baseURL.
func ExtractLinks(content string, baseURL string) []string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}

	seen := make(map[string]struct{})
	var result []string

	for _, pat := range linkPatterns {
		for _, match := range pat.FindAllStringSubmatch(content, -1) {
			link := normalizeLink(match[1], base)
			if link == "" {
				continue
			}
			if _, exists := seen[link]; exists {
				continue
			}
			seen[link] = struct{}{}
			result = append(result, link)
		}
	}

	slices.Sort(result)

	return result
}

// normalizeLink resolves a raw link against the base URL, strips fragments,
// and returns empty string if the link is cross-domain.
func normalizeLink(raw string, base *url.URL) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "javascript:") || strings.HasPrefix(raw, "mailto:") {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(parsed)
	resolved.Fragment = ""

	if !strings.EqualFold(resolved.Host, base.Host) {
		return ""
	}

	return resolved.String()
}
