// Command devtool runs the local development lifecycle through Testcontainers-Go and Podman.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ioplane/scrapedoctl/internal/devtool"
)

var errUsage = errors.New("usage: go run ./cmd/devtool <test|verify> [go test arguments]")

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errUsage
	}

	repository, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get repository directory: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	switch os.Args[1] {
	case "test":
		args := os.Args[2:]
		if len(args) == 0 {
			args = []string{"./..."}
		}
		return devtool.Run(ctx, repository, append([]string{"go", "test"}, args...), os.Stdout, os.Stderr)
	case "verify":
		commands := [][]string{
			{"go", "mod", "verify"},
			{"go", "test", "-race", "-count=1", "./..."},
			{"go", "build", "./cmd/scrapedoctl"},
		}
		for _, command := range commands {
			if err := devtool.Run(ctx, repository, command, os.Stdout, os.Stderr); err != nil {
				return err
			}
		}
		return nil
	default:
		return errUsage
	}
}
