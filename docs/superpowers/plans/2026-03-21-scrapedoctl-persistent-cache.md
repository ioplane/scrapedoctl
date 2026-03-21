# Persistent Caching & History Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a type-safe persistent caching layer using SQLite, sqlc, and goose to save tokens and provide request history.

**Architecture:**
- **`internal/db`**: Database migrations (`goose`) and generated type-safe code (`sqlc`).
- **`internal/cache`**: Caching logic, request normalization, and cleanup policies.
- **`pkg/scrapedo`**: Integration with the caching layer.
- **`cmd/scrapedoctl`**: New commands for history and cache management.

**Tech Stack:**
- `modernc.org/sqlite` (Pure Go SQLite)
- `github.com/pressly/goose/v3` (Migrations)
- `github.com/sqlc-dev/sqlc` (Type-safe SQL)
- `crypto/sha256` (Request hashing)

---

## Sprint 15: Database Infrastructure
**Goal:** Setup migrations and type-safe data access.

### Task 15.1: Database Schema & Tooling
**Files:**
- Create: `internal/db/migrations/001_init_cache.sql`
- Create: `internal/db/queries.sql`
- Create: `sqlc.yaml`

- [x] **Step 1: Install dependencies**
- [x] **Step 2: Define SQL schema**
- [x] **Step 3: Define SQL queries**
- [x] **Step 4: Configure and run sqlc**
- [x] **Step 5: Commit**

---

## Sprint 16: Caching Logic
**Goal:** Implement request normalization and cache management.

### Task 16.1: Cache Package Implementation
**Files:**
- Create: `internal/cache/cache.go`
- Create: `internal/cache/cache_test.go`

- [x] **Step 1: Implement request normalization**
- [x] **Step 2: Implement `Store` & `Retrieve` logic**
- [x] **Step 3: Write comprehensive tests**
- [x] **Step 4: Commit**

---

## Sprint 17: CLI & Integration
**Goal:** Expose cache controls to the user and Claude Code.

### Task 17.1: Scrape Client Integration
**Files:**
- Modify: `pkg/scrapedo/client.go`
- Modify: `cmd/scrapedoctl/main.go`

- [x] **Step 1: Update `Scrape` method**
- [x] **Step 2: Initialize DB in `main.go`**
- [x] **Step 3: Commit**

### Task 17.2: History & Maintenance Commands
**Files:**
- Create: `cmd/scrapedoctl/history.go`
- Create: `cmd/scrapedoctl/cache.go`

- [x] **Step 1: Implement `history` command**
- [x] **Step 2: Implement `cache stats` & `cache clear`**
- [x] **Step 3: Commit**

---

## Sprint 18: Validation & Docs
**Goal:** Final 100% coverage and documentation.

### Task 18.1: Final Polish
- [x] **Step 1: Update README.md with cache settings**
- [x] **Step 2: Verify 100% coverage for `internal/cache` and `internal/db`**
- [x] **Step 3: Final E2E test with real API calls**
- [x] **Step 4: Commit & Push**
