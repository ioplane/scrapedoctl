package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// simpleReader is a basic line reader using bufio.Scanner.
// It provides a reliable fallback when reeflective/readline
// causes issues with DSR (Device Status Report) in certain terminals.
// Enable via: SCRAPEDOCTL_SIMPLE_REPL=1 scrapedoctl repl.
type simpleReader struct {
	scanner *bufio.Scanner
	prompt  string
}

func newSimpleReader(prompt string) *simpleReader {
	return &simpleReader{
		scanner: bufio.NewScanner(os.Stdin),
		prompt:  prompt,
	}
}

// Readline prints the prompt and reads one line from stdin.
func (r *simpleReader) Readline() (string, error) {
	fmt.Fprint(os.Stdout, r.prompt)

	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return "", fmt.Errorf("read input: %w", err)
		}

		return "", io.EOF
	}

	return r.scanner.Text(), nil
}
