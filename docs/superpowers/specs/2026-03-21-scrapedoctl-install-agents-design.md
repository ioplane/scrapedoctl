# Spec: Scrapedoctl Installation & AI Agent Setup

**Date:** 2026-03-21  
**Status:** Approved  
**Author:** Gemini CLI Agent  
**Version:** 0.1.0  

---

## 1. Goal
Implement a seamless installation and configuration experience for both humans and machines. This includes an interactive setup for popular AI agents (Claude, Codex, Gemini, Junie, Kimi, OpenCode) and a first-run auto-setup mechanism.

## 2. Architecture & Components

### 2.1 First-Run Logic
-   The `root` command will check for the existence of the configuration file.
-   If the file is missing AND the current command is NOT `help`, `metadata`, or `install`, the program will automatically launch the `install` command in interactive mode.

### 2.2 Installation Command (`scrapedoctl install`)
-   **Directory Management**: Creates the default data directory (`~/.scrapedoctl/`) and configuration file (`conf.toml`).
-   **Interactive Mode**: Uses `charmbracelet/huh` for a modern CLI questionnaire.
    -   **API Token**: Prompts for the Scrape.do API key.
    -   **Agent Selection**: A multi-select list (Space to select) for configuring AI agents.
-   **Automated Mode**: Supports flags like `--agents claude,gemini` and `--token` for machine-driven setup.

### 2.3 Agent Configuration Support
The tool will inject the `scrapedoctl` server definition into the following paths:
-   **Claude Code**: `~/.claude.json` (MCP servers block)
-   **JetBrains Junie**: `~/.junie/mcp/mcp.json`
-   **Gemini CLI**: `~/.gemini/settings.json`
-   **Codex AI**: `~/.codex/config.toml`
-   **Kimi AI**: `~/.kimi/config.toml`
-   **OpenCode AI**: `~/.opencode.json`

### 2.4 In-Program CLI Configuration
A new `config` command for managing settings:
-   `scrapedoctl config set <key>=<value>`
-   `scrapedoctl config get <key>`
-   `scrapedoctl config list`

## 3. Detailed Requirements

### 3.1 Versioning
-   The tool version is reset to `0.1.0` for the initial release.

### 3.2 Resilience
-   `help` and `metadata` commands MUST work without a valid configuration file.
-   The installer MUST NOT overwrite existing agent configurations but instead append/update the `scrapedoctl` entry.

### 3.3 Security
-   Sensitive values (like the API token) MUST be handled securely during prompts (e.g., using masked input).
-   Tokens MUST NOT be logged to any output during the setup process.

## 4. Testing Strategy
1.  **Mock Filesystem**: Use a temporary directory to test config creation and agent file injection.
2.  **Interactive Logic**: Verify the first-run trigger works as expected.
3.  **Config CLI**: Unit tests for the `set/get` logic in the `config` sub-command.
4.  **Version Check**: Ensure all output (banners, help) reflects `0.1.0`.
