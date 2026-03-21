package repl

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ioplane/scrapedoctl/internal/cache"
	"github.com/ioplane/scrapedoctl/internal/config"
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
	s.registerCoreCommands()
	s.registerCacheCommands()
	s.registerConfigCommands()
}

func (s *Shell) registerCoreCommands() {
	exitHandler := func(_ context.Context, _ []string) error { return errExit }

	s.commands["search"] = &Command{
		Name:        "search",
		Usage:       "search <query> [engine=X] [provider=Y] [lang=X] [limit=N]",
		Description: "Search the web",
		Handler:     s.handleSearch,
		Completer:   s.completeSearch,
	}
	s.commands["scrape"] = &Command{
		Name:        "scrape",
		Usage:       "scrape <url> [render=true] [super=true] [no-cache=true] [refresh=true]",
		Description: "Scrape a URL",
		Handler:     s.handleScrape,
	}
	s.commands["history"] = &Command{
		Name: "history", Usage: "history <url>",
		Description: "Show scrape history", Handler: s.handleHistory,
	}
	s.commands["help"] = &Command{
		Name: "help", Usage: "help [command]",
		Description: "Show help", Handler: s.handleHelp,
	}
	s.commands["exit"] = &Command{
		Name: "exit", Usage: "exit",
		Description: "Exit REPL", Handler: exitHandler,
	}
	s.commands["quit"] = &Command{
		Name: "quit", Usage: "quit",
		Description: "Exit REPL", Handler: exitHandler,
	}
}

func (s *Shell) registerCacheCommands() {
	s.commands["cache"] = &Command{
		Name:        "cache",
		Usage:       "cache <stats|clear>",
		Description: "Cache management",
		Handler:     s.handleCacheHelp,
		SubCommands: map[string]*Command{
			"stats": {
				Name: "stats", Usage: "cache stats",
				Description: "Show cache statistics", Handler: s.handleCacheStats,
			},
			"clear": {
				Name: "clear", Usage: "cache clear",
				Description: "Clear the cache", Handler: s.handleCacheClear,
			},
		},
	}
}

func (s *Shell) registerConfigCommands() {
	s.commands["config"] = &Command{
		Name:        "config",
		Usage:       "config <list|get|set>",
		Description: "Config management",
		Handler:     s.handleConfigHelp,
		SubCommands: map[string]*Command{
			"list": {
				Name: "list", Usage: "config list",
				Description: "List all settings", Handler: s.handleConfigList,
			},
			"get": {
				Name: "get", Usage: "config get <key>",
				Description: "Get a setting value", Handler: s.handleConfigGet,
			},
			"set": {
				Name: "set", Usage: "config set <key>=<value>",
				Description: "Set a setting value", Handler: s.handleConfigSet,
			},
		},
	}
}

func (s *Shell) handleSearch(ctx context.Context, args []string) error {
	if s.router == nil {
		return errNoRouter
	}

	if len(args) == 0 {
		return fmt.Errorf("%w: search <query> [engine=X] [provider=Y] [lang=X] [limit=N]", errInvalidUsage)
	}

	var queryParts []string
	opts := search.Options{
		Engine: "google",
		Limit:  10,
	}
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
			if n, err := strconv.Atoi(strings.TrimPrefix(arg, "limit=")); err == nil {
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
			var completions []string
			for _, e := range s.router.AllEngines() {
				completions = append(completions, "engine="+e)
			}
			return completions
		}
	case strings.HasPrefix(last, "provider="):
		if s.router != nil {
			var completions []string
			for _, n := range s.router.ProviderNames() {
				completions = append(completions, "provider="+n)
			}
			return completions
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
		// Last arg doesn't look like a parameter prefix — return all params.
		return paramKeywords
	}

	return nil
}

func (s *Shell) handleHistory(ctx context.Context, args []string) error {
	if s.cache == nil {
		return errNoCache
	}

	if len(args) == 0 {
		return fmt.Errorf("%w: history <url>", errInvalidUsage)
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
		fmt.Printf("#%d  %s  %s\n", r.ID, r.CreatedAt.Format("2006-01-02 15:04:05"), r.URL)
	}

	return nil
}

func (s *Shell) handleCacheStats(ctx context.Context, _ []string) error {
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

func (s *Shell) handleCacheClear(ctx context.Context, _ []string) error {
	if s.cache == nil {
		return errNoCache
	}

	if err := s.cache.Clear(ctx); err != nil {
		return fmt.Errorf("cache clear failed: %w", err)
	}

	fmt.Println("Cache cleared successfully")

	return nil
}

func (s *Shell) handleCacheHelp(_ context.Context, _ []string) error {
	fmt.Println("Usage: cache <stats|clear>")

	return nil
}

func (s *Shell) handleConfigList(_ context.Context, _ []string) error {
	if s.config == nil {
		return errNoConfig
	}

	fmt.Printf("Global Token: %s\n", s.config.Global.Token)
	fmt.Printf("Global BaseURL: %s\n", s.config.Global.BaseURL)
	fmt.Printf("Global Timeout: %d\n", s.config.Global.Timeout)
	fmt.Printf("REPL History: %s\n", s.config.Repl.HistoryFile)
	fmt.Printf("Active Profile: %s\n", s.config.ActiveProfile)

	return nil
}

func (s *Shell) handleConfigGet(_ context.Context, args []string) error {
	if s.config == nil {
		return errNoConfig
	}

	if len(args) == 0 {
		return fmt.Errorf("%w: config get <key>", errInvalidUsage)
	}

	key := args[0]

	switch key {
	case "global.token":
		fmt.Println(s.config.Global.Token)
	case "global.base_url":
		fmt.Println(s.config.Global.BaseURL)
	case "global.timeout":
		fmt.Println(s.config.Global.Timeout)
	case "repl.history_file":
		fmt.Println(s.config.Repl.HistoryFile)
	default:
		return fmt.Errorf("%w: %s", errUnsupportedKey, key)
	}

	return nil
}

func (s *Shell) handleConfigSet(_ context.Context, args []string) error {
	if s.config == nil {
		return errNoConfig
	}

	if len(args) == 0 {
		return fmt.Errorf("%w: config set <key>=<value>", errInvalidUsage)
	}

	parts := strings.SplitN(args[0], "=", 2)
	if len(parts) != 2 {
		return errInvalidFormat
	}

	key, value := parts[0], parts[1]

	switch key {
	case "global.token":
		s.config.Global.Token = value
	case "global.base_url":
		s.config.Global.BaseURL = value
	case "repl.history_file":
		s.config.Repl.HistoryFile = value
	default:
		return fmt.Errorf("%w: %s", errUnsupportedKey, key)
	}

	if err := s.config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Successfully set %s\n", key)

	return nil
}

func (s *Shell) handleConfigHelp(_ context.Context, _ []string) error {
	fmt.Println("Usage: config <list|get|set>")

	return nil
}

func (s *Shell) handleHelp(_ context.Context, args []string) error {
	if len(args) > 0 {
		cmd, ok := s.commands[args[0]]
		if !ok {
			return fmt.Errorf("%w: %s", errUnknownCmd, args[0])
		}

		fmt.Printf("  %s  - %s\n", cmd.Usage, cmd.Description)

		if cmd.SubCommands != nil {
			fmt.Println("\nSubcommands:")
			for _, sub := range cmd.SubCommands {
				fmt.Printf("  %s  - %s\n", sub.Usage, sub.Description)
			}
		}

		return nil
	}

	s.printHelp()

	return nil
}
