# Advanced Logging Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a comprehensive logging system with rotation, JSON support, and detailed request tracing.

**Architecture:**
- **`internal/config`**: Extend `Config` struct with `LoggingConfig`.
- **`internal/logger`**: New package to initialize `slog` with `lumberjack` support.
- **`pkg/scrapedo`**: Refactor client to use global `slog` methods and add debug logging.
- **`cmd/scrapedoctl`**: Initialize logger in `main.go` and handle directory creation in `install.go`.

**Tech Stack:**
- `log/slog`
- `github.com/natefinch/lumberjack`
- `os`, `path/filepath`

---

## Sprint 12: Core Logger & Config
**Goal:** Implement the logging backend and configuration support.

### Task 12.1: Extend Configuration
**Files:**
- Modify: `internal/config/config.go`

- [ ] **Step 1: Add `LoggingConfig` struct and integrate into `Config`**
- [ ] **Step 2: Update `Load` function with logging defaults**
  - Default level: `info`, format: `json`, path: `/var/log/scrapedoctl/scrapedoctl.log`.

### Task 12.2: Logger Package
**Files:**
- Create: `internal/logger/logger.go`
- Create: `internal/logger/logger_test.go`

- [ ] **Step 1: Install `lumberjack` dependency**
  - `podman run --rm -v $(pwd):/src:Z -w /src scrapedoctl-dev go get github.com/natefinch/lumberjack`
- [ ] **Step 2: Implement `Init` function**
  - Setup `lumberjack.Logger`.
  - Create `slog.Handler` (JSON or Text).
  - Call `slog.SetDefault()`.
- [ ] **Step 3: Implement fallback logic**
  - Use `os.Stderr` if file path is not accessible.
- [ ] **Step 4: Write unit tests for initialization**
- [ ] **Step 5: Commit**
  - `git commit -m "feat: implement internal/logger with rotation and multi-format support"`

---

## Sprint 13: Request Tracing & Triggers
**Goal:** Integrate logging into the Scrape client and CLI lifecycle.

### Task 13.1: Client Refactoring
**Files:**
- Modify: `pkg/scrapedo/client.go`

- [ ] **Step 1: Replace manual logger with global `slog`**
- [ ] **Step 2: Add `debug` level logging for requests**
  - Mask the `token` query param before logging the URL.
- [ ] **Step 3: Update metadata logging to use the global logger**

### Task 13.2: CLI Integration
**Files:**
- Modify: `cmd/scrapedoctl/main.go`
- Modify: `cmd/scrapedoctl/install.go`

- [ ] **Step 1: Initialize logger in `PersistentPreRunE` in `main.go`**
- [ ] **Step 2: Add log directory creation to `install.go`**
  - Provide fallback instructions if `os.MkdirAll` fails.
- [ ] **Step 3: Commit**
  - `git commit -m "feat: integrate advanced logging into scrape client and installer"`

---

## Sprint 14: Final Polish & Coverage
**Goal:** Achieve 100% test coverage and verify behavior.

### Task 14.1: Logging Tests & Docs
- [ ] **Step 1: Verify JSON log format**
- [ ] **Step 2: Verify log rotation works (simulated test)**
- [ ] **Step 3: Update README.md with logging configuration details**
- [ ] **Step 4: Ensure all tests pass with 100% coverage for the new package**
- [ ] **Step 5: Commit & Final Push**
