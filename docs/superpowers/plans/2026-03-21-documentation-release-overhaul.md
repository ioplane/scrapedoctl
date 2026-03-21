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

- [x] **Step 1: Create directories**
- [x] **Step 2: Create shared index**
- [x] **Step 3: Create language-specific indices**
- [x] **Step 4: Commit**

### Task 19.2: Hierarchical Content & Mermaid
- [x] **Step 1: Port content to Introduction**
- [x] **Step 2: Implement Mermaid diagrams in Architecture**
- [x] **Step 3: Detailed Usage and Installation guides**
- [x] **Step 4: Commit**

---

## Sprint 20: AI Optimization & Standards
**Goal:** Make the repository friendly for AI agents and maintainers.

### Task 20.1: LLM Visibility
- [x] **Step 1: Implement `llms.txt`**
- [x] **Step 2: Commit**

### Task 20.2: README & Branding
- [x] **Step 1: Update README.md**
- [x] **Step 2: Commit**

---

## Sprint 21: Automated Release Pipeline
**Goal:** Ensure releases are complete and automated.

### Task 21.1: GitHub Actions Refinement
- [x] **Step 1: Update GoReleaser config**
- [x] **Step 2: Refine Release Action**
- [x] **Step 3: Commit**
