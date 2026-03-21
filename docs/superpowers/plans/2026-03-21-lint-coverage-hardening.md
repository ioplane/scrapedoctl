# Lint & Coverage Hardening Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate all 14 golangci-lint warnings and raise test coverage for `cmd/scrapedoctl` (58% -> 75%+) and `internal/mcp` (73% -> 85%+).

**Architecture:**
- Fix lint issues (goimports formatting, gosec G104, revive unused-parameter) across test and production files.
- Add unit tests for untested CLI command RunE paths, MCP server paths, and history table output.
- Use test helpers and dependency injection already present in the codebase.

**Tech Stack:**
- Go 1.26, `testify`, `httptest`, `cobra`, `koanf`
- Podman container `scrapedoctl-dev` for validation

---

## Sprint 26: Lint Fixes
**Goal:** Zero lint warnings from `golangci-lint run ./...`.

### Task 26.1: Fix goimports Formatting
**Files:**
- Modify: `pkg/scrapedo/coverage_test.go`

- [x] **Step 1: Fix import grouping**

Separate internal imports from third-party with a blank line:

```go
package scrapedo_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)
```

- [x] **Step 2: Verify**

Run: `podman run --rm -w /src scrapedoctl-dev goimports -l pkg/scrapedo/coverage_test.go`
Expected: no output (clean)

### Task 26.2: Fix gosec G104 in agents.go
**Files:**
- Modify: `internal/install/agents.go:122`

- [x] **Step 1: Handle the Load error explicitly**

Replace:
```go
_ = k.Load(fileProvider(path), toml.Parser())
```

With:
```go
// Ignore load error — file may not exist yet; we'll create it.
if err := k.Load(fileProvider(path), toml.Parser()); err != nil {
	slog.Debug("existing TOML config not found, creating new", "path", path)
}
```

Add `"log/slog"` to imports if not present.

- [x] **Step 2: Verify**

Run: `podman run --rm -w /src scrapedoctl-dev golangci-lint run ./internal/install/...`
Expected: no G104 warning

### Task 26.3: Fix revive unused-parameter Warnings
**Files:**
- Modify: `internal/cache/cache_test.go:113` — rename `t` to `_`
- Modify: `internal/config/config_test.go:217` — rename `t` to `_`
- Modify: `internal/mcp/server_test.go:45,85` — rename `r` to `_`
- Modify: `pkg/scrapedo/cache_test.go:46` — rename `t` to `_`
- Modify: `pkg/scrapedo/client_test.go:34,70,179,192,232,246` — rename unused params to `_`
- Modify: `pkg/scrapedo/coverage_test.go:34` — rename `r` to `_`

- [x] **Step 1: Rename all unused parameters to `_`**

Apply to each file. Example pattern:
```go
// Before:
func(w http.ResponseWriter, r *http.Request) {
// After:
func(w http.ResponseWriter, _ *http.Request) {
```

```go
// Before:
func TestScrape_CacheSaveError(t *testing.T) {
// After (if t is truly unused):
func TestScrape_CacheSaveError(_ *testing.T) {
```

Note: For top-level test functions, Go requires the `*testing.T` parameter. If `t` is unused, it's better to add a `t.Helper()` call or use `_ = t` at the start. Actually, `revive` allows `_` for closure params but not for top-level test funcs — check if the linter actually flags top-level `TestXxx(t` or only closures.

- [x] **Step 2: Verify all lint clean**

Run: `podman run --rm -w /src scrapedoctl-dev golangci-lint run ./...`
Expected: 0 issues

- [x] **Step 3: Commit**

```bash
git add -A
git commit -m "fix: resolve all golangci-lint warnings (goimports, gosec G104, unused params)"
```

---

## Sprint 27: Coverage — `internal/mcp` (73% -> 85%+)
**Goal:** Test `RunServer`, `NewServer` valid path, and resource error paths.

### Task 27.1: Test NewServer with Valid Token
**Files:**
- Modify: `internal/mcp/server_test.go`

- [x] **Step 1: Write test for valid token**

```go
func TestNewServer_ValidToken(t *testing.T) {
	srv, err := mcp.NewServer("valid-token")
	require.NoError(t, err)
	assert.NotNil(t, srv)
}
```

- [x] **Step 2: Run test**

Run: `podman run --rm -w /src scrapedoctl-dev go test ./internal/mcp/ -run TestNewServer_ValidToken -v`
Expected: PASS

### Task 27.2: Test RunServer Context Cancel with Real Setup
**Files:**
- Modify: `internal/mcp/run_test.go`

- [x] **Step 1: Write test for RunServer with cancelled context**

