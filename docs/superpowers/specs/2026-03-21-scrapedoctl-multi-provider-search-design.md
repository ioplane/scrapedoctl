# Multi-Provider Search & REPL Enhancement — Design Spec

## Goal

Add a `search` command to scrapedoctl with a pluggable multi-provider architecture supporting multiple search engines. Upgrade the REPL to support all CLI commands with tab-completion.

## Architecture Overview

```
┌──────────────────────────────────────────────────────────┐
│  CLI / REPL / MCP                                        │
│  scrapedoctl search "query" --engine yandex              │
└───────────────┬──────────────────────────────────────────┘
                │
        ┌───────▼────────┐
        │  Router         │  Resolves provider by engine name
        │  (search.Router)│  using config + provider registry
        └───────┬────────┘
                │
    ┌───────────┼───────────────┬──────────────┐
    ▼           ▼               ▼              ▼
┌────────┐ ┌─────────┐  ┌──────────┐  ┌──────────────┐
│Scrapedo│ │ SerpAPI  │  │  Brave   │  │ Exec Plugin  │
│Google  │ │ (multi)  │  │  (own)   │  │ (stdin/out)  │
│Plugin  │ │          │  │          │  │              │
└────┬───┘ └────┬─────┘  └────┬─────┘  └──────┬───────┘
     │          │              │               │
     ▼          ▼              ▼               ▼
  scrape.do  serpapi.com   brave.com     user binary
  /plugin/   /search?      /res/v1/     ~/.scrapedoctl/
  google/    engine=X      web/search   plugins/NAME
  search
```

## Core Interface

```go
// pkg/search/provider.go

type SearchResult struct {
    Position     int    `json:"position"`
    Title        string `json:"title"`
    URL          string `json:"url"`
    Snippet      string `json:"snippet"`
    DisplayedURL string `json:"displayed_url,omitempty"`
}

type SearchResponse struct {
    Query    string         `json:"query"`
    Engine   string         `json:"engine"`
    Provider string         `json:"provider"`
    Results  []SearchResult `json:"results"`
    Raw      any            `json:"raw,omitempty"` // Original provider response (--raw)
    Metadata map[string]any `json:"metadata,omitempty"`
}

type SearchOptions struct {
    Engine   string
    Lang     string
    Country  string
    Limit    int
    Page     int
    Raw      bool
}

type Provider interface {
    Name() string
    Engines() []string
    Search(ctx context.Context, query string, opts SearchOptions) (*SearchResponse, error)
}
```

## Built-in Providers

### 1. Scrape.do Google Provider
- Endpoint: `GET https://api.scrape.do/plugin/google/search`
- Engines: `["google"]`
- Params: `token`, `q`, `hl`, `gl`, `google_domain`, `start`, `device`
- Auth: existing `SCRAPEDO_TOKEN`

### 2. SerpAPI Provider
- Endpoint: `GET https://serpapi.com/search`
- Engines: `["google", "bing", "yandex", "duckduckgo", "baidu", "yahoo", "naver"]`
- Key param: `engine=НАЗВАНИЕ`
- Query param varies by engine: `q=` (most), `text=` (yandex), `p=` (yahoo), `query=` (naver)
- Auth: `SERPAPI_TOKEN` env or config

## Exec Plugin System

External binary in `~/.scrapedoctl/plugins/` or path specified in config.

Protocol:
- stdin: JSON `{"query": "...", "engine": "...", "options": {...}}`
- stdout: JSON matching `SearchResponse` schema
- stderr: errors/logs
- exit code 0 = success

Config:
```toml
[providers.my_custom]
type = "exec"
command = "~/.scrapedoctl/plugins/my-search"
engines = ["custom_engine"]
```

## Router Logic

1. User specifies `--engine X` (or default from config)
2. Router iterates registered providers, finds first with engine X
3. If `--provider Y` is explicit, use only that provider
4. If multiple providers support the same engine, `default_provider` from config wins

## Config Schema

