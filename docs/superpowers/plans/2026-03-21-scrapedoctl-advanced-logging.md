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

- [x] **Step 1: Add `LoggingConfig` struct and integrate into `Config`**
- [x] **Step 2: Update `Load` function with logging defaults**

### Task 12.2: Logger Package
**Files:**
- Create: `internal/logger/logger.go`
- Create: `internal/logger/logger_test.go`

- [x] **Step 1: Install `lumberjack` dependency**
- [x] **Step 2: Implement `Init` function**
- [x] **Step 3: Implement fallback logic**
- [x] **Step 4: Write unit tests for initialization**
- [x] **Step 5: Commit**

---

## Sprint 13: Request Tracing & Triggers
**Goal:** Integrate logging into the Scrape client and CLI lifecycle.

### Task 13.1: Client Refactoring
**Files:**
- Modify: `pkg/scrapedo/client.go`

- [x] **Step 1: Replace manual logger with global `slog`**
- [x] **Step 2: Add `debug` level logging for requests**
- [x] **Step 3: Update metadata logging to use the global logger**

### Task 13.2: CLI Integration
**Files:**
- Modify: `cmd/scrapedoctl/main.go`
- Modify: `cmd/scrapedoctl/install.go`

- [x] **Step 1: Initialize logger in `PersistentPreRunE` in `main.go`**
- [x] **Step 2: Add log directory creation to `install.go`**
- [x] **Step 3: Commit**

---

## Sprint 14: Final Polish & Coverage
**Goal:** Achieve 100% test coverage and verify behavior.

### Task 14.1: Logging Tests & Docs
- [x] **Step 1: Verify JSON log format**
- [x] **Step 2: Verify log rotation works (simulated test)**
- [x] **Step 3: Update README.md with logging configuration details**
- [x] **Step 4: Ensure all tests pass with 100% coverage for the new package**
- [x] **Step 5: Commit & Final Push**
