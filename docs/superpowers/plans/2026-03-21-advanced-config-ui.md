# Advanced Configuration & UI Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform `scrapedoctl` into a professional CLI tool with a robust configuration system (`koanf`), styled help (`lipgloss`), and machine-readable metadata.

**Architecture:** 
- **`internal/config`**: Configuration management using `koanf` with profile support and multi-layer merging.
- **`internal/ui`**: Stylized output generation, ASCII banner, and colored help templates using `lipgloss`.
- **`cmd/scrapedoctl`**: New commands (`metadata`, `help` override) and global flags (`--config`, `--profile`).
- **`internal/mcp`**: Exposing CLI documentation as an MCP Resource.

**Tech Stack:** 
- `github.com/knadh/koanf/v2`
- `github.com/knadh/koanf/parsers/toml`
- `github.com/knadh/koanf/providers/file`
- `github.com/knadh/koanf/providers/env`
- `github.com/knadh/koanf/providers/confmap`
- `github.com/charmbracelet/lipgloss`
- `github.com/spf13/cobra`
- `github.com/spf13/pflag`

---

## Sprint 1: Configuration System Implementation
**Goal:** Establish the `internal/config` package and integrate `koanf`.

### Task 1.1: Config Package & Defaults
**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Install dependencies**
  - `podman run --rm -v $(pwd):/src -w /src scrapedoctl-dev go get github.com/knadh/koanf/v2 github.com/knadh/koanf/parsers/toml github.com/knadh/koanf/providers/file github.com/knadh/koanf/providers/env github.com/knadh/koanf/providers/confmap`

- [ ] **Step 2: Define Configuration structs**
  - Define `Config`, `GlobalConfig`, `ReplConfig`, `ProfileConfig`.

- [ ] **Step 3: Implement `Load` function**
  - Implement logic to load defaults, file (if exists), environment, and then override with profile.

- [ ] **Step 4: Write tests for merging logic**
  - Verify that env variables override file settings.
  - Verify that profile settings override global settings.

- [ ] **Step 5: Commit**
  - `git commit -m "feat: implement internal/config with koanf and profile support"`

### Task 1.2: Integrate Config into CLI
**Files:**
- Modify: `cmd/scrapedoctl/main.go`

- [ ] **Step 1: Add global flags to rootCmd**
  - Add `--config` and `--profile` flags.

- [ ] **Step 2: Initialize config in PersistentPreRunE**
  - Call `config.Load` before executing any command.

- [ ] **Step 3: Update existing commands to use loaded config**
  - Pass the loaded config to `scrapedo.NewClient` and REPL.

- [ ] **Step 4: Commit**
  - `git commit -m "feat: integrate config system into CLI root command"`

---

## Sprint 2: Stylized Help & UI
**Goal:** Implement the ASCII banner and colored help system.

### Task 2.1: UI Package & ASCII Banner
**Files:**
- Create: `internal/ui/ui.go`
- Create: `internal/ui/help.go`

- [ ] **Step 1: Install lipgloss**
  - `podman run --rm -v $(pwd):/src -w /src scrapedoctl-dev go get github.com/charmbracelet/lipgloss`

- [ ] **Step 2: Add ASCII Banner**
  - Store the "scrapedoctl" ASCII art as a constant.

- [ ] **Step 3: Implement `PrintHelp` function**
  - Use `lipgloss` to style the output headers, flags, and descriptions.

- [ ] **Step 4: Override Cobra Help**
  - In `cmd/scrapedoctl/main.go`, use `rootCmd.SetHelpFunc` to call our custom UI helper.

- [ ] **Step 5: Commit**
  - `git commit -m "feat: add stylized help and ASCII banner using lipgloss"`

---

## Sprint 3: Machine Interface & MCP Resource
**Goal:** Add the `metadata` command and expose documentation to MCP.

### Task 3.1: Metadata Command
**Files:**
- Create: `cmd/scrapedoctl/metadata.go`

- [ ] **Step 1: Implement `metadata` command**
  - Walk the Cobra command tree and collect names, descriptions, and flags.
  - Output as JSON to `Stdout`.

- [ ] **Step 2: Verify JSON output**
  - `podman run ... bin/scrapedoctl metadata | jq .`

- [ ] **Step 3: Commit**
  - `git commit -m "feat: add metadata command for machine discovery"`

### Task 3.2: MCP Resource Integration
**Files:**
- Modify: `internal/mcp/server.go`

- [ ] **Step 1: Implement `ListResources` in MCP server**
  - Expose `resource://cli/help`.

- [ ] **Step 2: Implement `ReadResource` in MCP server**
  - Return the JSON metadata for the help resource.

- [ ] **Step 3: Commit**
  - `git commit -m "feat: expose CLI documentation as an MCP Resource"`

---

## Sprint 4: Final Validation & Testing
**Goal:** Ensure everything works together and satisfies all requirements.

### Task 4.1: End-to-End Tests
- [ ] **Step 1: Verify config file creation and loading**
- [ ] **Step 2: Verify profile switching works via flags**
- [ ] **Step 3: Verify help output looks correct (visual check)**
- [ ] **Step 4: Verify MCP server correctly serves the documentation resource**
- [ ] **Step 5: Final commit and cleanup**
