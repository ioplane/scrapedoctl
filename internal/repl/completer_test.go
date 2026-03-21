package repl_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ioplane/scrapedoctl/internal/repl"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
	"github.com/ioplane/scrapedoctl/pkg/search"
)

func newTestShellForCompletion(t *testing.T) *repl.Shell {
	t.Helper()

	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)

	router := search.NewRouter()
	router.Register(&mockProvider{
		name:    "mock",
		engines: []string{"google", "bing"},
	})

	return repl.NewShell(client, repl.WithSearchRouter(router))
}

func TestCompleter_Commands(t *testing.T) {
	s := newTestShellForCompletion(t)
	c := repl.NewCompleterFromShell(s)

	t.Run("prefix se matches search", func(t *testing.T) {
		completions := c.GetCompletions("se", 2)
		assert.Contains(t, completions, "search")
		assert.NotContains(t, completions, "exit")
	})

	t.Run("prefix s matches search and scrape", func(t *testing.T) {
		completions := c.GetCompletions("s", 1)
		sort.Strings(completions)
		assert.Contains(t, completions, "search")
		assert.Contains(t, completions, "scrape")
	})

	t.Run("prefix h matches help and history", func(t *testing.T) {
		completions := c.GetCompletions("h", 1)
		sort.Strings(completions)
		assert.Contains(t, completions, "help")
		assert.Contains(t, completions, "history")
	})
}

func TestCompleter_SearchParams(t *testing.T) {
	s := newTestShellForCompletion(t)
	c := repl.NewCompleterFromShell(s)

	t.Run("partial eng completes to engine=", func(t *testing.T) {
		completions := c.GetCompletions("search foo eng", 14)
		assert.Contains(t, completions, "engine=")
	})

	t.Run("engine= completes with engine names", func(t *testing.T) {
		completions := c.GetCompletions("search foo engine=", 18)
		assert.Contains(t, completions, "engine=google")
		assert.Contains(t, completions, "engine=bing")
	})

	t.Run("trailing space shows all params", func(t *testing.T) {
		completions := c.GetCompletions("search foo ", 11)
		assert.Contains(t, completions, "engine=")
		assert.Contains(t, completions, "provider=")
		assert.Contains(t, completions, "lang=")
		assert.Contains(t, completions, "limit=")
	})
}

func TestCompleter_CacheSubcmd(t *testing.T) {
	s := newTestShellForCompletion(t)
	c := repl.NewCompleterFromShell(s)

	completions := c.GetCompletions("cache ", 6)
	sort.Strings(completions)

	assert.Contains(t, completions, "stats")
	assert.Contains(t, completions, "clear")
}

func TestCompleter_ConfigSubcmd(t *testing.T) {
	s := newTestShellForCompletion(t)
	c := repl.NewCompleterFromShell(s)

	completions := c.GetCompletions("config ", 7)
	sort.Strings(completions)

	assert.Contains(t, completions, "list")
	assert.Contains(t, completions, "get")
	assert.Contains(t, completions, "set")
}

func TestCompleter_Empty(t *testing.T) {
	s := newTestShellForCompletion(t)
	c := repl.NewCompleterFromShell(s)

	completions := c.GetCompletions("", 0)

	// Should return all command names (minus quit duplicate).
	assert.GreaterOrEqual(t, len(completions), 6)
	assert.Contains(t, completions, "search")
	assert.Contains(t, completions, "scrape")
	assert.Contains(t, completions, "history")
	assert.Contains(t, completions, "cache")
	assert.Contains(t, completions, "config")
	assert.Contains(t, completions, "help")
	assert.Contains(t, completions, "exit")
	// "quit" is filtered from completions.
	assert.NotContains(t, completions, "quit")
}

func TestCompleter_HelpSubcommand(t *testing.T) {
	s := newTestShellForCompletion(t)
	c := repl.NewCompleterFromShell(s)

	completions := c.GetCompletions("help ", 5)

	assert.Contains(t, completions, "search")
	assert.Contains(t, completions, "scrape")
}

func TestCompleter_PartialSearch(t *testing.T) {
	s := newTestShellForCompletion(t)
	c := repl.NewCompleterFromShell(s)

	completions := c.GetCompletions("search foo provider=", 20)

	assert.Contains(t, completions, "provider=mock")
}
