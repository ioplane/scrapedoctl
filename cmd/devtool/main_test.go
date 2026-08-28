package main

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLintCommandIsPinnedAndShellFree(t *testing.T) {
	t.Parallel()

	command := lintCommand()
	require.Equal(t, []string{
		"go", "run", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2",
		"run", "./...",
	}, command)
	require.False(t, slices.Contains([]string{"sh", "bash", "zsh", "python", "python3"}, command[0]))
}
