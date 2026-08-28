package search_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ioplane/scrapedoctl/pkg/search"
)

func TestHTTPGetContract(t *testing.T) {
	t.Parallel()

	const secret = "fixture-secret"
	statusErr := errors.New("provider status")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/created":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"ok":true}`))
		case "/large":
			_, _ = w.Write([]byte(strings.Repeat("x", (8<<20)+1)))
		default:
			w.Header().Set("Retry-After", "3")
			w.Header().Set("X-Request-ID", "request-123")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(strings.Repeat("detail ", 800) + secret))
		}
	}))
	t.Cleanup(server.Close)

	body, err := search.HTTPGet(
		context.Background(), server.Client(), server.URL+"/created", "fixture", statusErr, secret,
	)
	require.NoError(t, err)
	require.JSONEq(t, `{"ok":true}`, string(body))

	_, err = search.HTTPGet(
		context.Background(), server.Client(), server.URL+"/large", "fixture", statusErr, secret,
	)
	require.ErrorIs(t, err, search.ErrHTTPResponseTooLarge)

	_, err = search.HTTPGet(
		context.Background(), server.Client(), server.URL+"/error", "fixture", statusErr, secret,
	)
	require.ErrorIs(t, err, statusErr)
	var apiErr *search.HTTPError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, http.StatusTooManyRequests, apiErr.StatusCode)
	require.Equal(t, "request-123", apiErr.RequestID)
	require.Equal(t, 3*time.Second, apiErr.RetryAfter)
	require.LessOrEqual(t, len(apiErr.BodySample), 4<<10)
	require.NotContains(t, apiErr.BodySample, secret)
}
