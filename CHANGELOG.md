# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-03-21

### Added
- **Interactive Installer**: New `install` command using `charmbracelet/huh` for multi-agent setup (Claude, Junie, Gemini, Codex, Kimi, OpenCode).
- **First-Run Setup**: Automatic trigger of the installer if no configuration is found.
- **Config Management CLI**: New `config set/list` commands to manage settings from the terminal.
- **Advanced Configuration**: Integrated `koanf` for multi-layer configuration (defaults, file, env, flags).
- **Profiles Support**: Added support for named profiles in `conf.toml`.
- **Stylized UI**: Added a professional ASCII banner and colorized help output using `lipgloss`.
- **Interactive REPL**: New `repl` command using `reeflective/readline` for manual scraping sessions.
- **Machine Interface**: Added `metadata` command (JSON) and MCP Resource `resource://cli/help` for automated discovery.
- **Enhanced Scrape.do Support**: Added `geoCode`, `session`, `device`, `method`, `headers`, `body`, and `actions` (playWithBrowser) parameters.
- **Response Metadata**: Credits, cost, and target status are now logged to `os.Stderr`.
- **CI/CD**: Added GitHub Actions workflows for linting, testing, and automated releases via GoReleaser.
