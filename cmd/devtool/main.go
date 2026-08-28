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

const allPackages = "./..."

var errUsage = errors.New(
	"usage: go run ./cmd/devtool <test|lint|verify|audit|release-gate|release-tools|release-manifest> [arguments]",
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

//nolint:cyclop // The command switch intentionally keeps CLI dispatch explicit.
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
	case "lint":
		return runLint(ctx, repository)
	case "verify":
		return runVerification(ctx, repository)
	case "audit":
		return runAudit(ctx, repository, os.Args[2:])
	case "release-gate":
		if len(os.Args) != 2 {
			return errUsage
		}
		return runReleaseGate(
			func() error { return runVerification(ctx, repository) },
			func() error { return runLint(ctx, repository) },
			func() error { return runAudit(ctx, repository, nil) },
		)
	case "release-tools":
		if len(os.Args) != 2 {
			return errUsage
		}
		if err := devtool.InstallReleaseTools(ctx, repository, os.Stdout); err != nil {
			return fmt.Errorf("install release tools: %w", err)
		}
		return nil
	case "release-manifest":
		if len(os.Args) != 3 {
			return errUsage
		}
		if err := devtool.ValidateReleaseManifest(os.Args[2]); err != nil {
			return fmt.Errorf("validate release manifest: %w", err)
		}
		return nil
	default:
		return errUsage
	}
}

func runReleaseGate(steps ...func() error) error {
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func runTests(ctx context.Context, repository string, args []string) error {
	if len(args) == 0 {
		args = []string{allPackages}
	}
	command := append([]string{"go", "test"}, args...)
	if err := devtool.Run(ctx, repository, command, os.Stdout, os.Stderr); err != nil {
		return fmt.Errorf("run tests: %w", err)
	}
	return nil
}

func runLint(ctx context.Context, repository string) error {
	if err := devtool.Run(ctx, repository, lintCommand(), os.Stdout, os.Stderr); err != nil {
		return fmt.Errorf("run linter: %w", err)
	}
	return nil
}

func lintCommand() []string {
	return []string{
		"go", "run", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2",
		"run", allPackages,
	}
}

func runVerification(ctx context.Context, repository string) error {
	if err := devtool.RunMany(ctx, repository, verificationCommands(), os.Stdout, os.Stderr); err != nil {
		return fmt.Errorf("run verification commands: %w", err)
	}
	return nil
}

func verificationCommands() [][]string {
	return [][]string{
		{"go", "mod", "verify"},
		{"go", "mod", "tidy", "-diff"},
		{"go", "list", allPackages},
		{"go", "vet", allPackages},
		{"go", "test", "-race", "-count=1", "-shuffle=on", allPackages},
		{"go", "build", "./cmd/scrapedoctl"},
	}
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
