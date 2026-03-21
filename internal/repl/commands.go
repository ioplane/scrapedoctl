package repl

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ioplane/scrapedoctl/internal/cache"
	"github.com/ioplane/scrapedoctl/internal/config"
	"github.com/ioplane/scrapedoctl/internal/version"
	"github.com/ioplane/scrapedoctl/pkg/search"
)

// Command represents a REPL command.
type Command struct {
	Name        string
	Usage       string
	Description string
	Handler     func(ctx context.Context, args []string) error
	SubCommands map[string]*Command
	// Completer returns completions for the given partial arg.
	Completer func(args []string) []string
}

// CacheStore is the interface for cache operations used by the REPL.
type CacheStore interface {
	GetHistory(ctx context.Context, url string) ([]cache.ScrapeRecord, error)
	GetStats(ctx context.Context) (cache.Stats, error)
	Clear(ctx context.Context) error
	GetUsageSummary(ctx context.Context, since time.Time) ([]cache.UsageSummary, error)
}

// ShellOption configures a Shell.
type ShellOption func(*Shell)

// WithSearchRouter sets the search router on the shell.
func WithSearchRouter(r *search.Router) ShellOption {
	return func(s *Shell) {
		s.router = r
	}
}

// WithCache sets the cache store on the shell.
func WithCache(c CacheStore) ShellOption {
	return func(s *Shell) {
		s.cache = c
	}
}

// WithConfig sets the config on the shell.
func WithConfig(c *config.Config) ShellOption {
	return func(s *Shell) {
		s.config = c
	}
}

func (s *Shell) registerCommands() {
	s.commands = make(map[string]*Command)
	s.registerShowCommands()
	s.registerActionCommands()
	s.registerSessionCommands()
}

func (s *Shell) registerShowCommands() {
	s.commands["show"] = &Command{
		Name:        "show",
		Usage:       "show <account|config|cache|history|usage|version>",
		Description: "Show system information",
		Handler:     s.handleShowHelp,
		SubCommands: map[string]*Command{
			"account": {
				Name: "account", Usage: "show account",
				Description: "Show provider account usage", Handler: s.handleShowAccount,
			},
			"config": {
				Name: "config", Usage: "show config [key]",
				Description: "Show configuration", Handler: s.handleShowConfig,
			},
			"cache": {
				Name: "cache", Usage: "show cache",
				Description: "Show cache statistics", Handler: s.handleShowCache,
			},
			"history": {
				Name: "history", Usage: "show history <url>",
				Description: "Show scrape history", Handler: s.handleShowHistory,
			},
			"usage": {
				Name: "usage", Usage: "show usage [--week|--month|--all]",
				Description: "Show API usage statistics", Handler: s.handleShowUsage,
			},
			"version": {
				Name: "version", Usage: "show version",
				Description: "Show version info", Handler: s.handleShowVersion,
			},
		},
	}
}

func (s *Shell) registerActionCommands() {
	s.commands["set"] = &Command{
		Name: "set", Usage: "set <key> <value>",
		Description: "Set a configuration value", Handler: s.handleSet,
	}

	s.commands["clear"] = &Command{
		Name: "clear", Usage: "clear <cache>",
		Description: "Clear data", Handler: s.handleClearHelp,
		SubCommands: map[string]*Command{
			"cache": {
				Name: "cache", Usage: "clear cache",
				Description: "Clear the persistent cache", Handler: s.handleClearCache,
			},
		},
	}

	s.commands["search"] = &Command{
		Name: "search", Description: "Search the web",
		Usage:     "search <query> [engine=X] [provider=Y] [lang=X] [limit=N]",
		Handler:   s.handleSearch,
		Completer: s.completeSearch,
	}

	s.commands["scrape"] = &Command{
		Name: "scrape", Usage: "scrape <url> [render=true] [super=true]",
		Description: "Scrape a URL", Handler: s.handleScrape,
	}

	s.commands["map"] = &Command{
		Name: "map", Usage: "map <url> [search=keyword] [limit=N]",
		Description: "Discover same-domain URLs on a page", Handler: s.handleMap,
	}

	s.commands["crawl"] = &Command{
		Name: "crawl", Usage: "crawl <url> [depth=N] [limit=N]",
		Description: "Recursively crawl a site", Handler: s.handleCrawl,
	}
}

