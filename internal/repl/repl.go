// Package repl provides an interactive shell for the scrapedoctl CLI.
package repl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/reeflective/readline"

	"github.com/ioplane/scrapedoctl/internal/config"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
	"github.com/ioplane/scrapedoctl/pkg/search"
)

// Reader is an interface for reading lines of input.
type Reader interface {
	Readline() (string, error)
}

// Shell implements an interactive CLI for Scrape.do.
type Shell struct {
	client   *scrapedo.Client
	reader   Reader
	commands map[string]*Command
	router   *search.Router
	cache    CacheStore
	config   *config.Config
}

var (
	errExit           = errors.New("exit")
	errUnknownCmd     = errors.New("unknown command")
	errInvalidUsage   = errors.New("invalid usage")
	errNoRouter       = errors.New("search router not configured")
	errNoCache        = errors.New("cache not configured")
	errNoConfig       = errors.New("config not available")
	errUnsupportedKey = errors.New("unknown or unsupported key")
	errInvalidFormat  = errors.New("invalid format, use key=value")
)

// NewShell creates a new REPL shell with the given Scrape.do client.
func NewShell(client *scrapedo.Client, opts ...ShellOption) *Shell {
	s := &Shell{client: client}

	for _, opt := range opts {
		opt(s)
	}

	s.registerCommands()

	return s
}

// SetReader allows setting a custom reader for the REPL (primarily for testing).
func (s *Shell) SetReader(r Reader) {
	s.reader = r
}

// Run starts the interactive REPL loop.
func (s *Shell) Run(ctx context.Context) error {
	if s.reader == nil {
		rl := readline.NewShell()
		rl.Prompt.Primary(func() string { return "scrapedoctl> " })

		completer := NewCompleter(s.commands)
		rl.Completer = completer.Complete

		s.reader = rl
	}

	fmt.Println("Scrape.do Interactive REPL. Type 'help' for commands, 'exit' to quit.")

	for {
		line, err := s.reader.Readline()
		if err != nil {
			return fmt.Errorf("readline failed: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if err := s.ExecuteCommand(ctx, line); err != nil {
			if errors.Is(err, errExit) {
				return nil
			}
			fmt.Printf("Error: %v\n", err)
		}
	}
}

// ExecuteCommand parses and runs a single command string in the REPL.
func (s *Shell) ExecuteCommand(ctx context.Context, line string) error {
	args := strings.Fields(line)
	if len(args) == 0 {
		return nil
	}

	cmd, ok := s.commands[args[0]]
	if !ok {
		return fmt.Errorf("%w: %s", errUnknownCmd, args[0])
	}

	// Check for subcommand.
	if len(args) > 1 && cmd.SubCommands != nil {
		if sub, ok := cmd.SubCommands[args[1]]; ok {
			return sub.Handler(ctx, args[2:])
		}
	}

	return cmd.Handler(ctx, args[1:])
}

func (s *Shell) printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  search <query> [engine=X] [provider=Y] [lang=X] [limit=N]  - Search the web")
	fmt.Println("  scrape <url> [render=true] [super=true]                     - Scrape a URL")
	fmt.Println("  history <url>                                               - Show scrape history")
	fmt.Println("  cache stats                                                 - Show cache statistics")
	fmt.Println("  cache clear                                                 - Clear the cache")
	fmt.Println("  config list                                                 - List all settings")
	fmt.Println("  config get <key>                                            - Get a setting value")
	fmt.Println("  config set <key>=<value>                                    - Set a setting value")
	fmt.Println("  help [command]                                              - Show help")
	fmt.Println("  exit, quit                                                  - Exit REPL")
}

func (s *Shell) handleScrape(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("%w: scrape <url> [render=true] [super=true]", errInvalidUsage)
	}

	url := args[0]
	req := scrapedo.ScrapeRequest{URL: url}

	// Simple parameter parsing.
	for _, arg := range args[1:] {
		switch arg {
		case "render=true":
			req.Render = true
		case "super=true":
			req.Super = true
		case "no-cache=true":
			req.NoCache = true
		case "refresh=true":
			req.Refresh = true
		}
	}

	result, err := s.client.Scrape(ctx, req)
	if err != nil {
		return fmt.Errorf("scrape failed: %w", err)
	}

	fmt.Println(result)

	return nil
}
