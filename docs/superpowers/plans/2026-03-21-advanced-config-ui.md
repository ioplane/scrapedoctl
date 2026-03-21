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

- [x] **Step 1: Install dependencies**
- [x] **Step 2: Define Configuration structs**
- [x] **Step 3: Implement `Load` function**
- [x] **Step 4: Write tests for merging logic**
- [x] **Step 5: Commit**

### Task 1.2: Integrate Config into CLI
**Files:**
- Modify: `cmd/scrapedoctl/main.go`

- [x] **Step 1: Add global flags to rootCmd**
- [x] **Step 2: Initialize config in PersistentPreRunE**
- [x] **Step 3: Update existing commands to use loaded config**
- [x] **Step 4: Commit**

---

## Sprint 2: Stylized Help & UI
**Goal:** Implement the ASCII banner and colored help system.

### Task 2.1: UI Package & ASCII Banner
**Files:**
- Create: `internal/ui/ui.go`
- Create: `internal/ui/help.go`

- [x] **Step 1: Install lipgloss**
- [x] **Step 2: Add ASCII Banner**
- [x] **Step 3: Implement `PrintHelp` function**
- [x] **Step 4: Override Cobra Help**
- [x] **Step 5: Commit**

---

## Sprint 3: Machine Interface & MCP Resource
**Goal:** Add the `metadata` command and expose documentation to MCP.

### Task 3.1: Metadata Command
**Files:**
- Create: `cmd/scrapedoctl/metadata.go`

- [x] **Step 1: Implement `metadata` command**
- [x] **Step 2: Verify JSON output**
- [x] **Step 3: Commit**

### Task 3.2: MCP Resource Integration
**Files:**
- Modify: `internal/mcp/server.go`

- [x] **Step 1: Implement `ListResources` in MCP server**
- [x] **Step 2: Implement `ReadResource` in MCP server**
- [x] **Step 3: Commit**

---

## Sprint 4: Final Validation & Testing
**Goal:** Ensure everything works together and satisfies all requirements.

### Task 4.1: End-to-End Tests
- [x] **Step 1: Verify config file creation and loading**
- [x] **Step 2: Verify profile switching works via flags**
- [x] **Step 3: Verify help output looks correct (visual check)**
- [x] **Step 4: Verify MCP server correctly serves the documentation resource**
- [x] **Step 5: Final commit and cleanup**
