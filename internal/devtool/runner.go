// Package devtool provides the Go-only local development lifecycle.
package devtool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/testcontainers/testcontainers-go"
)

const (
	goImage = "docker.io/library/golang@sha256:6c83d163c89d1dbe66f8fc466870d66fe6325d15bb5ad1730cdc0b482fab1eec"
	workDir = "/workspace"
)

// ErrEmptyCommand is returned when no container command was provided.
var ErrEmptyCommand = errors.New("container command is required")

// CommandError reports a non-zero exit from the toolchain container.
type CommandError struct {
	ExitCode int
}

// Error implements error.
func (e *CommandError) Error() string {
	return fmt.Sprintf("container command failed with exit code %d", e.ExitCode)
}

// Run executes args inside the exact Go toolchain image through Testcontainers-Go and Podman.
func Run(
	ctx context.Context,
	repository string,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	if len(args) == 0 {
		return ErrEmptyCommand
	}

	repo, err := filepath.Abs(repository)
	if err != nil {
		return fmt.Errorf("resolve repository: %w", err)
	}

	if err := configurePodman(); err != nil {
		return err
	}

	runCtx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	ctr, err := testcontainers.GenericContainer(runCtx, testcontainers.GenericContainerRequest{
		ProviderType: testcontainers.ProviderPodman,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      goImage,
			Entrypoint: []string{"sleep"},
			Cmd:        []string{"infinity"},
			WorkingDir: workDir,
			Binds:      []string{repo + ":" + workDir + ":rw"},
			Labels: map[string]string{
				"io.ioplane.project": "scrapedoctl",
				"io.ioplane.purpose": "local-development",
			},
		},
		Started: true,
	})
	if err != nil {
		return fmt.Errorf("start toolchain container: %w", err)
	}
	defer func() {
		_ = testcontainers.TerminateContainer(ctr)
	}()

	exitCode, output, err := ctr.Exec(runCtx, args)
	if err != nil {
		return fmt.Errorf("execute container command: %w", err)
	}
	if _, err := stdcopy.StdCopy(stdout, stderr, output); err != nil {
		return fmt.Errorf("copy container output: %w", err)
	}
	if exitCode != 0 {
		return &CommandError{ExitCode: exitCode}
	}

	return nil
}

func configurePodman() error {
	socket := filepath.Join(os.TempDir(), "podman", "podman-machine-default-api.sock")
	if _, err := os.Stat(socket); err != nil {
		return fmt.Errorf("locate Podman API socket %q: %w", socket, err)
	}

	settings := map[string]string{
		"DOCKER_HOST": "unix://" + socket,
		"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
		"TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE":    "/run/podman/podman.sock",
	}
	for name, value := range settings {
		if err := os.Setenv(name, value); err != nil {
			return fmt.Errorf("set %s: %w", name, err)
		}
	}

	return nil
}
