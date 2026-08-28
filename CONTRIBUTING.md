# Contributing to scrapedoctl

Thank you for your interest in contributing to `scrapedoctl`!

## Code of Conduct

Please be respectful and professional in all interactions.

## Development Workflow

### Prerequisites

- Go 1.27.0
- Podman with a running machine

### Setting Up

1.  Fork and clone the repository.
2.  Install dependencies: `go mod download`.

### Local Development

Use the Go development entrypoint. It runs the pinned Go 1.27.0 Trixie
toolchain through Testcontainers-Go and Podman:

```bash
go run ./cmd/devtool test
go run ./cmd/devtool lint
go run ./cmd/devtool verify
go run ./cmd/devtool audit
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
