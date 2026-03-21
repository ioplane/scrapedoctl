# Documentation & Release Overhaul Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform the project documentation into a professional, multi-language system with Mermaid diagrams, AI-optimized descriptors (`llms.txt`), and automated release pipelines.

**Architecture:**
- **`docs/en` & `docs/ru`**: Hierarchical Markdown documentation with shared numbering.
- **Mermaid diagrams**: Embedded in Markdown for architectural visualization.
- **`llms.txt`**: Standardized file for LLM tool discovery and context.
- **GitHub Actions**: Enhanced `release.yml` to package CHANGELOG.md and plugin artifacts.

---

## Sprint 19: Multi-Language Documentation
**Goal:** Create a structured, bi-lingual documentation base.

### Task 19.1: Directory Structure & Indexing
**Files:**
- Create: `docs/en/00-index.md`
- Create: `docs/ru/00-index.md`
- Create: `docs/README.md` (shared entry point)

- [ ] **Step 1: Create directories**
  - `mkdir -p docs/en docs/ru`
- [ ] **Step 2: Create shared index**
  - Create `docs/README.md` with links to both languages.
- [ ] **Step 3: Create language-specific indices**
  - Implement `docs/en/00-index.md` and `docs/ru/00-index.md` with a table of contents.
- [ ] **Step 4: Commit**
  - `git commit -m "docs: init multi-language structure and indices"`

### Task 19.2: Hierarchical Content & Mermaid
**Files:**
- Create: `docs/en/01-introduction.md`, `docs/en/02-installation.md`, `docs/en/03-usage.md`, `docs/en/04-architecture.md`
- Create: `docs/ru/01-introduction.md`, `docs/ru/02-installation.md`, `docs/ru/03-usage.md`, `docs/ru/04-architecture.md`

- [ ] **Step 1: Port content to Introduction**
  - English and Russian versions.
- [ ] **Step 2: Implement Mermaid diagrams in Architecture**
  - Diagram 1: CLI Flow (User -> Commands -> Cache -> API).
  - Diagram 2: MCP Interaction (Agent -> MCP Server -> Scrape.do).
- [ ] **Step 3: Detailed Usage and Installation guides**
  - Include examples for all new features (Cache, History, Config).
- [ ] **Step 4: Commit**
  - `git commit -m "docs: add detailed multi-language guides and mermaid diagrams"`

---

## Sprint 20: AI Optimization & Standards
**Goal:** Make the repository friendly for AI agents and maintainers.

### Task 20.1: LLM Visibility
**Files:**
- Create: `llms.txt`

- [ ] **Step 1: Implement `llms.txt`**
  - Include project summary, key tools, and links to documentation.
- [ ] **Step 2: Commit**
  - `git commit -m "docs: add llms.txt for AI agent discovery"`

### Task 20.2: README & Branding
**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update README.md**
  - Align with the new documentation structure.
  - Add links to `docs/en` and `docs/ru`.
  - Use advanced Markdown formatting (collapsible sections, stylized quotes).
- [ ] **Step 2: Commit**
  - `git commit -m "docs: update main README to project standards"`

---

## Sprint 21: Automated Release Pipeline
**Goal:** Ensure releases are complete and automated.

### Task 21.1: GitHub Actions Refinement
**Files:**
- Modify: `.github/workflows/release.yml`
- Modify: `.goreleaser.yaml`

- [ ] **Step 1: Update GoReleaser config**
  - Ensure `CHANGELOG.md` and `LICENSE` are included in all archives.
  - Include `.claude-plugin/plugin.json` if required for distributions.
- [ ] **Step 2: Refine Release Action**
  - Ensure the action correctly extracts the latest section of `CHANGELOG.md` for the GitHub Release description.
- [ ] **Step 3: Commit**
  - `git commit -m "chore: refine release automation and artifact packaging"`