func (s *Shell) registerSessionCommands() {
	exitHandler := func(_ context.Context, _ []string) error { return errExit }

	s.commands["exit"] = &Command{
		Name: "exit", Usage: "exit",
		Description: "Exit REPL", Handler: exitHandler,
	}
	s.commands["quit"] = &Command{
		Name: "quit", Usage: "quit",
		Description: "Exit REPL", Handler: exitHandler,
	}
	s.commands["help"] = &Command{
		Name: "help", Usage: "help [command]",
		Description: "Show help (alias for '?')", Handler: s.handleHelp,
	}
}

// ── show handlers ──

func (s *Shell) handleShowConfig(_ context.Context, args []string) error {
	if s.config == nil {
		return errNoConfig
	}

	if len(args) > 0 {
		return s.showConfigKey(args[0])
	}

	fmt.Printf("global.token:        %s\n", maskToken(s.config.Global.Token))
	fmt.Printf("global.base_url:     %s\n", s.config.Global.BaseURL)
	fmt.Printf("global.timeout:      %d\n", s.config.Global.Timeout)
	fmt.Printf("repl.history_file:   %s\n", s.config.Repl.HistoryFile)
	fmt.Printf("search.default_engine:   %s\n", s.config.Search.DefaultEngine)
	fmt.Printf("search.default_provider: %s\n", s.config.Search.DefaultProvider)
	fmt.Printf("search.default_limit:    %d\n", s.config.Search.DefaultLimit)

	return nil
}

func (s *Shell) showConfigKey(key string) error {
	switch key {
	case "global.token":
		fmt.Println(maskToken(s.config.Global.Token))
	case "global.base_url":
		fmt.Println(s.config.Global.BaseURL)
	case "global.timeout":
		fmt.Println(s.config.Global.Timeout)
	case "repl.history_file":
		fmt.Println(s.config.Repl.HistoryFile)
	case "search.default_engine":
		fmt.Println(s.config.Search.DefaultEngine)
	case "search.default_provider":
		fmt.Println(s.config.Search.DefaultProvider)
	case "search.default_limit":
		fmt.Println(s.config.Search.DefaultLimit)
	default:
		return fmt.Errorf("%w: %s", errUnsupportedKey, key)
	}

	return nil
}

func (s *Shell) handleShowCache(ctx context.Context, _ []string) error {
	if s.cache == nil {
		return errNoCache
	}

	stats, err := s.cache.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("cache stats failed: %w", err)
	}

	fmt.Printf("Total entries: %d\n", stats.TotalCount)
	fmt.Printf("Total size:    %.2f MB\n", float64(stats.TotalSize)/(1024*1024))

	return nil
}

func (s *Shell) handleShowHistory(ctx context.Context, args []string) error {
	if s.cache == nil {
		return errNoCache
	}

	if len(args) == 0 {
		return fmt.Errorf("%w: show history <url>", errInvalidUsage)
	}

	records, err := s.cache.GetHistory(ctx, args[0])
	if err != nil {
		return fmt.Errorf("history failed: %w", err)
	}

	if len(records) == 0 {
		fmt.Printf("No history found for %s\n", args[0])
		return nil
	}

	for _, r := range records {
		fmt.Printf(
			"#%d  %s  %s\n",
			r.ID, r.CreatedAt.Format("2006-01-02 15:04:05"), r.URL,
		)
	}

	return nil
}

