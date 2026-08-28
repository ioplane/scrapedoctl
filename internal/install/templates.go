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

const claudeMDContent = "@AGENTS.md\n"

const agentsMDContent = `# scrapedoctl — Agent Contract

This project provides ` + "`scrapedoctl`" + `, a CLI and MCP server for web scraping and search.

## MCP Server
Run ` + "`scrapedoctl mcp`" + ` to start the MCP server. Available tools:
- ` + "`scrape_url`" + `: Scrape any URL to Markdown
- ` + "`web_search`" + `: Multi-engine web search (Brave, Google, Bing, Yandex, DuckDuckGo, Baidu)

## Build & Test
` + "```bash" + `
go run ./cmd/devtool test
go run ./cmd/devtool lint
go run ./cmd/devtool verify
` + "```" + `

## Code Conventions
- Go 1.27.0, golangci-lint v2.13.2 strict
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
- ` + "`Containerfile`" + ` — Go 1.27.0 Trixie build and minimal production image
`
