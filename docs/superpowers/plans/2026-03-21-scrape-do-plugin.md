# Scrape.do Claude Code Plugin Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a Claude Code plugin featuring an MCP (Model Context Protocol) server written in Golang 1.26. The server will act as an API client for Scrape.do, providing Claude Code with high-quality scraping tools capable of rendering JS, bypassing captchas, and returning LLM-optimized markdown.

**Architecture:**
- **Go 1.26 MCP Server:** A CLI utility (`cmd/scrape-do-mcp`) serving the Model Context Protocol over `stdio`. It exposes tools (e.g., `scrape_url`) that Claude Code can invoke.
- **Scrape.do Client:** An internal Go package for making structured HTTP requests to the Scrape.do REST API, handling parameters like `render`, `super`, and `output=markdown`.
- **Plugin Manifest:** A `.claude-plugin/plugin.json` file configuring the plugin and auto-starting the Go MCP server.
- **Containerized Dev/Build:** An OCI-compliant multi-stage `Containerfile` using `oraclelinux:10` as a base for building the Go binary and exporting a `scratch` production image, inspired by `gobfd`.
- **Linting:** Strict `golangci-lint` configuration matching high-quality Go projects.

---

## Sprint 1: Go Project Initialization & Tooling
**Goal:** Establish the Go environment, dependency management, linter configuration, and OCI container structure.

### Task 1.1: Go Module and Linter Setup
**Files:**
- Create: `go.mod`
- Modify: `.golangci.yml` (already generated)

- [x] **Step 1: Initialize go module**
  - Run: `go mod init github.com/your-org/cc-scrapedo-plugin`

- [x] **Step 2: Add essential dependencies**
  - Add Mark3d's MCP Go SDK (or similar stdio MCP SDK).
  - Run: `go get github.com/mark3d-xyz/mark3d/mcp-go@latest` (or an appropriate context protocol implementation).

- [x] **Step 3: Commit**
  - `git add go.mod go.sum .golangci.yml Containerfile`
  - `git commit -m "chore: init go 1.26 module, linting, and containerfile"`

---

## Sprint 2: Scrape.do API Client implementation
**Goal:** Write a clean, well-tested Go package for interacting with the Scrape.do API.

### Task 2.1: Scrape.do Client Package
**Files:**
- Create: `pkg/scrapedo/client.go`
- Create: `pkg/scrapedo/client_test.go`

- [x] **Step 1: Write failing test**
  - Mock an HTTP server.
  - Assert that URL, `token`, `render=true`, and `output=markdown` are formatted correctly in the GET request.

- [x] **Step 2: Implement Client structure**
  - Define `Client` struct holding the `token`.
  - Define `ScrapeRequest` struct for parameters (`URL`, `Render`, `Super`, etc.).
  - Implement `Scrape(ctx context.Context, req ScrapeRequest) (string, error)`.

- [x] **Step 3: Verify tests and lint**
  - Run: `go test ./pkg/scrapedo/...`
  - Run: `golangci-lint run`

- [x] **Step 4: Commit**
  - `git add pkg/scrapedo/`
  - `git commit -m "feat: implement scrape.do api client"`

---

## Sprint 3: MCP Server & Plugin Integration
**Goal:** Wrap the Scrape.do client in an MCP server and expose it to Claude Code.

### Task 3.1: Go MCP CLI Implementation
**Files:**
- Create: `cmd/scrape-do-mcp/main.go`
- Create: `internal/mcp/server.go`

- [x] **Step 1: Setup MCP stdio server**
  - Initialize an MCP server instance.
  - Read `SCRAPEDO_TOKEN` from the environment.
  - Register a tool called `scrape_url`.

- [x] **Step 2: Define tool schema and handler**
  - Schema: requires `url` (string), optional `render` (boolean), optional `super` (boolean).
  - Handler: Parses arguments, calls `scrapedo.Client`, returns the resulting markdown as MCP text content.

- [x] **Step 3: Compile**
  - Run: `go build -o bin/scrape-do-mcp ./cmd/scrape-do-mcp`

- [x] **Step 4: Commit**
  - `git add cmd/ internal/`
  - `git commit -m "feat: implement stdio mcp server for scrape.do"`

### Task 3.2: Claude Code Manifest
**Files:**
- Create: `.claude-plugin/plugin.json`
- Create: `README.md`

- [x] **Step 1: Create plugin manifest**
  - Define `.claude-plugin/plugin.json` to automatically register the MCP server.
```json
{
  "name": "scrape-do",
  "version": "1.0.0",
  "description": "Scrape.do integration via MCP",
  "mcpServers": {
    "scrape-do": {
      "command": "${CLAUDE_PLUGIN_ROOT}/bin/scrape-do-mcp",
      "args": [],
      "env": {
        "SCRAPEDO_TOKEN": "${SCRAPEDO_TOKEN}"
      }
    }
  }
}
```

- [x] **Step 2: Write README**
  - Document how to build the Go binary (`make build` or `go build`).
  - Document how to use it in Claude Code (e.g., asking Claude to use `scrape_url`).

- [x] **Step 3: Commit**
  - `git add .claude-plugin/plugin.json README.md`
  - `git commit -m "docs: add plugin manifest and readme"`