```toml
[search]
default_provider = "scrapedo"
default_engine = "google"
default_limit = 10

[providers.scrapedo]
# Uses existing [global] token
engines = ["google"]

[providers.serpapi]
token = ""  # or SERPAPI_TOKEN env
engines = ["google", "bing", "yandex", "duckduckgo", "baidu", "yahoo", "naver"]

[providers.brave]
token = ""  # or BRAVE_SEARCH_TOKEN env
engines = ["brave"]

[providers.my_plugin]
type = "exec"
command = "~/.scrapedoctl/plugins/my-search"
engines = ["custom"]
```

## CLI Command

```
scrapedoctl search <query> [flags]

Flags:
  --engine string      Search engine (default from config)
  --provider string    Force specific provider
  --lang string        Language code (hl)
  --country string     Country code (gl)
  --limit int          Max results (default 10)
  --page int           Page number (default 1)
  --raw                Include raw provider response
  --json               Output as JSON
  --markdown           Output as markdown
```

## MCP Tool

```json
{
  "name": "search",
  "description": "Search the web using multiple engines and providers",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query":    {"type": "string", "description": "Search query"},
      "engine":   {"type": "string", "description": "Search engine: google, bing, yandex, duckduckgo, baidu, brave"},
      "provider": {"type": "string", "description": "Provider: scrapedo, serpapi, brave"},
      "lang":     {"type": "string", "description": "Language code"},
      "country":  {"type": "string", "description": "Country code"},
      "limit":    {"type": "integer", "description": "Max results", "default": 10}
    },
    "required": ["query"]
  }
}
```

## REPL Enhancement

Current REPL only supports `scrape`, `help`, `exit`. Upgrade to support all CLI commands:

```
scrapedoctl> help
Commands:
  search <query> [engine=google] [provider=serpapi] [lang=ru] [limit=10]
  scrape <url> [render=true] [super=true] [no-cache=true] [refresh=true]
  history <url>
  cache stats
  cache clear
  config get <key>
  config set <key>=<value>
  config list
  help [command]
  exit, quit

scrapedoctl> search golang mcp sdk engine=yandex
 #  Title                                     URL
 1  go-sdk - Model Context Protocol           github.com/modelcontextprotocol/go-sdk
 ...
```

Tab-completion in REPL:
- Command names: `search`, `scrape`, `history`, `cache`, `config`, `help`, `exit`
- After `search`: param completions `engine=`, `provider=`, `lang=`, `limit=`
- After `engine=`: values from all configured providers' engines
- After `provider=`: values from config
- After `cache`: `stats`, `clear`
- After `config`: `get`, `set`, `list`

Implementation: extend `Shell` struct with a command registry. Each command = handler + completer. The `reeflective/readline` library already supports custom completers.

## Output Formats

### Table (default for CLI/REPL):
```
 #  Title                                     URL
 1  go-sdk - Model Context Protocol           github.com/...
 2  Building MCP Servers in Go                modelcontextprotocol.io/...
```

### JSON (`--json`):
Full `SearchResponse` struct.

### Markdown (`--markdown`, default for MCP):
```markdown
## Search: "golang mcp sdk" (google via scrapedo)

1. **[go-sdk - Model Context Protocol](https://github.com/...)**
   The official Go SDK for the Model Context Protocol...

2. **[Building MCP Servers in Go](https://modelcontextprotocol.io/...)**
   Learn how to build MCP servers using the Go SDK...
```

## Error Handling

- Missing provider token → clear error with config instructions
- Provider API error → wrap with provider name and engine
- Exec plugin timeout → 30s default, configurable
- Exec plugin bad JSON → parse error with plugin name
- No provider for engine → list available engines from all configured providers

## Package Layout

```
pkg/search/
  provider.go        # Interface + SearchResult/SearchResponse types
  router.go          # Provider registry + routing logic
  scrapedo.go        # Scrape.do Google provider
  serpapi.go         # SerpAPI multi-engine provider
  exec.go            # Exec plugin provider
  format.go          # Table/JSON/Markdown formatters

cmd/scrapedoctl/
  search.go          # CLI search command

internal/repl/
  repl.go            # Enhanced with command registry + completions
  commands.go        # Command handlers (search, scrape, history, cache, config)
  completer.go       # Tab-completion logic

internal/mcp/
  server.go          # Add search tool
```
