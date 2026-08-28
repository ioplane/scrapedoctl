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

var errUsage = errors.New("usage: go run ./cmd/devtool <test|verify|audit> [arguments]")

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
		return runTests(ctx, repository, os.Args[2:])
	case "verify":
		return runVerification(ctx, repository)
	case "audit":
		return runAudit(ctx, repository, os.Args[2:])
	default:
		return errUsage
	}
}

func runTests(ctx context.Context, repository string, args []string) error {
	if len(args) == 0 {
		args = []string{"./..."}
	}
	command := append([]string{"go", "test"}, args...)
	if err := devtool.Run(ctx, repository, command, os.Stdout, os.Stderr); err != nil {
		return fmt.Errorf("run tests: %w", err)
	}
	return nil
}

func runVerification(ctx context.Context, repository string) error {
	commands := [][]string{
		{"go", "mod", "verify"},
		{"go", "list", "./..."},
		{"go", "vet", "./..."},
		{"go", "test", "-race", "-count=1", "-shuffle=on", "./..."},
		{"go", "build", "./cmd/scrapedoctl"},
	}
	if err := devtool.RunMany(ctx, repository, commands, os.Stdout, os.Stderr); err != nil {
		return fmt.Errorf("run verification commands: %w", err)
	}
	return nil
}

func runAudit(ctx context.Context, repository string, args []string) error {
	if len(args) > 1 {
		return errUsage
	}
	image := devtool.DefaultAuditImage
	if len(args) == 0 {
		if err := devtool.BuildAuditImage(ctx, repository, os.Stdout); err != nil {
			return fmt.Errorf("prepare local security audit: %w", err)
		}
	}
	if len(args) == 1 {
		image = args[0]
	}
	if err := devtool.AuditLocal(ctx, repository, image, os.Stdout, os.Stderr); err != nil {
		return fmt.Errorf("run local security audit: %w", err)
	}
	return nil
}
