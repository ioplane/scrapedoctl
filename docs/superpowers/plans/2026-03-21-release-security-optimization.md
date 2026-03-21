# Release Optimization & Enhanced Security Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Optimize binary size with UPX, provide standard Linux/macOS packages (deb, rpm, pkg), and implement advanced security scanning (Trivy, SonarCloud).

**Architecture:**
- **GoReleaser**: Configure `upx` for binary compression and `nfpms` for Linux package generation.
- **GitHub Actions**: Add `security.yml` enhancements for Trivy and SonarCloud integration.

---

## Sprint 22: Binary & Package Optimization
**Goal:** Reduce binary footprint and provide platform-native packages.

### Task 22.1: UPX Integration
**Files:**
- Modify: `.goreleaser.yaml`

- [x] **Step 1: Add UPX configuration**
  - Enabled `upx` for all builds with `compress: best`. See `.goreleaser.yaml:32-35`.
- [x] **Step 2: Commit**
  - Committed in `9bba4fa`.

### Task 22.2: Linux & macOS Packaging
**Files:**
- Modify: `.goreleaser.yaml`

- [x] **Step 1: Configure `nfpms`**
  - Added `.deb` and `.rpm` support with completions, metadata. See `.goreleaser.yaml:37-62`.
- [ ] **Step 2: macOS Packaging**
  - `.pkg` generation not yet implemented (requires GoReleaser Pro or custom hook).
- [x] **Step 3: Commit**
  - Committed in `9bba4fa`.

---

## Sprint 23: Advanced Security Scanning
**Goal:** Implement industry-standard security analysis.

### Task 23.1: Trivy Integration
**Files:**
- Modify: `.github/workflows/security.yml`

- [x] **Step 1: Add Trivy Scan step**
  - Trivy job added to `security.yml:33-45` — filesystem scan with CRITICAL,HIGH severity.
- [x] **Step 2: Commit**
  - Committed in `9bba4fa`.

### Task 23.2: SonarCloud & Quality Gate
**Files:**
- Create: `sonar-project.properties`
- Modify: `.github/workflows/ci.yml`

- [x] **Step 1: Add SonarCloud Analysis**
  - SonarCloud scan added to `ci.yml:39-44` with `continue-on-error: true`.
  - `sonar-project.properties` created at project root.
- [x] **Step 2: Commit**
  - Committed in `9bba4fa`.

---

## Status Summary

| Task | Status | Notes |
|------|--------|-------|
| UPX compression | Done | `.goreleaser.yaml` |
| nfpms (deb/rpm) | Done | `.goreleaser.yaml` |
| macOS .pkg | Not started | Requires GoReleaser Pro or custom hook |
| Trivy scanner | Done | `security.yml` |
| SonarCloud | Done | `ci.yml` + `sonar-project.properties` |

**Overall: ~90% complete.** Only macOS `.pkg` packaging remains — low priority, requires GoReleaser Pro.
