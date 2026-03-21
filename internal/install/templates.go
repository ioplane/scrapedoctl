// Package install handles the installation and configuration of scrapedoctl.
package install

const mcpJSONContent = `{
  "mcpServers": {
    "scrape-do": {
      "command": "scrapedoctl",
      "args": ["mcp"],
      "env": {
        "SCRAPEDO_TOKEN": "${SCRAPEDO_TOKEN}"
      }
    }
  }
}
`

const claudeMDContent = `# scrapedoctl — Project Guidelines

## MCP Tools Available
- ` + "`scrape_url`" + `: Scrape any URL to optimized Markdown
- ` + "`web_search`" + `: Search the web (Google, Bing, Yandex, DuckDuckGo via multiple providers)

## Commands
- ` + "`scrapedoctl search <query>`" + ` — Multi-engine web search
- ` + "`scrapedoctl scrape <url>`" + ` — Scrape URL to markdown
- ` + "`scrapedoctl repl`" + ` — Interactive Cisco-style shell
- ` + "`scrapedoctl version`" + ` — Check version and updates

## Architecture
- Go 1.26, Cobra CLI, MCP over stdio
- ` + "`pkg/search/`" + ` — Multi-provider search (Scrape.do, ScraperAPI, SerpAPI, exec plugins)
- ` + "`pkg/scrapedo/`" + ` — Scrape.do API client
- ` + "`internal/mcp/`" + ` — MCP server with scrape_url and web_search tools
- ` + "`internal/repl/`" + ` — Cisco-style REPL with prefix matching

## Code Style
- golangci-lint v2 with strict config (.golangci.yml)
- Tests in external _test packages
- Error wrapping with %w, sentinel errors with Err prefix
- Functions max 60 lines
`

const agentsMDContent = `# scrapedoctl — Agent Contract

This project provides ` + "`scrapedoctl`" + `, a CLI and MCP server for web scraping and search.

## MCP Server
Run ` + "`scrapedoctl mcp`" + ` to start the MCP server. Available tools:
- ` + "`scrape_url`" + `: Scrape any URL to Markdown
- ` + "`web_search`" + `: Multi-engine web search (Google, Bing, Yandex, DuckDuckGo, Baidu)

## Build & Test
` + "```" + `
podman build -t scrapedoctl-dev --target builder -f Containerfile .
podman run --rm -w /src scrapedoctl-dev go test ./... -count=1
podman run --rm -w /src scrapedoctl-dev golangci-lint run ./...
` + "```" + `

## Code Conventions
- Go 1.26, golangci-lint v2 strict
- TDD: write test first, then implementation
- External test packages (*_test)
- Wrap errors with %w
- Max function length: 60 lines
`

const geminiMDContent = `# scrapedoctl — Gemini Notes

Read ` + "`AGENTS.md`" + ` first if it exists.

## MCP Integration
This project exposes ` + "`scrape_url`" + ` and ` + "`web_search`" + ` MCP tools via ` + "`scrapedoctl mcp`" + `.

## Key Files
- ` + "`pkg/search/`" + ` — Search provider interface and implementations
- ` + "`internal/mcp/server.go`" + ` — MCP server with tool registration
- ` + "`.golangci.yml`" + ` — Linter configuration (v2 format)
- ` + "`Containerfile`" + ` — Development container (OracleLinux 10, Go 1.26)
`
