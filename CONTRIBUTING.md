# Contributing to scrapedoctl

Thank you for your interest in contributing! This document explains the process for contributing changes and the standards we follow.

## Getting Started

1. Fork the repository and clone it locally
2. Make your changes on a feature branch
3. Ensure all code passes formatting and linting
4. Submit a pull request

## Development Workflow

All Go development should be done inside our Podman development container to guarantee consistency and avoid polluting your host machine.

### Container Setup

```bash
# Build the dev container image
podman build -t scrapedoctl-dev --target builder .
```

### Build, Test, and Lint

```bash
# Download dependencies
podman run --rm -v $(pwd):/src -w /src scrapedoctl-dev go mod tidy

# Run tests
podman run --rm -v $(pwd):/src -w /src scrapedoctl-dev go test ./...

# Build binary
podman run --rm -v $(pwd):/src -w /src scrapedoctl-dev go build -o bin/scrapedoctl ./cmd/scrapedoctl

# Run linter
podman run --rm -v $(pwd):/src -w /src golangci/golangci-lint:v1.56.2 golangci-lint run
```

## Code Standards

### Go Conventions

- **Standard Library First**: Try to use the standard library before adding third-party dependencies.
- **Errors**: Always wrap with context using `%w`: `fmt.Errorf("failed to scrape %s: %w", targetURL, err)`.
- **Error handling**: Use `errors.Is`/`errors.As`, never string matching.
- **Context**: Pass `context.Context` as the first parameter, never store it in a struct.
- **Logging**: Use `log/slog` for all logging. Since this is an MCP server operating over `stdio`, logs MUST be written to `os.Stderr`.

### Linting

The project uses `golangci-lint` with a very strict configuration (see `.golangci.yml`). Your code must pass all linting checks before a PR can be merged.

## Pull Request Process

1. Open an issue first to discuss significant changes.
2. Create a feature branch from `main`.
3. Make focused, reviewable commits with descriptive messages.
4. Ensure tests and linting pass.
5. Add or update tests for new functionality.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
