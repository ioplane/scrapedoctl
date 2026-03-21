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

- [x] **Step 1: Define `Invoke-Scrapedoctl` function**
  - Implemented with `CmdletBinding`, `ValueFromRemainingArguments`.
  - `$local:PSNativeCommandUseErrorActionPreference = $true` set in `process` block.
  - See `scrapedoctl.psm1:9-71`.
- [x] **Step 2: Implement Modern Completer**
  - `[ArgumentCompleter()]` attribute with `scrapedoctl __complete` call.
  - Maps to `[System.Management.Automation.CompletionResult]`.
  - See `scrapedoctl.psm1:20-53`.
- [x] **Step 3: Binary Discovery**
  - `Get-ScrapedoctlBinary` function checks module dir then PATH.
  - See `scrapedoctl.psm1:73-89`.
- [x] **Step 4: Exports & Aliases**
  - `scrapedoctl` alias via `Set-Alias`.
  - `Export-ModuleMember` for functions and alias.
  - See `scrapedoctl.psm1:92-93`.
- [x] **Step 5: Commit**
  - Committed in `5409094`.

### Task 24.2: Module Manifest (`.psd1`)
**Files:**
- Modify: `scrapedoctl.psd1`

- [x] **Step 1: Update metadata**
  - `PowerShellVersion = '7.4'` set.
  - `FunctionsToExport = @('Invoke-Scrapedoctl')` set.
  - `AliasesToExport = @('scrapedoctl')` set.
  - See `scrapedoctl.psd1`.
- [x] **Step 2: Commit**
  - Committed in `5409094`.

---

## Sprint 25: Quality Assurance & Validation
**Goal:** Ensure the module is lint-free and functional.

### Task 25.1: Linting & Fixes
- [x] **Step 1: Run PSScriptAnalyzer**
  - Zero warnings/errors on `scrapedoctl.psm1`.
- [x] **Step 2: Address all warnings**
  - No `Invoke-Expression` usage. Uses call operator `&`. Added `-CommandType Application` to `Get-Command` to avoid circular alias resolution. Fixed `Test-Path` to use `-LiteralPath`.
- [ ] **Step 3: Commit** (pending — part of batch commit)

### Task 25.2: Functional Verification
- [x] **Step 1: Test module import and metadata command**

```bash
podman run --rm -v $(pwd):/src:Z -w /src scrapedoctl-dev pwsh -Command "Import-Module ./scrapedoctl.psd1; Get-Command -Module scrapedoctl"
```

Expected: `Invoke-Scrapedoctl` function and `scrapedoctl` alias listed.

- [ ] **Step 2: Test binary invocation**

```bash
podman run --rm -v $(pwd):/src:Z -w /src scrapedoctl-dev pwsh -Command "Import-Module ./scrapedoctl.psd1; scrapedoctl metadata"
```

Expected: JSON metadata output. Verified — full JSON metadata returned successfully.

- [x] **Step 2: Test binary invocation**
  - `scrapedoctl metadata` via module alias returns correct JSON output.
- [ ] **Step 3: Final Commit** (pending — part of batch commit)

---

## Status Summary

| Task | Status | Notes |
|------|--------|-------|
| Invoke-Scrapedoctl wrapper | Done | `scrapedoctl.psm1` |
| ArgumentCompleter | Done | `scrapedoctl.psm1:20-53` |
| Binary discovery | Done | Fixed `-LiteralPath` + `-CommandType Application` |
| Manifest update | Done | `scrapedoctl.psd1` |
| PSScriptAnalyzer lint | Done | Zero warnings |
| Functional verification | Done | Module imports, metadata works |

**Overall: 100% complete.** All tasks done. Pending batch commit only.
