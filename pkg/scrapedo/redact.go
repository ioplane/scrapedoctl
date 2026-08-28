package scrapedo

import (
	"net/http"
	"net/url"
	"strings"
)

var safeLoggedHeaders = map[string]struct{}{
	"Accept":       {},
	"Content-Type": {},
	"User-Agent":   {},
}

func redactURL(input *url.URL) string {
	if input == nil {
		return ""
	}

	redacted := *input
	query := redacted.Query()
	if query.Has("token") {
		query.Set("token", "***")
	}
	redacted.RawQuery = query.Encode()

	return redacted.String()
}

func loggedHeaders(headers http.Header) map[string]string {
	safe := make(map[string]string)
	for name := range safeLoggedHeaders {
		if value := headers.Get(name); value != "" {
			safe[name] = value
		}
	}

	return safe
}

func redactText(value, token string) string {
	if token == "" {
		return value
	}

	return strings.ReplaceAll(value, token, "***")
}
