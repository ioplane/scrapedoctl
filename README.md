# scrapedoctl

<p align="center">
  <strong>scrapedoctl</strong><br>
  Go 1.26 MCP Server & CLI for the Scrape.do API
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go 1.26">
  <a href="https://github.com/modelcontextprotocol"><img src="https://img.shields.io/badge/MCP-Protocol-1a73e8?style=for-the-badge" alt="MCP Protocol"></a>
  <a href="https://github.com/ioplane/scrapedoctl/blob/master/LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue?style=for-the-badge" alt="License"></a>
</p>

---

`scrapedoctl` is a fast, dependency-light API client and Model Context Protocol (MCP) server for [Scrape.do](https://scrape.do/). Built specifically to serve as a high-performance, robust tool for AI agents (like Claude Code) to scrape JavaScript-rendered pages and bypass anti-bot protections.

## Key Features

- **MCP Integration:** Exposes Scrape.do over the `stdio` transport using the official Model Context Protocol Go SDK.
- **AI-Optimized Markdown:** Supports the `output=markdown` parameter natively, feeding LLMs exactly the format they need.
- **Go 1.26:** Compiled statically, utilizing modern Go features and standard libraries.
- **Zero-Dependency Core:** The core API client (`pkg/scrapedo`) uses nothing but the standard `net/http` package.
- **Containerized Build:** Ships with an OCI-compliant multi-stage Containerfile for reproducible builds and `scratch` image deployments.

## Quick Start

### For Claude Code Users

Add the plugin to your Claude Code workspace by adding it to your `.mcp.json` or `.claude-plugin/plugin.json`:

```json
{
  "mcpServers": {
    "scrape-do": {
      "command": "path/to/bin/scrapedoctl",
      "args": ["mcp"],
      "env": {
        "SCRAPEDO_TOKEN": "your_api_token_here"
      }
    }
  }
}
```

### For Developers

**Build the CLI:**
All build steps happen inside Podman for a clean environment.
```bash
# Build the dev container
podman build -t scrapedoctl-dev --target builder .

# Compile the binary
podman run --rm -v $(pwd):/src -w /src scrapedoctl-dev go build -o bin/scrapedoctl ./cmd/scrapedoctl
```

## Tools Provided to MCP

The server currently provides the following tool to the LLM:

*   **`scrape_url`**: 
    *   `url` (string, required): The target URL to scrape.
    *   `render` (boolean, optional): Set to `true` to execute JavaScript on the target page.
    *   `super` (boolean, optional): Set to `true` to utilize residential/mobile proxy networks for bypassing blocks.

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on code standards, testing, and adding new features.

## License

[MIT License](LICENSE)
