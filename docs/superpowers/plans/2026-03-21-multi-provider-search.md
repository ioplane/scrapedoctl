# Multi-Provider Search & REPL Enhancement — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `search` command with pluggable multi-provider architecture (Scrape.do, SerpAPI, exec plugins) and upgrade REPL to support all CLI commands with tab-completion.

**Architecture:**
- `pkg/search/` — Provider interface, router, built-in providers, formatters
- `cmd/scrapedoctl/search.go` — CLI command
- `internal/repl/` — Enhanced REPL with command registry and completions
- `internal/mcp/server.go` — MCP search tool

**Tech Stack:**
- Go 1.26, `net/http`, `encoding/json`, `os/exec`
- `github.com/reeflective/readline` (REPL completions)
- `github.com/spf13/cobra` (CLI)
- `github.com/knadh/koanf/v2` (config)

**Spec:** `docs/superpowers/specs/2026-03-21-scrapedoctl-multi-provider-search-design.md`

---

## Sprint 30: Search Provider Interface & Router
**Goal:** Define the core types, interface, and routing logic.

### Task 30.1: Provider Interface & Types
**Files:**
- Create: `pkg/search/provider.go`
- Create: `pkg/search/provider_test.go`

- [x] **Step 1: Write failing test for SearchResult and SearchResponse serialization**

```go
func TestSearchResponse_JSON(t *testing.T) {
    resp := &SearchResponse{
        Query:    "test",
        Engine:   "google",
        Provider: "scrapedo",
        Results: []SearchResult{
            {Position: 1, Title: "Test", URL: "https://example.com", Snippet: "A test"},
        },
    }
    data, err := json.Marshal(resp)
    require.NoError(t, err)
    assert.Contains(t, string(data), `"position":1`)
}
```

- [x] **Step 2: Implement types**

```go
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
    Raw      any            `json:"raw,omitempty"`
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

- [x] **Step 3: Run test, verify pass**

Run: `podman run --rm -w /src scrapedoctl-dev go test ./pkg/search/ -v -run TestSearchResponse`

- [x] **Step 4: Commit**

```bash
git add pkg/search/provider.go pkg/search/provider_test.go
git commit -m "feat(search): add Provider interface and core types"
```

### Task 30.2: Router — Provider Registry & Engine Resolution
**Files:**
- Create: `pkg/search/router.go`
- Create: `pkg/search/router_test.go`

- [x] **Step 1: Write failing tests for router**

```go
func TestRouter_Register(t *testing.T) {
    r := NewRouter()
    mock := &mockProvider{name: "mock", engines: []string{"google", "bing"}}
    r.Register(mock)
    assert.Len(t, r.Providers(), 1)
}

func TestRouter_Resolve_ByEngine(t *testing.T) {
    r := NewRouter()
    r.Register(&mockProvider{name: "p1", engines: []string{"google"}})
    r.Register(&mockProvider{name: "p2", engines: []string{"bing", "yandex"}})

    p, err := r.Resolve("yandex", "")
    require.NoError(t, err)
    assert.Equal(t, "p2", p.Name())
}

func TestRouter_Resolve_ExplicitProvider(t *testing.T) {
    r := NewRouter()
    r.Register(&mockProvider{name: "p1", engines: []string{"google"}})
    r.Register(&mockProvider{name: "p2", engines: []string{"google"}})

    p, err := r.Resolve("google", "p2")
    require.NoError(t, err)
    assert.Equal(t, "p2", p.Name())
}

func TestRouter_Resolve_NoMatch(t *testing.T) {
    r := NewRouter()
    _, err := r.Resolve("yandex", "")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "no provider found")
}

func TestRouter_AllEngines(t *testing.T) {
    r := NewRouter()
    r.Register(&mockProvider{name: "p1", engines: []string{"google"}})
    r.Register(&mockProvider{name: "p2", engines: []string{"bing", "yandex"}})
    engines := r.AllEngines()
    assert.ElementsMatch(t, []string{"google", "bing", "yandex"}, engines)
}
```

- [x] **Step 2: Implement Router**

```go
type Router struct {
    providers []Provider
}

