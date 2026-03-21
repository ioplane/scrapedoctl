// Package repl provides an interactive shell for the scrapedoctl CLI.
package repl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/reeflective/readline"

	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

// Reader is an interface for reading lines of input.
type Reader interface {
	Readline() (string, error)
}

// Shell implements an interactive CLI for Scrape.do.
type Shell struct {
	client *scrapedo.Client
	reader Reader
}

var (
	errExit         = errors.New("exit")
	errUnknownCmd   = errors.New("unknown command")
	errInvalidUsage = errors.New("usage: scrape <url> [render=true] [super=true]")
)

// NewShell creates a new REPL shell with the given Scrape.do client.
func NewShell(client *scrapedo.Client) *Shell {
	return &Shell{client: client}
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

func (s *Shell) ExecuteCommand(ctx context.Context, line string) error {
	args := strings.Fields(line)
	cmd := args[0]

	switch cmd {
	case "exit", "quit":
		return errExit
	case "help":
		s.printHelp()
		return nil
	case "scrape":
		return s.handleScrape(ctx, args)
	default:
		return fmt.Errorf("%w: %s", errUnknownCmd, cmd)
	}
}

func (s *Shell) printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  scrape <url> [render=true] [super=true] - Scrape a URL")
	fmt.Println("  exit, quit                              - Exit REPL")
}

func (s *Shell) handleScrape(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return errInvalidUsage
	}
	url := args[1]
	req := scrapedo.ScrapeRequest{URL: url}

	// Simple parameter parsing
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "render=true":
			req.Render = true
		case "super=true":
			req.Super = true
		}
	}

	result, err := s.client.Scrape(ctx, req)
	if err != nil {
		return fmt.Errorf("scrape failed: %w", err)
	}
	fmt.Println(result)
	return nil
}
