package repl

import (
	"strings"

	"github.com/reeflective/readline"
)

// Completer provides context-aware tab completion for the REPL.
type Completer struct {
	commands map[string]*Command
}

// NewCompleter creates a new Completer from the command registry.
func NewCompleter(commands map[string]*Command) *Completer {
	return &Completer{commands: commands}
}

// Complete returns completion candidates for the given line and cursor position.
func (c *Completer) Complete(line []rune, cursor int) readline.Completions {
	input := string(line[:cursor])
	fields := strings.Fields(input)
	trailingSpace := len(input) > 0 && input[len(input)-1] == ' '

	// No input or partial first word — complete command names.
	if len(fields) == 0 || (len(fields) == 1 && !trailingSpace) {
		return c.completeCommandName(fields)
	}

	cmdName := fields[0]
	cmd, ok := c.commands[cmdName]
	if !ok {
		return readline.Completions{}
	}

	// Subcommand completion.
	if result, handled := c.completeSubCommand(cmd, fields, trailingSpace); handled {
		return result
	}

	// Custom completer on the command.
	if cmd.Completer != nil {
		return c.completeWithHandler(cmd, fields, trailingSpace)
	}

	// "help" completes with command names.
	if cmdName == "help" {
		return c.completeHelpArg(fields, trailingSpace)
	}

	return readline.Completions{}
}

func (c *Completer) completeCommandName(fields []string) readline.Completions {
	prefix := ""
	if len(fields) == 1 {
		prefix = fields[0]
	}

	return readline.CompleteValues(c.matchCommands(prefix)...)
}

func (c *Completer) completeSubCommand(
	cmd *Command, fields []string, trailingSpace bool,
) (readline.Completions, bool) {
	if cmd.SubCommands == nil {
		return readline.Completions{}, false
	}

	if len(fields) == 1 && trailingSpace {
		return readline.CompleteValues(c.subCommandNames(cmd)...), true
	}

	if len(fields) == 2 && !trailingSpace {
		return readline.CompleteValues(c.matchSubCommands(cmd, fields[1])...), true
	}

	return readline.Completions{}, false
}

func (c *Completer) completeWithHandler(
	cmd *Command, fields []string, trailingSpace bool,
) readline.Completions {
	argsList := fields[1:]
	candidates := cmd.Completer(argsList)

	if !trailingSpace && len(fields) > 1 {
		partial := fields[len(fields)-1]
		candidates = filterByPrefix(candidates, partial)
	}

	return readline.CompleteValues(candidates...)
}

func (c *Completer) completeHelpArg(fields []string, trailingSpace bool) readline.Completions {
	if len(fields) == 1 && trailingSpace {
		return readline.CompleteValues(c.matchCommands("")...)
	}

	if len(fields) == 2 && !trailingSpace {
		return readline.CompleteValues(c.matchCommands(fields[1])...)
	}

	return readline.Completions{}
}

func filterByPrefix(candidates []string, prefix string) []string {
	var filtered []string

	for _, cand := range candidates {
		if strings.HasPrefix(cand, prefix) {
			filtered = append(filtered, cand)
		}
	}

	return filtered
}

func (c *Completer) matchCommands(prefix string) []string {
	var matches []string

	for name := range c.commands {
		if name == "quit" {
			continue // Don't duplicate exit/quit in completions.
		}

		if strings.HasPrefix(name, prefix) {
			matches = append(matches, name)
		}
	}

	return matches
}

func (c *Completer) subCommandNames(cmd *Command) []string {
	names := make([]string, 0, len(cmd.SubCommands))

	for name := range cmd.SubCommands {
		names = append(names, name)
	}

	return names
}

func (c *Completer) matchSubCommands(cmd *Command, prefix string) []string {
	var matches []string

	for name := range cmd.SubCommands {
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, name)
		}
	}

	return matches
}