func NewRouter() *Router
func (r *Router) Register(p Provider)
func (r *Router) Resolve(engine, providerName string) (Provider, error)
func (r *Router) Providers() []Provider
func (r *Router) AllEngines() []string
```

- [x] **Step 3: Run tests, verify pass**

- [x] **Step 4: Commit**

```bash
git add pkg/search/router.go pkg/search/router_test.go
git commit -m "feat(search): add Router with provider registry and engine resolution"
```

---

## Sprint 31: Scrape.do Google Search Provider
**Goal:** First real provider using the existing Scrape.do Google Search plugin API.

### Task 31.1: Scrape.do Provider Implementation
**Files:**
- Create: `pkg/search/scrapedo.go`
- Create: `pkg/search/scrapedo_test.go`

- [x] **Step 1: Write failing test with httptest mock**

Test that the provider:
- Sends correct params to `/plugin/google/search` (`token`, `q`, `hl`, `gl`, `start`)
- Parses the JSON response into `SearchResponse`
- Sets `Provider: "scrapedo"`, `Engine: "google"`
- Handles API errors

- [x] **Step 2: Implement ScrapedoProvider**

```go
type ScrapedoProvider struct {
    token   string
    baseURL string
    client  *http.Client
}

func NewScrapedoProvider(token string) *ScrapedoProvider
func (p *ScrapedoProvider) Name() string       // "scrapedo"
func (p *ScrapedoProvider) Engines() []string   // ["google"]
func (p *ScrapedoProvider) Search(ctx, query, opts) (*SearchResponse, error)
```

Key mapping: Scrape.do returns `organic_results` array. Each item has `title`, `link`, `snippet`, `position`. Map to `SearchResult`.

- [x] **Step 3: Run tests, verify pass**

- [x] **Step 4: Commit**

```bash
git add pkg/search/scrapedo.go pkg/search/scrapedo_test.go
git commit -m "feat(search): add Scrape.do Google Search provider"
```

---

## Sprint 32: SerpAPI Multi-Engine Provider
**Goal:** Provider covering Google, Bing, Yandex, DuckDuckGo, Baidu, Yahoo, Naver through single SerpAPI endpoint.

### Task 32.1: SerpAPI Provider Implementation
**Files:**
- Create: `pkg/search/serpapi.go`
- Create: `pkg/search/serpapi_test.go`

- [x] **Step 1: Write failing tests**

Test per-engine query param mapping:
- `engine=google` → `q=`
- `engine=yandex` → `text=`
- `engine=yahoo` → `p=`
- `engine=naver` → `query=`
- Common: `api_key=`, pagination param varies

Test response normalization:
- SerpAPI `organic_results[].title, link, snippet` → `SearchResult`
- `--raw` includes full SerpAPI response

Test error handling:
- Missing API key → clear error message
- API error response → wrapped error

- [x] **Step 2: Implement SerpAPIProvider**

```go
type SerpAPIProvider struct {
    token   string
    baseURL string
    client  *http.Client
    engines []string
}

func NewSerpAPIProvider(token string) *SerpAPIProvider
func (p *SerpAPIProvider) Name() string
func (p *SerpAPIProvider) Engines() []string
func (p *SerpAPIProvider) Search(ctx, query, opts) (*SearchResponse, error)

