// Package repl provides an interactive shell for the scrapedoctl CLI.
package repl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
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

// Printer is an interface for outputting text in the REPL.
// In interactive mode, this wraps readline.Shell.Printf for
// proper prompt redraw. In tests, it writes to a buffer.
type Printer interface {
	Printf(format string, args ...any) (int, error)
}

// printerWriter adapts a Printer to io.Writer for use with tabwriter, etc.
type printerWriter struct {
	p Printer
}

func (w *printerWriter) Write(data []byte) (int, error) {
	return w.p.Printf("%s", data)
}

// Shell implements an interactive CLI for Scrape.do.
type Shell struct {
	client   *scrapedo.Client
	reader   Reader
	printer  Printer
	out      io.Writer // wraps printer as io.Writer
	commands map[string]*Command
	router   *search.Router
	cache    CacheStore
	config   *config.Config
}

// stdPrinter writes to stdout (used when no readline is available, e.g. tests).
type stdPrinter struct{}

func (p *stdPrinter) Printf(format string, args ...any) (int, error) {
	return fmt.Fprintf(os.Stdout, format, args...)
}

var (
	errExit           = errors.New("exit")
	errUnknownCmd     = errors.New("unknown command")
	errAmbiguousCmd   = errors.New("ambiguous command")
	errInvalidUsage   = errors.New("invalid usage")
	errNoRouter       = errors.New("search router not configured")
	errNoCache        = errors.New("cache not configured")
	errNoConfig       = errors.New("config not available")
	errUnsupportedKey = errors.New("unknown or unsupported key")
	errInvalidFormat  = errors.New("invalid format, use: set <key> <value>")
)

// NewShell creates a new REPL shell with the given Scrape.do client.
func NewShell(client *scrapedo.Client, opts ...ShellOption) *Shell {
	p := &stdPrinter{}
	s := &Shell{
		client:  client,
		printer: p,
		out:     &printerWriter{p},
	}

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
		// Use readline's Printf for proper prompt redraw after output.
		s.printer = rl
		s.out = &printerWriter{rl}
	}

	s.printer.Printf("Scrape.do Interactive REPL. Type '?' for commands, 'exit' to quit.\n")

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
			s.printer.Printf("%% %v\n", err)
		}
	}
}

// ExecuteCommand parses and runs a single command string in the REPL.
// Supports Cisco-style prefix matching: "sh con" → "show config".
func (s *Shell) ExecuteCommand(ctx context.Context, line string) error {
	// Handle "?" anywhere — context help.
	if before, found := strings.CutSuffix(line, "?"); found {
		s.contextHelp(before)
		return nil
	}

	args := strings.Fields(line)
	if len(args) == 0 {
		return nil
	}

	cmd, err := s.resolveCommand(args[0])
	if err != nil {
		return err
	}

	// Check for subcommand with prefix matching.
	if len(args) > 1 && cmd.SubCommands != nil {
		sub, subErr := resolveSubcommand(cmd, args[1])
		if subErr == nil {
			return sub.Handler(ctx, args[2:])
		}
		// If ambiguous, report it. If not found, fall through to parent handler.
		if errors.Is(subErr, errAmbiguousCmd) {
			return subErr
		}
	}

	return cmd.Handler(ctx, args[1:])
}

// resolveCommand finds a command by exact or prefix match.
func (s *Shell) resolveCommand(input string) (*Command, error) {
	// Exact match first.
	if cmd, ok := s.commands[input]; ok {
		return cmd, nil
	}

	// Prefix match.
	var matches []*Command
	var matchNames []string

	for name, cmd := range s.commands {
		if strings.HasPrefix(name, input) {
			matches = append(matches, cmd)
			matchNames = append(matchNames, name)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("%w: %s", errUnknownCmd, input)
	case 1:
		return matches[0], nil
	default:
		return nil, fmt.Errorf(
			"%w: %q matches: %s",
			errAmbiguousCmd, input, strings.Join(matchNames, ", "),
		)
	}
}

// resolveSubcommand finds a subcommand by exact or prefix match.
func resolveSubcommand(cmd *Command, input string) (*Command, error) {
	if sub, ok := cmd.SubCommands[input]; ok {
		return sub, nil
	}

	var matches []*Command
	var matchNames []string

	for name, sub := range cmd.SubCommands {
		if strings.HasPrefix(name, input) {
			matches = append(matches, sub)
			matchNames = append(matchNames, name)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("%w: %s", errUnknownCmd, input)
	case 1:
		return matches[0], nil
	default:
		return nil, fmt.Errorf(
			"%w: %q matches: %s",
			errAmbiguousCmd, input, strings.Join(matchNames, ", "),
		)
	}
}

// contextHelp shows available completions for the given partial input.
func (s *Shell) contextHelp(partial string) {
	partial = strings.TrimSpace(partial)

	// Top-level "?" — show all commands.
	if partial == "" {
		s.printHelp()
		return
	}

	args := strings.Fields(partial)

	// Try to resolve the command.
	cmd, err := s.resolveCommand(args[0])
	if err != nil {
		// Show matching commands.
		for name, c := range s.commands {
			if strings.HasPrefix(name, args[0]) {
				s.printer.Printf("  %-12s %s\n", name, c.Description)
			}
		}
		return
	}

	// Command resolved — show subcommands or usage.
	if cmd.SubCommands != nil && len(args) <= 1 {
		for name, sub := range cmd.SubCommands {
			s.printer.Printf("  %-12s %s\n", name, sub.Description)
		}
		return
	}

	s.printer.Printf("  %s\n", cmd.Usage)
}

func (s *Shell) printHelp() {
	s.printer.Printf("Commands:\n")
	s.printer.Printf("  show account          Show provider account usage\n")
	s.printer.Printf("  show config [key]     Show configuration (or a specific key)\n")
	s.printer.Printf("  show cache            Show cache statistics\n")
	s.printer.Printf("  show history <url>    Show scrape history for URL\n")
	s.printer.Printf("  show usage [--week|--month|--all]\n")
	s.printer.Printf("                        Show API usage statistics\n")
	s.printer.Printf("  show version          Show version and check for updates\n")
	s.printer.Printf("  set <key> <value>     Set a configuration value\n")
	s.printer.Printf("  clear cache           Clear the persistent cache\n")
	s.printer.Printf("  search <query> [engine=X] [provider=Y] [lang=X] [limit=N]\n")
	s.printer.Printf("                        Search the web\n")
	s.printer.Printf("  scrape <url> [render=true] [super=true]\n")
	s.printer.Printf("                        Scrape a URL and output markdown\n")
	s.printer.Printf("  map <url> [search=keyword] [limit=N]\n")
	s.printer.Printf("                        Discover same-domain URLs on a page\n")
	s.printer.Printf("  crawl <url> [depth=N] [limit=N]\n")
	s.printer.Printf("                        Recursively crawl a site\n")
	s.printer.Printf("  exit, quit            Exit REPL\n")
	s.printer.Printf("\n")
	s.printer.Printf("Abbreviations: 'sh con' = 'show config', 'se golang' = 'search golang'\n")
	s.printer.Printf("Context help:  type '?' after any partial command\n")
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

	s.printer.Printf("%s\n", result)

	return nil
}
