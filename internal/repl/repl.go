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

// Shell implements an interactive CLI for Scrape.do.
type Shell struct {
	client *scrapedo.Client
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

// Run starts the interactive REPL loop.
func (s *Shell) Run(ctx context.Context) error {
	rl := readline.NewShell()
	rl.Prompt.Primary(func() string { return "scrapedoctl> " })

	fmt.Println("Scrape.do Interactive REPL. Type 'help' for commands, 'exit' to quit.")

	for {
		line, err := rl.Readline()
		if err != nil {
			return fmt.Errorf("readline failed: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if err := s.executeCommand(ctx, line); err != nil {
			if errors.Is(err, errExit) {
				return nil
			}
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func (s *Shell) executeCommand(ctx context.Context, line string) error {
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
