# Scrape.do Claude Code Plugin Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a high-performance Scrape.do MCP server in Go 1.26 with strict linting, containerized builds, and direct integration with Claude Code.

**Architecture:**
- **`pkg/scrapedo`**: Core API client using standard `net/http`.
- **`internal/mcp`**: MCP server implementation handling tools and resources.
- **`cmd/scrapedoctl`**: Main entry point with Cobra commands.
- **`Containerfile`**: Multi-stage OCI build for development and production.

**Tech Stack:**
- Go 1.26
- `github.com/modelcontextprotocol/go-sdk`
- `github.com/spf13/cobra`
- `github.com/stretchr/testify`
- Podman (for containerized dev)

---

## Sprint 1: Project Foundation
**Goal:** Initialize project structure, linting, and containerized development environment.

### Task 1.1: Go Module and Linter Setup
**Files:**
- Create: `go.mod`
- Create: `.golangci.yml`
- Create: `Containerfile`

- [x] **Step 1: Initialize go module**
- [x] **Step 2: Add essential dependencies**
- [x] **Step 3: Commit**

---

## Sprint 2: Scrape.do API Client implementation
**Goal:** Write a clean, well-tested Go package for interacting with the Scrape.do API.

### Task 2.1: Scrape.do Client Package
**Files:**
- Create: `pkg/scrapedo/client.go`
- Create: `pkg/scrapedo/client_test.go`

- [x] **Step 1: Write failing test**
- [x] **Step 2: Implement Client structure**
- [x] **Step 3: Verify tests and lint**
- [x] **Step 4: Commit**

---

## Sprint 3: MCP Server & Plugin Integration
**Goal:** Wrap the Scrape.do client in an MCP server and expose it to Claude Code.

### Task 3.1: Go MCP CLI Implementation
**Files:**
- Create: `cmd/scrapedoctl/main.go`
- Create: `internal/mcp/server.go`

- [x] **Step 1: Setup MCP stdio server**
- [x] **Step 2: Define tool schema and handler**
- [x] **Step 3: Compile**
- [x] **Step 4: Commit**

### Task 3.2: Claude Code Manifest
**Files:**
- Create: `.claude-plugin/plugin.json`
- Create: `README.md`

- [x] **Step 1: Create plugin manifest**
- [x] **Step 2: Write README**
- [x] **Step 3: Commit**
