package search

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// httpGet performs an HTTP GET request and returns the response body.
// It returns an error wrapping statusErr if the response status is not 200.
func httpGet(ctx context.Context, client *http.Client, url, prefix string, statusErr error) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: create request: %w", prefix, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: HTTP request: %w", prefix, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: read response: %w", prefix, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w %d: %s", statusErr, resp.StatusCode, body)
	}

	return body, nil
}
