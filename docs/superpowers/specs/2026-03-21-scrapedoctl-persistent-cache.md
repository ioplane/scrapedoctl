# Spec: Scrapedoctl Persistent Caching & History System

**Date:** 2026-03-21  
**Status:** Approved  
**Author:** Gemini CLI Agent  
**Version:** 0.1.0  

---

## 1. Goal
Implement a persistent caching layer for `scrapedoctl` to save Scrape.do tokens, enable offline access to previous results, and support version comparison.

## 2. Architecture & Components

### 2.1 Database Stack
- **Engine**: `modernc.org/sqlite` (Pure Go, no CGO).
- **Schema Management**: `pressly/goose` for SQL-based migrations.
- **Data Access**: `sqlc` for generating type-safe Go code from SQL queries.

### 2.2 Schema Design (`internal/db`)
**Table: `scrapes`**
- `id` (INTEGER, PK): Unique identifier.
- `request_hash` (TEXT, NOT NULL): SHA256 of the normalized request (URL + sorted parameters/headers).
- `url` (TEXT, NOT NULL): Target URL.
- `method` (TEXT, NOT NULL): HTTP method.
- `content` (TEXT, NOT NULL): The Markdown result from Scrape.do.
- `metadata` (TEXT, NOT NULL): JSON string containing credits, cost, status, and response headers.
- `created_at` (DATETIME, NOT NULL): Timestamp of the record creation.

**Indices**:
- `idx_scrapes_hash`: Fast lookup for cache hits.
- `idx_scrapes_url`: Filtering history by URL.
- `idx_scrapes_date`: Sorting results by time.

### 2.3 Caching Logic
1. **Normalization**: Before every request, the tool generates a `request_hash`.
2. **Lookup**:
   - Check if a record exists for the hash.
   - If `created_at` < `ttl_days` (Config), return cached `content`.
3. **Storage**:
   - Every real API call (including `--refresh`) inserts a **new** row.
   - **Cleanup**: After insertion, delete records for that hash exceeding `keep_versions`.
4. **Flags**:
   - `--no-cache`: Bypass DB entirely.
   - `--refresh`: Force API call and store as a new version.

### 2.4 Configuration (`conf.toml`)
```toml
[cache]
enabled = true
path = "~/.scrapedoctl/cache.db"
ttl_days = 7
keep_versions = 5
max_size_mb = 100
```

## 3. CLI Commands
- `history <url>`: List all stored versions for a specific URL.
- `cache stats`: Show DB size and savings (tokens saved).
- `cache clear`: Delete all stored cache data.

## 4. Testing Strategy
1. **Migration Tests**: Ensure `goose` applies and reverts schema correctly.
2. **SQLC Tests**: Verify generated queries against a real (temporary) SQLite DB.
3. **Logic Tests**: Mock Scrape.do API to verify that cache hits don't trigger network calls and `--refresh` correctly adds versions.
4. **Cleanup Tests**: Verify that `keep_versions` and `max_size_mb` limits are enforced.

## 5. Security Mandates
- **Metadata Sanitization**: Ensure the API token is never stored in the `metadata` JSON field.
- **File Permissions**: The `cache.db` file should be created with `0600` permissions.
