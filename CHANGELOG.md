# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0] - 2026-03-21

### Added
- **Advanced Configuration**: Integrated `koanf` for multi-layer configuration (defaults, file, env, flags).
- **Profiles Support**: Added support for named profiles in `conf.toml`.
- **Stylized UI**: Added a professional ASCII banner and colorized help output using `lipgloss`.
- **Interactive REPL**: New `repl` command using `reeflective/readline` for manual scraping sessions.
- **Machine Interface**: Added `metadata` command (JSON) and MCP Resource `resource://cli/help` for automated discovery.
- **Enhanced Scrape.do Support**: Added `geoCode`, `session`, `device`, `method`, `headers`, `body`, and `actions` (playWithBrowser) parameters.
- **Response Metadata**: Credits, cost, and target status are now logged to `os.Stderr`.

### Changed
- Standardized help output to follow CLI best practices.
- Refactored `internal/config` and `internal/mcp` for better maintainability and strict linting compliance.

## [1.0.0] - 2026-03-21

### Added
- Initial release of `scrapedoctl`.
- Basic MCP server implementation.
- Core Scrape.do API client with `render` and `super` support.
- Containerized build system (Podman/Containerfile).