func (s *Shell) handleShowVersion(ctx context.Context, _ []string) error {
	fmt.Println(version.Info())

	tag, url, newer, err := version.CheckLatest(ctx)
	if err != nil {
		fmt.Printf("Update check failed: %v\n", err)
		return nil
	}

	if newer {
		fmt.Printf("New version available: %s — %s\n", tag, url)
	} else {
		fmt.Printf("Up to date (%s)\n", tag)
	}

	return nil
}

func (s *Shell) handleShowAccount(ctx context.Context, _ []string) error {
	if s.router == nil {
		return errNoRouter
	}

	var found bool

	for _, p := range s.router.Providers() {
		checker, ok := p.(search.AccountChecker)
		if !ok {
			continue
		}

		info, err := checker.Account(ctx)
		if err != nil {
			fmt.Printf("%s: error: %v\n", p.Name(), err)
			continue
		}

		found = true
		s.printAccountInfo(info)
	}

	if !found {
		fmt.Println("No providers support account info.")
	}

	return nil
}

func (s *Shell) printAccountInfo(info *search.AccountInfo) {
	fmt.Printf("%-12s used=%d limit=%d remaining=%d",
		info.Provider, info.UsedRequests, info.MaxRequests, info.RemainingRequests)

	if info.Plan != "" {
		fmt.Printf(" plan=%s", info.Plan)
	}

	if info.Concurrency > 0 {
		fmt.Printf(" concurrency=%d", info.Concurrency)
	}

	if info.RateLimit > 0 {
		fmt.Printf(" rate_limit=%d/h", info.RateLimit)
	}

	fmt.Println()
}

func (s *Shell) handleShowUsage(ctx context.Context, args []string) error {
	if s.cache == nil {
		return errNoCache
	}

	since := usageSinceFromArgs(args)
	summary, err := s.cache.GetUsageSummary(ctx, since)
	if err != nil {
		return fmt.Errorf("usage query failed: %w", err)
	}

	sinceStr := "all time"
	if !since.IsZero() {
		sinceStr = since.Format("2006-01-02")
	}

	fmt.Printf("Usage since %s:\n\n", sinceStr)
	fmt.Printf("%-13s%-9s%-8s%s\n", "Provider", "Action", "Count", "Credits")

	var totalCount, totalCredits int64
	for _, row := range summary {
		totalCount += row.Count
		totalCredits += row.TotalCredits
		fmt.Printf("%-13s%-9s%-8d%d\n", row.Provider, row.Action, row.Count, row.TotalCredits)
	}

	fmt.Printf("\nTotal: %d requests, %d credits\n", totalCount, totalCredits)
	return nil
}

func usageSinceFromArgs(args []string) time.Time {
	now := time.Now()

	for _, arg := range args {
		switch arg {
		case "--all":
			return time.Time{}
		case "--month":
			return now.AddDate(0, 0, -30) //nolint:mnd // 30 days
		case "--week":
			return now.AddDate(0, 0, -7) //nolint:mnd // 7 days
		}
	}

	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
}

func (s *Shell) handleShowHelp(_ context.Context, _ []string) error {
	fmt.Println("Usage: show <account|config|cache|history|usage|version>")
	return nil
}

// ── set handler ──

func (s *Shell) handleSet(_ context.Context, args []string) error {
	if s.config == nil {
		return errNoConfig
	}

	if len(args) < 2 { //nolint:mnd // need key + value.
		return fmt.Errorf("%w: set <key> <value>", errInvalidUsage)
	}

	key := args[0]
	value := strings.Join(args[1:], " ")

	switch key {
	case "global.token":
		s.config.Global.Token = value
	case "global.base_url":
		s.config.Global.BaseURL = value
	case "repl.history_file":
		s.config.Repl.HistoryFile = value
	case "search.default_engine":
		s.config.Search.DefaultEngine = value
	case "search.default_provider":
		s.config.Search.DefaultProvider = value
	case "search.default_limit":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer: %w", err)
		}
		s.config.Search.DefaultLimit = n
	default:
		return fmt.Errorf("%w: %s", errUnsupportedKey, key)
	}

	if err := s.config.Save(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("%s = %s\n", key, value)

	return nil
}