// Internal: per-engine param mapping
func (p *SerpAPIProvider) queryParam(engine string) string
func (p *SerpAPIProvider) paginationParam(engine string, page int) (string, string)
```

- [x] **Step 3: Run tests, verify pass**

- [x] **Step 4: Commit**

```bash
git add pkg/search/serpapi.go pkg/search/serpapi_test.go
git commit -m "feat(search): add SerpAPI multi-engine provider (google, bing, yandex, ddg, baidu, yahoo, naver)"
```

---

## Sprint 33: Exec Plugin Provider
**Goal:** External binary plugin support via stdin/stdout JSON protocol.

### Task 33.1: Exec Provider Implementation
**Files:**
- Create: `pkg/search/exec.go`
- Create: `pkg/search/exec_test.go`

- [x] **Step 1: Write failing tests**

Test with a mock script (use `os.CreateTemp` + `chmod +x`):
- Script reads JSON stdin, returns valid SearchResponse JSON on stdout
- Timeout handling (30s default)
- Non-zero exit code → error
- Invalid JSON output → parse error
- Missing binary → clear error

- [x] **Step 2: Implement ExecProvider**

```go
type ExecProvider struct {
    name    string
    command string
    engines []string
    timeout time.Duration
}

func NewExecProvider(name, command string, engines []string) *ExecProvider
func (p *ExecProvider) Name() string
func (p *ExecProvider) Engines() []string
func (p *ExecProvider) Search(ctx, query, opts) (*SearchResponse, error)
```

Protocol: write JSON to stdin, read JSON from stdout, timeout via context.

- [x] **Step 3: Run tests, verify pass**

- [x] **Step 4: Commit**

```bash
git add pkg/search/exec.go pkg/search/exec_test.go
git commit -m "feat(search): add exec-based plugin provider (stdin/stdout JSON)"
```

---

## Sprint 34: Output Formatters
**Goal:** Table, JSON, and Markdown output for search results.

### Task 34.1: Formatters
**Files:**
- Create: `pkg/search/format.go`
- Create: `pkg/search/format_test.go`

- [x] **Step 1: Write failing tests**

```go
func TestFormatTable(t *testing.T)    // tabwriter output with position, title, url
func TestFormatJSON(t *testing.T)     // json.MarshalIndent
func TestFormatMarkdown(t *testing.T) // numbered markdown with links and snippets
func TestFormatEmpty(t *testing.T)    // "No results found" message
```

- [x] **Step 2: Implement formatters**

```go
func FormatTable(w io.Writer, resp *SearchResponse) error
func FormatJSON(w io.Writer, resp *SearchResponse) error
func FormatMarkdown(w io.Writer, resp *SearchResponse) error
```

- [x] **Step 3: Run tests, verify pass**

- [x] **Step 4: Commit**

```bash
git add pkg/search/format.go pkg/search/format_test.go
git commit -m "feat(search): add table, JSON, and markdown output formatters"
```

---

## Sprint 35: Config Extension & Provider Initialization
**Goal:** Extend config schema for search providers and wire up initialization.

### Task 35.1: Config Schema Extension
**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

- [x] **Step 1: Write failing tests for new config fields**

Test loading config with `[search]` and `[providers.serpapi]` sections.

- [x] **Step 2: Extend Config struct**

```go
type SearchConfig struct {
    DefaultProvider string `koanf:"default_provider"`
    DefaultEngine   string `koanf:"default_engine"`
    DefaultLimit    int    `koanf:"default_limit"`
}

type ProviderConfig struct {
    Token   string   `koanf:"token"`
    Type    string   `koanf:"type"`    // "" (built-in) or "exec"
    Command string   `koanf:"command"` // for exec type
    Engines []string `koanf:"engines"`
}

// Add to Config:
// Search    SearchConfig                `koanf:"search"`
// Providers map[string]ProviderConfig   `koanf:"providers"`
```

- [x] **Step 3: Add defaults in Load()**

```go
defaults["search.default_provider"] = "scrapedo"
defaults["search.default_engine"] = "google"
defaults["search.default_limit"] = 10
```

- [x] **Step 4: Run tests, verify pass**

- [x] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add search and providers configuration schema"
```

### Task 35.2: Provider Initialization in main.go
**Files:**
- Modify: `cmd/scrapedoctl/main.go`

- [x] **Step 1: Add provider initialization function**

