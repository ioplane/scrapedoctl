package repl

import "github.com/reeflective/readline"

var (
	ErrExit           = errExit
	ErrUnknownCmd     = errUnknownCmd
	ErrAmbiguousCmd   = errAmbiguousCmd
	ErrInvalidUsage   = errInvalidUsage
	ErrNoRouter       = errNoRouter
	ErrNoCache        = errNoCache
	ErrNoConfig       = errNoConfig
	ErrUnsupportedKey = errUnsupportedKey
	ErrInvalidFormat  = errInvalidFormat
)

// NewCompleterFromShell creates a test helper wrapping the shell's completer.
func NewCompleterFromShell(s *Shell) *TestCompleterHelper {
	return &TestCompleterHelper{completer: NewCompleter(s.commands)}
}

// TestCompleterHelper wraps a Completer for testing without importing readline in _test files.
type TestCompleterHelper struct {
	completer *Completer
}

// GetCompletions returns string candidates for the given line and cursor position.
func (h *TestCompleterHelper) GetCompletions(line string, cursor int) []string {
	comps := h.completer.Complete([]rune(line), cursor)

	var result []string

	comps.EachValue(func(comp readline.Completion) readline.Completion {
		result = append(result, comp.Value)
		return comp
	})

	return result
}