// ── clear handlers ──

func (s *Shell) handleClearCache(ctx context.Context, _ []string) error {
	if s.cache == nil {
		return errNoCache
	}

	if err := s.cache.Clear(ctx); err != nil {
		return fmt.Errorf("clear cache: %w", err)
	}

	fmt.Println("Cache cleared")

	return nil
}

func (s *Shell) handleClearHelp(_ context.Context, _ []string) error {
	fmt.Println("Usage: clear <cache>")
	return nil
}

// ── search handler ──

func (s *Shell) handleSearch(ctx context.Context, args []string) error {
	if s.router == nil {
		return errNoRouter
	}

	if len(args) == 0 {
		return fmt.Errorf(
			"%w: search <query> [engine=X] [provider=Y]",
			errInvalidUsage,
		)
	}

	var queryParts []string

	opts := search.Options{Engine: "google", Limit: 10}
	var providerName string

	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "engine="):
			opts.Engine = strings.TrimPrefix(arg, "engine=")
		case strings.HasPrefix(arg, "provider="):
			providerName = strings.TrimPrefix(arg, "provider=")
		case strings.HasPrefix(arg, "lang="):
			opts.Lang = strings.TrimPrefix(arg, "lang=")
		case strings.HasPrefix(arg, "limit="):
			if n, err := strconv.Atoi(
				strings.TrimPrefix(arg, "limit="),
			); err == nil {
				opts.Limit = n
			}
		default:
			queryParts = append(queryParts, arg)
		}
	}

	query := strings.Join(queryParts, " ")
	if query == "" {
		return fmt.Errorf("%w: search requires a query", errInvalidUsage)
	}

	p, err := s.router.Resolve(opts.Engine, providerName)
	if err != nil {
		return fmt.Errorf("resolve provider: %w", err)
	}

	resp, err := p.Search(ctx, query, opts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if err := search.FormatTable(os.Stdout, resp); err != nil {
		return fmt.Errorf("format results: %w", err)
	}

	return nil
}

func (s *Shell) completeSearch(args []string) []string {
	paramKeywords := []string{"engine=", "provider=", "lang=", "limit="}

	if len(args) == 0 {
		return paramKeywords
	}

	last := args[len(args)-1]

	switch {
	case strings.HasPrefix(last, "engine="):
		if s.router != nil {
			var c []string
			for _, e := range s.router.AllEngines() {
				c = append(c, "engine="+e)
			}
			return c
		}
	case strings.HasPrefix(last, "provider="):
		if s.router != nil {
			var c []string
			for _, n := range s.router.ProviderNames() {
				c = append(c, "provider="+n)
			}
			return c
		}
	case strings.HasPrefix(last, "eng"):
		return []string{"engine="}
	case strings.HasPrefix(last, "pro"):
		return []string{"provider="}
	case strings.HasPrefix(last, "lan"):
		return []string{"lang="}
	case strings.HasPrefix(last, "lim"):
		return []string{"limit="}
	default:
		return paramKeywords
	}

	return nil
}

// ── help handler ──

func (s *Shell) handleHelp(_ context.Context, args []string) error {
	if len(args) > 0 {
		cmd, err := s.resolveCommand(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("  %s  — %s\n", cmd.Usage, cmd.Description)

		if cmd.SubCommands != nil {
			fmt.Println("\nSubcommands:")
			for _, sub := range cmd.SubCommands {
				fmt.Printf("  %-24s %s\n", sub.Usage, sub.Description)
			}
		}

		return nil
	}

	s.printHelp()

	return nil
}

// maskToken shows only first 4 chars of a token.
func maskToken(token string) string {
	if len(token) <= 4 { //nolint:mnd // keep 4 visible chars.
		return "****"
	}

	return token[:4] + "****"
}