```go
func initSearchRouter(cfg *config.Config) *search.Router {
    router := search.NewRouter()

    // Always register scrapedo if token exists
    if token := cfg.Global.Token; token != "" {
        router.Register(search.NewScrapedoProvider(token))
    }

    // Register configured providers
    for name, pcfg := range cfg.Providers {
        switch {
        case pcfg.Type == "exec":
            router.Register(search.NewExecProvider(name, pcfg.Command, pcfg.Engines))
        case name == "serpapi" && pcfg.Token != "":
            router.Register(search.NewSerpAPIProvider(pcfg.Token))
        }
    }
    return router
}
```

- [x] **Step 2: Wire into PersistentPreRunE, store as package var**

- [x] **Step 3: Commit**

```bash
git add cmd/scrapedoctl/main.go
git commit -m "feat: wire search provider initialization into CLI startup"
```

---

## Sprint 36: CLI Search Command
**Goal:** `scrapedoctl search` command with all flags.

### Task 36.1: Search Command
**Files:**
- Create: `cmd/scrapedoctl/search.go`
- Modify: `cmd/scrapedoctl/main.go` (register command)
- Modify: `cmd/scrapedoctl/integration_test.go` (add tests)

- [x] **Step 1: Write failing integration tests**

```go
func TestSearchCmd_MissingQuery(t *testing.T)
func TestSearchCmd_NoProviders(t *testing.T)
func TestSearchCmd_UnsupportedEngine(t *testing.T)
func TestSearchCmd_JSONOutput(t *testing.T)       // with mock provider
func TestSearchCmd_MarkdownOutput(t *testing.T)
```

- [x] **Step 2: Implement search command**

```go
func newSearchCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "search <query>",
        Short: "Search the web using multiple engines",
        Args:  cobra.MinimumNArgs(1),
        RunE:  runSearch,
    }
    cmd.Flags().String("engine", "", "Search engine")
    cmd.Flags().String("provider", "", "Force specific provider")
    cmd.Flags().String("lang", "", "Language code")
    cmd.Flags().String("country", "", "Country code")
    cmd.Flags().Int("limit", 0, "Max results")
    cmd.Flags().Int("page", 1, "Page number")
    cmd.Flags().Bool("raw", false, "Include raw provider response")
    cmd.Flags().Bool("json", false, "Output as JSON")
    cmd.Flags().Bool("markdown", false, "Output as markdown")
    return cmd
}
```

- [x] **Step 3: Run tests, verify pass**

- [x] **Step 4: Commit**

```bash
git add cmd/scrapedoctl/search.go cmd/scrapedoctl/main.go cmd/scrapedoctl/integration_test.go
git commit -m "feat: add 'search' CLI command with multi-provider support"
```

---

## Sprint 37: MCP Search Tool
**Goal:** Expose search as MCP tool for AI agents.

### Task 37.1: MCP Tool Registration
**Files:**
- Modify: `internal/mcp/server.go`
- Modify: `internal/mcp/server_test.go`

- [x] **Step 1: Write failing test for search tool**

Test that the MCP server has a `search` tool, accepts query/engine/provider params, returns markdown-formatted results.

- [x] **Step 2: Add search tool in addSearchTool()**

Register tool with `name: "search"`, inputSchema matching the spec. Handler calls `router.Resolve()` then `provider.Search()`, returns markdown format.

- [x] **Step 3: Run tests, verify pass**

- [x] **Step 4: Commit**

```bash
git add internal/mcp/server.go internal/mcp/server_test.go
git commit -m "feat(mcp): add search tool for AI agents"
```

---

## Sprint 38: REPL Enhancement — Command Registry
**Goal:** Refactor REPL to support all CLI commands via a command registry pattern.

### Task 38.1: Command Registry
**Files:**
- Create: `internal/repl/commands.go`
- Modify: `internal/repl/repl.go`
- Modify: `internal/repl/repl_test.go`

- [x] **Step 1: Write failing tests for new commands**

