# Modernized PowerShell Module Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the Scrapedoctl PowerShell module to follow modern best practices (PS 7.4+), including a cmdlet wrapper, advanced completion, and error handling.

**Architecture:**
- **`scrapedoctl.psm1`**: Native Go binary wrapper with `[ArgumentCompleter()]`.
- **`scrapedoctl.psd1`**: Updated manifest with proper versioning and exported members.
- **PSScriptAnalyzer**: Static analysis to ensure code quality.

**Tech Stack:**
- PowerShell 7.6 (in dev container)
- PSScriptAnalyzer
- Go (for generating initial completion templates if needed)

---

## Sprint 24: PowerShell Module Refactoring
**Goal:** Implement the new module structure and logic.

### Task 24.1: Module Logic (`.psm1`)
**Files:**
- Modify: `scrapedoctl.psm1`

- [ ] **Step 1: Define `Invoke-Scrapedoctl` function**
  - Implement the function with `CmdletBinding`.
  - Add `ValueFromRemainingArguments` parameter.
  - Set `$local:PSNativeCommandUseErrorActionPreference = $true`.
- [ ] **Step 2: Implement Modern Completer**
  - Move logic from old script block to `[ArgumentCompleter()]`.
  - Ensure it calls `scrapedoctl __complete`.
  - Map output to `[System.Management.Automation.CompletionResult]`.
- [ ] **Step 3: Binary Discovery**
  - Add logic to find the binary relative to the module or in `PATH`.
- [ ] **Step 4: Exports & Aliases**
  - Set the `scrapedoctl` alias.
  - Export members properly.
- [ ] **Step 5: Commit**
  - `git commit -m "feat: refactor PowerShell module to use Invoke-Scrapedoctl wrapper and modern completer"`

### Task 24.2: Module Manifest (`.psd1`)
**Files:**
- Modify: `scrapedoctl.psd1`

- [ ] **Step 1: Update metadata**
  - Set `PowerShellVersion = '7.4'`.
  - Set `FunctionsToExport = @('Invoke-Scrapedoctl')`.
  - Set `AliasesToExport = @('scrapedoctl')`.
- [ ] **Step 2: Commit**
  - `git commit -m "chore: update PowerShell module manifest for version 0.1.0"`

---

## Sprint 25: Quality Assurance & Validation
**Goal:** Ensure the module is lint-free and functional.

### Task 25.1: Linting & Fixes
- [ ] **Step 1: Run PSScriptAnalyzer**
  - `podman run --rm -v $(pwd):/src:Z -w /src scrapedoctl-dev pwsh -Command "Invoke-ScriptAnalyzer -Path ./scrapedoctl.psm1"`
- [ ] **Step 2: Address all warnings**
  - Specifically replace `Invoke-Expression` with the call operator `&`.
- [ ] **Step 3: Commit**
  - `git commit -m "chore: fix PowerShell linting issues and remove Invoke-Expression"`

### Task 25.2: Functional Verification
- [ ] **Step 1: Test in container**
  - `podman run --rm -v $(pwd):/src:Z -w /src scrapedoctl-dev pwsh -Command "Import-Module ./scrapedoctl.psd1; scrapedoctl metadata"`
- [ ] **Step 2: Verify completion trigger**
  - (Manual verification or simulated Tab-completion test).
- [ ] **Step 3: Final Commit**
  - `git commit -m "docs: finalize PowerShell module overhaul"`
