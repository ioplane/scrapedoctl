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

- [ ] **Step 1: Install `huh` dependency**
  - `podman run --rm -v $(pwd):/src:Z -w /src scrapedoctl-dev go get github.com/charmbracelet/huh`

- [ ] **Step 2: Implement basic `install` command**
  - Add `newInstallCmd` to `cmd/scrapedoctl/main.go`.
  - Create `cmd/scrapedoctl/install.go` with placeholders for the questionnaire.

- [ ] **Step 3: Implement `huh` questionnaire**
  - Prompt for Scrape.do token.
  - Multi-select for agents: Claude, Junie, Gemini, Codex, Kimi, OpenCode.

- [ ] **Step 4: Commit**
  - `git commit -m "feat: add interactive install command with huh UI"`

---

## Sprint 9: Agent Config Injection
**Goal:** Implement logic to modify various AI agent configuration files.

### Task 9.1: Agent Configuration Logic
**Files:**
- Create: `internal/install/agents.go`
- Create: `internal/install/agents_test.go`

- [ ] **Step 1: Define Agent Configuration Schemas**
  - Create constants for default config paths and server entry templates for each agent.

- [ ] **Step 2: Implement Injection Logic**
  - For each agent, implement a function that reads the existing config (if any), appends/updates the `scrapedoctl` server, and writes it back.
  - Support JSON (Claude, Gemini, Junie, OpenCode) and TOML (Codex, Kimi).

- [ ] **Step 3: Write tests for injection**
  - Use `t.TempDir()` to simulate agent config files and verify correct injection.

- [ ] **Step 4: Connect UI to Logic**
  - Update `cmd/scrapedoctl/install.go` to call the injection functions based on user selection.

- [ ] **Step 5: Commit**
  - `git commit -m "feat: implement configuration injection for 6 AI agents"`

---

## Sprint 10: Config CLI & First-Run Logic
**Goal:** Management of settings via CLI and automatic setup on first run.

### Task 10.1: Config Management Command
**Files:**
- Create: `cmd/scrapedoctl/config.go`

- [ ] **Step 1: Implement `config` subcommands**
  - `config list`: Show current config.
  - `config set <key>=<value>`: Update `conf.toml`.
  - `config get <key>`: Show specific value.

- [ ] **Step 2: Implement `set` logic**
  - Update `internal/config` if needed to support writing changes back to the TOML file.

- [ ] **Step 3: Commit**
  - `git commit -m "feat: add config management CLI commands"`

### Task 10.2: First-Run Auto-Trigger
**Files:**
- Modify: `cmd/scrapedoctl/main.go`

- [ ] **Step 1: Implement detection logic**
  - In `PersistentPreRunE`, if `config.Load` fails due to missing file, AND command is not `help`, `metadata`, or `install`, trigger `install`.

- [ ] **Step 2: Commit**
  - `git commit -m "feat: trigger interactive setup on first run"`

---

## Sprint 11: Versioning & Final Polish
**Goal:** Reset version to `0.1.0` and verify all help/metadata outputs.

### Task 11.1: Version Update & Docs
**Files:**
- Modify: `cmd/scrapedoctl/main.go`
- Modify: `cmd/scrapedoctl/metadata.go`
- Modify: `CHANGELOG.md`

- [ ] **Step 1: Update version constants**
  - Set version to `0.1.0` in `main.go` and `metadata.go`.

- [ ] **Step 2: Update README/Help**
  - Ensure the `install` and `config` commands are documented.

- [ ] **Step 3: Final E2E Test**
  - Run the tool in a clean environment (delete `~/.scrapedoctl`).
  - Verify auto-setup trigger and agent config creation.

- [ ] **Step 4: Commit & Tag**
  - `git commit -m "chore: release version 0.1.0"`
  - `git tag v0.1.0`
