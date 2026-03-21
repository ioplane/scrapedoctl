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

- [ ] **Step 1: Install dependencies**
  - `podman run --rm -v $(pwd):/src:Z -w /src scrapedoctl-dev go get github.com/pressly/goose/v3 modernc.org/sqlite`
- [ ] **Step 2: Define SQL schema**
  - Create `001_init_cache.sql` with the `scrapes` table and indices.
- [ ] **Step 3: Define SQL queries**
  - Create `queries.sql` for `GetLatestScrape`, `InsertScrape`, `GetHistoryByUrl`, `DeleteOldVersions`, `ClearCache`.
- [ ] **Step 4: Configure and run sqlc**
  - Create `sqlc.yaml`.
  - Run `podman run --rm -v $(pwd):/src:Z -w /src sqlc/sqlc generate` (or equivalent).
- [ ] **Step 5: Commit**
  - `git commit -m "feat: setup database schema, migrations and sqlc queries"`

---

## Sprint 16: Caching Logic
**Goal:** Implement request normalization and cache management.

### Task 16.1: Cache Package Implementation
**Files:**
- Create: `internal/cache/cache.go`
- Create: `internal/cache/cache_test.go`

- [ ] **Step 1: Implement request normalization**
  - Function to create a stable SHA256 hash from `ScrapeRequest`.
- [ ] **Step 2: Implement `Store` & `Retrieve` logic**
  - Integration with generated sqlc code.
  - Handle TTL and `keep_versions` cleanup.
- [ ] **Step 3: Write comprehensive tests**
  - Verify normalization, TTL hits/misses, and version cleanup.
- [ ] **Step 4: Commit**
  - `git commit -m "feat: implement caching logic with normalization and versioning"`

---

## Sprint 17: CLI & Integration
**Goal:** Expose cache controls to the user and Claude Code.

### Task 17.1: Scrape Client Integration
**Files:**
- Modify: `pkg/scrapedo/client.go`
- Modify: `cmd/scrapedoctl/main.go`

- [ ] **Step 1: Update `Scrape` method**
  - Inject cache check before API call.
  - Support `--refresh` and `--no-cache` flags.
- [ ] **Step 2: Initialize DB in `main.go`**
  - Run migrations on startup.
- [ ] **Step 3: Commit**
  - `git commit -m "feat: integrate persistent cache into scrape client and cli"`

### Task 17.2: History & Maintenance Commands
**Files:**
- Create: `cmd/scrapedoctl/history.go`
- Create: `cmd/scrapedoctl/cache.go`

- [ ] **Step 1: Implement `history` command**
- [ ] **Step 2: Implement `cache stats` & `cache clear`**
- [ ] **Step 3: Commit**
  - `git commit -m "feat: add history and cache management commands"`

---

## Sprint 18: Validation & Docs
**Goal:** Final 100% coverage and documentation.

### Task 18.1: Final Polish
- [ ] **Step 1: Update README.md with cache settings**
- [ ] **Step 2: Verify 100% coverage for `internal/cache` and `internal/db`**
- [ ] **Step 3: Final E2E test with real API calls**
- [ ] **Step 4: Commit & Push**
