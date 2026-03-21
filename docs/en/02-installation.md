# 02 - Installation & Setup

## Prerequisites

- **Go 1.26+** (for building from source)
- **Podman/Docker** (optional, for containerized development)
- **Scrape.do API Token** (available at [scrape.do](https://scrape.do/))

## Building from Source

To build the `scrapedoctl` binary locally:

```bash
# Clone the repository
git clone https://github.com/ioplane/scrapedoctl.git
cd scrapedoctl

# Build the binary
go build -o bin/scrapedoctl ./cmd/scrapedoctl
```

## Interactive Installation

`scrapedoctl` features a built-in interactive installer that sets up your configuration file and automatically integrates with your AI agents.

To trigger the installer, simply run any command without a configuration file:

```bash
./bin/scrapedoctl scrape https://example.com
```

### Supported AI Agents

The installer currently supports automatic configuration for:
- **Claude Code**
- **JetBrains Junie**
- **Gemini CLI**
- **Codex AI**
- **Kimi AI**
- **OpenCode AI**
