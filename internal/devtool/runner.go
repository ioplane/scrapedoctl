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
	return RunMany(ctx, repository, [][]string{args}, stdout, stderr)
}

// RunMany executes commands sequentially in one exact Go toolchain container.
func RunMany(
	ctx context.Context,
	repository string,
	commands [][]string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	if len(commands) == 0 {
		return ErrEmptyCommand
	}
	for _, command := range commands {
		if len(command) == 0 {
			return ErrEmptyCommand
		}
	}

	repo, resolveErr := filepath.Abs(repository)
	if resolveErr != nil {
		return fmt.Errorf("resolve repository: %w", resolveErr)
	}

	if configureErr := configurePodman(); configureErr != nil {
		return fmt.Errorf("configure Podman: %w", configureErr)
	}

	runCtx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	//nolint:modernize // Explicit embedded request improves Testcontainers configuration readability.
	ctr, containerErr := testcontainers.GenericContainer(runCtx, testcontainers.GenericContainerRequest{
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
	if containerErr != nil {
		return fmt.Errorf("start toolchain container: %w", containerErr)
	}
	defer func() {
		if terminateErr := testcontainers.TerminateContainer(ctr); terminateErr != nil {
			_, _ = fmt.Fprintf(stderr, "terminate toolchain container: %v\n", terminateErr)
		}
	}()

	for _, command := range commands {
		if err := runContainerCommand(runCtx, ctr, command, stdout, stderr); err != nil {
			return err
		}
	}

	return nil
}

func runContainerCommand(
	ctx context.Context,
	container testcontainers.Container,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	exitCode, output, execErr := container.Exec(ctx, args)
	if execErr != nil {
		return fmt.Errorf("execute container command: %w", execErr)
	}
	if _, copyErr := stdcopy.StdCopy(stdout, stderr, output); copyErr != nil {
		return fmt.Errorf("copy container output: %w", copyErr)
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