```go
func TestRunServer_ImmediateCancel(t *testing.T) {
	t.Setenv("SCRAPEDO_TOKEN", "test-token")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	err := mcp.RunServer(ctx)
	// Should return quickly due to cancelled context
	assert.Error(t, err)
}
```

- [x] **Step 2: Run test**

Run: `podman run --rm -w /src scrapedoctl-dev go test ./internal/mcp/ -run TestRunServer_ImmediateCancel -v`
Expected: PASS

### Task 27.3: Test RunServerWithClient with Valid Client
**Files:**
- Modify: `internal/mcp/run_test.go`

- [x] **Step 1: Write test**

```go
func TestRunServerWithClient_Cancel(t *testing.T) {
	client, err := scrapedo.NewClient("test-token")
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = mcp.RunServerWithClient(ctx, client)
	assert.Error(t, err)
}
```

- [x] **Step 2: Run and verify coverage**

Run: `podman run --rm -w /src scrapedoctl-dev go test ./internal/mcp/ -cover -v`
Expected: coverage >= 85%

- [x] **Step 3: Commit**

```bash
git add internal/mcp/
git commit -m "test: improve internal/mcp coverage to 85%+"
```

---

## Sprint 28: Coverage — `cmd/scrapedoctl` (58% -> 75%+)
**Goal:** Test CLI command RunE paths for scrape, repl, mcp, history, and cache error paths.

### Task 28.1: Test Scrape Command Errors
**Files:**
- Modify: `cmd/scrapedoctl/integration_test.go`

- [x] **Step 1: Write test for missing token**

```go
func TestScrapeCmd_MissingToken(t *testing.T) {
	root := newTestRootCmd(t)
	root.SetArgs([]string{"scrape", "https://example.com"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SCRAPEDO_TOKEN")
}
```

- [x] **Step 2: Write test for missing URL**

```go
func TestScrapeCmd_MissingURL(t *testing.T) {
	t.Setenv("SCRAPEDO_TOKEN", "test-token")
	root := newTestRootCmd(t)
	root.SetArgs([]string{"scrape"})
	err := root.Execute()
	assert.Error(t, err)
}
```

- [x] **Step 3: Run tests**

Expected: PASS

### Task 28.2: Test REPL/MCP Command Token Validation
**Files:**
- Modify: `cmd/scrapedoctl/integration_test.go`

- [x] **Step 1: Write tests for missing token errors**

```go
func TestREPLCmd_MissingToken(t *testing.T) {
	root := newTestRootCmd(t)
	root.SetArgs([]string{"repl"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SCRAPEDO_TOKEN")
}

func TestMCPCmd_MissingToken(t *testing.T) {
	root := newTestRootCmd(t)
	root.SetArgs([]string{"mcp"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SCRAPEDO_TOKEN")
}
```

- [x] **Step 2: Run tests**

Expected: PASS

### Task 28.3: Test History and Cache Error Paths
**Files:**
- Modify: `cmd/scrapedoctl/integration_test.go`

- [x] **Step 1: Write test for history with URL arg**

```go
func TestHistoryCmd_WithURL(t *testing.T) {
	root := newTestRootCmd(t)
	root.SetArgs([]string{"history", "https://example.com"})
	// Should succeed but show empty results (no DB entries)
	err := root.Execute()
	assert.NoError(t, err)
}
```

- [x] **Step 2: Write test for config set unsupported key**

```go
func TestConfigSetCmd_UnsupportedKey(t *testing.T) {
	root := newTestRootCmd(t)
	root.SetArgs([]string{"config", "set", "invalid_key=value"})
	err := root.Execute()
	assert.Error(t, err)
}
```

- [x] **Step 3: Run all tests and check coverage**

Run: `podman run --rm -w /src scrapedoctl-dev go test ./cmd/scrapedoctl/ -cover -v`
Expected: coverage >= 75%

- [x] **Step 4: Commit**

```bash
git add cmd/scrapedoctl/
git commit -m "test: improve cmd/scrapedoctl coverage to 75%+"
```

---

## Sprint 29: Final Validation
**Goal:** Confirm zero lint issues and all coverage targets met.

### Task 29.1: Full Validation
- [x] **Step 1: Rebuild container**

```bash
podman build -t scrapedoctl-dev --target builder -f Containerfile .
```

- [x] **Step 2: Run full test suite with coverage**

```bash
podman run --rm -w /src scrapedoctl-dev go test ./... -coverprofile=/tmp/cover.out -covermode=atomic
```

Expected: all PASS, `cmd/scrapedoctl` >= 75%, `internal/mcp` >= 85%

- [x] **Step 3: Run lint**

```bash
podman run --rm -w /src scrapedoctl-dev golangci-lint run ./...
```

Expected: 0 issues

- [x] **Step 4: Final commit if needed**

```bash
git commit -m "chore: lint and coverage hardening complete"
```
