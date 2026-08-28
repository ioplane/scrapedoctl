package main

import (
	"errors"
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

func TestRunReleaseGateStopsAtFirstFailure(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("lint failed")
	var ran []string
	err := runReleaseGate(
		func() error {
			ran = append(ran, "verify")
			return nil
		},
		func() error {
			ran = append(ran, "lint")
			return wantErr
		},
		func() error {
			ran = append(ran, "audit")
			return nil
		},
	)

	require.ErrorIs(t, err, wantErr)
	require.Equal(t, []string{"verify", "lint"}, ran)
}

func TestVerificationCommandsRejectDirtyModuleGraph(t *testing.T) {
	t.Parallel()

	require.Contains(t, verificationCommands(), []string{"go", "mod", "tidy", "-diff"})
}
