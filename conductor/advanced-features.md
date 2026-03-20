# Scrape.do Plugin & CLI Advanced Implementation Plan

> **Goal:** Transform `scrapedoctl` into a fully-featured CLI with an interactive REPL mode (`readline`) and support for advanced Scrape.do API features (POST, custom headers, response metadata).

---

## Sprint 4: Interactive REPL Mode
**Goal:** Add a `repl` command to the CLI using `github.com/reeflective/readline` for a rich interactive experience.

### Task 4.1: Add `readline` and Setup REPL
**Files:**
- Modify: `go.mod`
- Create: `internal/repl/repl.go`
- Modify: `cmd/scrapedoctl/main.go`

- [ ] **Step 1: Install dependencies**
  - `podman run --rm -v $(pwd):/src -w /src scrapedoctl-dev go get github.com/reeflective/readline`

- [ ] **Step 2: Implement REPL handler**
  - Initialize `readline.NewShell()`.
  - Add simple `scrape` command within the REPL that takes `URL` and parameters.
  - Implement basic tab completion for commands and common domains.

- [ ] **Step 3: Update CLI root command**
  - Add `repl` command to `cmd/scrapedoctl/main.go`.
  - Ensure `SCRAPEDO_TOKEN` is passed to the REPL session.

- [ ] **Step 4: Verify & Commit**
  - Test locally in the container.
  - `git commit -m "feat: add interactive REPL mode using reeflective/readline"`

---

## Sprint 5: Advanced Scrape.do Features
**Goal:** Implement POST requests, custom headers, and response metadata.

### Task 5.1: POST & Headers Support
**Files:**
- Modify: `pkg/scrapedo/client.go`
- Modify: `internal/mcp/server.go`

- [ ] **Step 1: Update `ScrapeRequest`**
  - Add `Method string` (default "GET").
  - Add `Headers map[string]string`.
  - Add `Body []byte`.

- [ ] **Step 2: Implement POST logic in Client**
  - Use `http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))`.
  - Forward provided headers to Scrape.do (prefixed with `Sd-` if appropriate, or using `customHeaders=true`).

- [ ] **Step 3: Update MCP Server**
  - Add `method`, `headers`, and `body` to `toolArgs` in `internal/mcp/server.go`.

### Task 5.2: Response Metadata & Logging
**Files:**
- Modify: `pkg/scrapedo/client.go`
- Modify: `internal/mcp/server.go`

- [ ] **Step 1: Extract Scrape.do Headers**
  - Capture `Scrape-Do-Remaining-Credits`, `Scrape-Do-Target-Status`, and `Scrape-Do-Cost`.
  - Log these to `os.Stderr` using `log/slog` so they are visible in Claude Code's log view but don't interfere with the MCP `stdio` stream.

---

## Sprint 6: Browser Actions & Scripting
**Goal:** Support `playWithBrowser` for complex interactions.

### Task 6.1: Browser Actions Support
**Files:**
- Modify: `pkg/scrapedo/client.go`
- Modify: `internal/mcp/server.go`

- [ ] **Step 1: Implement `Actions` parameter**
  - Add `Actions []Action` to `ScrapeRequest`.
  - Serialize to JSON and pass as `playWithBrowser` query parameter.

- [ ] **Step 2: Update MCP Tool**
  - Expose `actions` as a JSON-encoded string or complex object in the `scrape_url` tool schema.

---

## Sprint 7: Testing & Security Validation
**Goal:** Exhaustive testing with a real API key (user-provided).

- [ ] **Step 1: Real-world smoke test**
  - Run `SCRAPEDO_TOKEN=... bin/scrapedoctl scrape "https://httpbin.org/get"` inside podman.
  - Verify `Remaining-Credits` log output.
  - Verify markdown output quality.

- [ ] **Step 2: Security Audit**
  - Ensure the token is NEVER logged to stdout or committed.
  - Check `os.Stderr` usage for sensitive leaks.
