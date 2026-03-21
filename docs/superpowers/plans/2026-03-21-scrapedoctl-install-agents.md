# Installation & AI Agent Setup Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement an interactive installer and automated configuration system for Scrape.do AI agent integration.

**Architecture:**
- **`cmd/scrapedoctl/install.go`**: CLI command for installation, using `charmbracelet/huh` for interactive UI.
- **`internal/install`**: Core logic for creating config files and injecting Scrape.do server definitions into AI agent config files (Claude, Gemini, etc.).
- **`cmd/scrapedoctl/config.go`**: CLI command for managing settings (`set`, `get`, `list`).
- **`cmd/scrapedoctl/main.go`**: First-run detection logic and version management.

**Tech Stack:**
- `github.com/charmbracelet/huh` (interactive prompts)
- `github.com/spf13/cobra` (CLI framework)
- `encoding/json` / `github.com/knadh/koanf/parsers/toml` (config parsing/writing)

---

## Sprint 8: Interactive Installer
**Goal:** Create the `install` command and the interactive multi-select UI.

### Task 8.1: Install Command & UI Setup
**Files:**
- Create: `cmd/scrapedoctl/install.go`
- Modify: `cmd/scrapedoctl/main.go`

- [x] **Step 1: Install `huh` dependency**
- [x] **Step 2: Implement basic `install` command**
- [x] **Step 3: Implement `huh` questionnaire**
- [x] **Step 4: Commit**

---

## Sprint 9: Agent Config Injection
**Goal:** Implement logic to modify various AI agent configuration files.

### Task 9.1: Agent Configuration Logic
**Files:**
- Create: `internal/install/agents.go`
- Create: `internal/install/agents_test.go`

- [x] **Step 1: Define Agent Configuration Schemas**
- [x] **Step 2: Implement Injection Logic**
- [x] **Step 3: Write tests for injection**
- [x] **Step 4: Connect UI to Logic**
- [x] **Step 5: Commit**

---

## Sprint 10: Config CLI & First-Run Logic
**Goal:** Management of settings via CLI and automatic setup on first run.

### Task 10.1: Config Management Command
**Files:**
- Create: `cmd/scrapedoctl/config.go`

- [x] **Step 1: Implement `config` subcommands**
- [x] **Step 2: Implement `set` logic**
- [x] **Step 3: Commit**

### Task 10.2: First-Run Auto-Trigger
**Files:**
- Modify: `cmd/scrapedoctl/main.go`

- [x] **Step 1: Implement detection logic**
- [x] **Step 2: Commit**

---

## Sprint 11: Versioning & Final Polish
**Goal:** Reset version to `0.1.0` and verify all help/metadata outputs.

### Task 11.1: Version Update & Docs
**Files:**
- Modify: `cmd/scrapedoctl/main.go`
- Modify: `cmd/scrapedoctl/metadata.go`
- Modify: `CHANGELOG.md`

- [x] **Step 1: Update version constants**
- [x] **Step 2: Update README/Help**
- [x] **Step 3: Final E2E Test**
- [x] **Step 4: Commit & Tag**
