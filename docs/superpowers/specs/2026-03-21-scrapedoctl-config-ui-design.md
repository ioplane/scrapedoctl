# Spec: Scrapedoctl Configuration & Advanced UI Design

**Date:** 2026-03-21  
**Status:** Approved  
**Author:** Gemini CLI Agent  

---

## 1. Goal
Transform `scrapedoctl` into a professional, human-and-machine-friendly CLI. This includes implementing a robust configuration system via `koanf`, a stylized help system with ASCII art, and a machine-readable JSON metadata interface.

## 2. Architecture & Components

### 2.1 Configuration System (`internal/config`)
We will use `knadh/koanf` to merge multiple configuration layers with the following precedence (highest to lowest):
1.  **Command-line Flags**
2.  **Environment Variables** (`SCRAPEDO_*`)
3.  **Active Profile Settings** (from `conf.toml`)
4.  **Global Settings** (from `conf.toml`)
5.  **Default Values**

**Config File Location:**  
Default: `~/.scrapedoctl/conf.toml`. Can be overridden with `--config` flag.

**`conf.toml` Structure:**
```toml
[global]
token = "..." # Global API key
base_url = "https://api.scrape.do"
timeout = 60000

[repl]
history_file = "~/.scrapedoctl/history"

[profiles.stealth]
render = true
super = true
geo_code = "us"

[profiles.mobile]
device = "mobile"
render = true
```

### 2.2 Styled UI & Help (`internal/ui`)
We will use `charmbracelet/lipgloss` to render a visually appealing help message.
-   **ASCII Banner:** High-quality "scrapedoctl" ASCII logo.
-   **ANSI Styling:** Color-coded headers, flags, and command descriptions.
-   **Standard Pattern:** Follows standard CLI conventions (Usage, Commands, Flags).

### 2.3 Machine-Readable Interface
To support AI agents and automated discovery:
1.  **`metadata` Command:** A hidden or explicit command that outputs the entire CLI structure (all commands, flags, and descriptions) in JSON.
2.  **MCP Resource:** Expose the CLI manual at `resource://cli/help` within the MCP server.
3.  **MCP Tool:** A tool `get_cli_docs` that returns the same structure for agents to discover capabilities dynamically.

## 3. Detailed Logic & Requirements

### 3.1 Profile Switching
-   If `--profile <name>` is provided, the settings from `[profiles.<name>]` in the config file are merged into the request.
-   Flags provided on the command line override any settings from the profile.

### 3.2 JSON Metadata Schema
```json
{
  "name": "scrapedoctl",
  "version": "1.3.0",
  "commands": [
    {
      "name": "scrape",
      "description": "Scrape a single URL...",
      "flags": [
        { "name": "render", "type": "bool", "default": false, "description": "Execute JS" }
      ]
    }
  ],
  "config": {
    "active_profile": "default",
    "config_file": "/root/.scrapedoctl/conf.toml"
  }
}
```

## 4. Testing Strategy
1.  **Configuration Tests:** Verify correct merging of defaults, files, env vars, and flags.
2.  **UI Tests:** Capture stdout/stderr of the `help` command to ensure correct rendering.
3.  **Metadata Tests:** Validate the JSON structure of the `metadata` command against a predefined schema.
4.  **MCP Tests:** Mock the stdio transport to verify the documentation resource is correctly served.

---

## 5. Security Mandates
-   **No Token Logging:** The API token MUST NEVER be included in the JSON metadata or any help output.
-   **Safe Paths:** Ensure `history_file` and `config_file` paths are correctly expanded (e.g., `~` to home directory).