```go
func TestREPL_SearchCommand(t *testing.T)
func TestREPL_HistoryCommand(t *testing.T)
func TestREPL_CacheStatsCommand(t *testing.T)
func TestREPL_CacheClearCommand(t *testing.T)
func TestREPL_ConfigListCommand(t *testing.T)
func TestREPL_ConfigSetCommand(t *testing.T)
func TestREPL_ConfigGetCommand(t *testing.T)
func TestREPL_HelpWithSubcommand(t *testing.T)
```

- [x] **Step 2: Implement command registry**

```go
type Command struct {
    Name        string
    Usage       string
    Description string
    Handler     func(ctx context.Context, args []string) error
    Completer   func(prefix string) []string
}

// Shell gets a commands map[string]*Command instead of switch/case
```

- [x] **Step 3: Register all commands**

handlers in `commands.go`:
- `search` → calls search router
- `scrape` → existing logic
- `history` → calls cache.GetHistory
- `cache stats` / `cache clear` → calls cache store
- `config list` / `config get` / `config set` → calls config
- `help` / `exit` / `quit` → built-in

- [x] **Step 4: Update help to list all commands dynamically**

- [x] **Step 5: Run tests, verify pass**

- [x] **Step 6: Commit**

```bash
git add internal/repl/commands.go internal/repl/repl.go internal/repl/repl_test.go
git commit -m "feat(repl): add command registry with all CLI commands"
```

---

## Sprint 39: REPL Tab-Completion
**Goal:** Context-aware tab completion for all commands and parameters.

### Task 39.1: Completer Implementation
**Files:**
- Create: `internal/repl/completer.go`
- Create: `internal/repl/completer_test.go`
- Modify: `internal/repl/repl.go`

- [x] **Step 1: Write failing tests**

```go
func TestCompleter_Commands(t *testing.T)       // "se" → ["search", "scrape"]
func TestCompleter_SearchEngine(t *testing.T)   // "search foo eng" → ["engine=google", "engine=bing", ...]
func TestCompleter_SearchProvider(t *testing.T) // "search foo pro" → ["provider=scrapedo", ...]
func TestCompleter_CacheSubcmd(t *testing.T)    // "cache " → ["stats", "clear"]
func TestCompleter_ConfigSubcmd(t *testing.T)   // "config " → ["list", "get", "set"]
```

- [x] **Step 2: Implement Completer**

```go
type Completer struct {
    commands map[string]*Command
    engines  []string  // from router.AllEngines()
    providers []string // from router.Providers()
}

func (c *Completer) Complete(line string, pos int) []string
```

Wire into `reeflective/readline` via `rl.Completer` callback.

- [x] **Step 3: Run tests, verify pass**

- [x] **Step 4: Commit**

```bash
git add internal/repl/completer.go internal/repl/completer_test.go internal/repl/repl.go
git commit -m "feat(repl): add context-aware tab-completion for all commands"
```

---

## Sprint 40: Final Validation & Documentation
**Goal:** Full test suite, lint, coverage, and updated docs.

### Task 40.1: Validation
- [x] **Step 1: Rebuild container**

```bash
podman build -t scrapedoctl-dev --target builder -f Containerfile .
```

- [x] **Step 2: Run full test suite with coverage**

```bash
podman run --rm -w /src scrapedoctl-dev go test ./... -coverprofile=/tmp/cover.out -covermode=atomic
```

Target: all packages PASS, `pkg/search` >= 90%, `internal/repl` >= 85%

- [x] **Step 3: Run lint**

```bash
podman run --rm -w /src scrapedoctl-dev golangci-lint run ./...
```

Target: 0 issues

- [x] **Step 4: Update metadata command**

Add `search` to the metadata JSON output in `cmd/scrapedoctl/metadata.go`.

- [x] **Step 5: Update help text**

Ensure `scrapedoctl --help` shows `search` command.

- [x] **Step 6: Final commit**

```bash
git commit -m "feat: multi-provider search with REPL enhancement complete"
```
