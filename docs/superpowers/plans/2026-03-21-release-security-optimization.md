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

- [ ] **Step 1: Add UPX configuration**
  - Enable `upx` for Linux and Windows builds.
  - Set compression level to `best`.
- [ ] **Step 2: Commit**
  - `git commit -m "chore: enable UPX compression for Linux and Windows binaries"`

### Task 22.2: Linux & macOS Packaging
**Files:**
- Modify: `.goreleaser.yaml`

- [ ] **Step 1: Configure `nfpms`**
  - Add support for `.deb` and `.rpm` formats.
  - Define file mappings and metadata (maintainer, description).
- [ ] **Step 2: macOS Packaging**
  - Add support for `.pkg` generation (using a custom hook or Pro feature placeholder).
- [ ] **Step 3: Commit**
  - `git commit -m "chore: add .deb, .rpm, and .pkg package generation"`

---

## Sprint 23: Advanced Security Scanning
**Goal:** Implement industry-standard security analysis.

### Task 23.1: Trivy Integration
**Files:**
- Modify: `.github/workflows/security.yml`

- [ ] **Step 1: Add Trivy Scan step**
  - Configure Trivy to scan the repository for vulnerabilities and misconfigurations.
- [ ] **Step 2: Commit**
  - `git commit -m "chore: integrate Trivy security scanner into CI"`

### Task 23.2: SonarCloud & Quality Gate
**Files:**
- Create: `sonar-project.properties`
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Add SonarCloud Analysis**
  - Configure SonarCloud scanner in the CI pipeline.
- [ ] **Step 2: Commit**
  - `git commit -m "chore: add SonarCloud analysis and quality gate"`
