package devtool

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageRequestKeepsNamedProductionImage(t *testing.T) {
	t.Parallel()

	repository := t.TempDir()
	request := imageRequest(repository)

	require.Equal(t, filepath.Clean(repository), request.Context)
	require.Equal(t, "Containerfile", request.Dockerfile)
	require.Equal(t, "localhost/scrapedoctl", request.Repo)
	require.Equal(t, "p0", request.Tag)
	require.True(t, request.KeepImage)
}
