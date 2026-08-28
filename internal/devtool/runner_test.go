package devtool_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ioplane/scrapedoctl/internal/devtool"
)

func TestRunRejectsEmptyCommand(t *testing.T) {
	t.Parallel()

	err := devtool.Run(t.Context(), t.TempDir(), nil, io.Discard, io.Discard)

	require.ErrorIs(t, err, devtool.ErrEmptyCommand)
}
