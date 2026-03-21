# Contributing to scrapedoctl

Thank you for your interest in contributing to `scrapedoctl`!

## Code of Conduct

Please be respectful and professional in all interactions.

## Development Workflow

### Prerequisites

- Go 1.26+
- Podman (optional, but recommended for consistent builds)

### Setting Up

1.  Fork and clone the repository.
2.  Install dependencies: `go mod download`.

### Local Development

We recommend developing inside the provided Podman container:

```bash
# Build the dev image
podman build -t scrapedoctl-dev --target builder .

# Run tests
podman run --rm -v $(pwd):/src:Z -w /src scrapedoctl-dev go test ./...

# Run linter
podman run --rm -v $(pwd):/src:Z -w /src scrapedoctl-dev golangci-lint run
```

### Pull Request Process

1.  Create a feature branch from `main`.
2.  Ensure all tests pass and the linter is happy.
3.  Add or update tests for any new functionality.
4.  Update `CHANGELOG.md` following the [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format.
5.  Submit a PR with a clear description of the changes.

## Security

Please refer to [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
