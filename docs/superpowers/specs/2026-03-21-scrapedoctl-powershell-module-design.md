# Spec: Modernized Scrapedoctl PowerShell Module

**Date:** 2026-03-21  
**Status:** Approved  
**Author:** Gemini CLI Agent  
**Version:** 0.1.0  

---

## 1. Goal
Transform the existing PowerShell completion script into a first-class PowerShell module optimized for version 7.4+. The module should provide a native-feeling experience, robust error handling, and modern argument completion.

## 2. Architecture & Components

### 2.1 Function Wrapper (`Invoke-Scrapedoctl`)
- **Purpose**: Wrap the native `scrapedoctl` binary.
- **Alias**: `scrapedoctl`.
- **Error Handling**: Use `$local:PSNativeCommandUseErrorActionPreference = $true` to ensure non-zero exit codes from the native binary are treated as PowerShell errors.
- **Parameter Passing**: Use `ValueFromRemainingArguments` to pass all CLI flags and arguments directly to the binary.

### 2.2 Modern Completion
- **Mechanism**: Use the `[ArgumentCompleter()]` attribute.
- **Logic**: The completer will call the native binary with the `__complete` hidden command (standard Cobra functionality) to retrieve context-aware suggestions.
- **Formatting**: Results will be returned as `[System.Management.Automation.CompletionResult]` objects for rich UI integration (tooltips, etc.).

### 2.3 Binary Discovery
- **Logic**: Search for the `scrapedoctl` binary in:
    1.  The module's directory (for bundled distributions).
    2.  The system `PATH`.
    3.  Common installation directories (e.g., `~/.local/bin`, `/usr/local/bin`).

## 3. Detailed Requirements

### 3.1 PowerShell Compatibility
- Target **PowerShell 7.4+**.
- Ensure cross-platform compatibility (Windows, Linux, macOS).

### 3.2 Module Structure
- **`scrapedoctl.psm1`**: Contains the logic and function definitions.
- **`scrapedoctl.psd1`**: Manifest file with metadata and exported members.

### 3.3 Linting & Standards
- MUST pass **PSScriptAnalyzer** with zero warnings (except where strictly necessary for compatibility).
- Follow standard naming conventions (Verb-Noun).

## 4. Testing Strategy
1.  **Static Analysis**: Run `Invoke-ScriptAnalyzer` on the module files.
2.  **Functional Tests**:
    *   Verify the `scrapedoctl` alias works.
    *   Verify that completion suggestions appear when pressing Tab.
    *   Verify that non-zero exit codes throw catchable exceptions.
3.  **Cross-Platform Verification**: Run tests in the Linux dev container and (if possible) simulate Windows paths.

---

## 5. Security Mandates
- **Safe Execution**: Never use `Invoke-Expression` on user-supplied strings. Use the call operator `&` with array-based arguments.
