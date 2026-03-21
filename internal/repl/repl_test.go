package repl_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ioplane/scrapedoctl/internal/cache"
	"github.com/ioplane/scrapedoctl/internal/config"
	"github.com/ioplane/scrapedoctl/internal/repl"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
	"github.com/ioplane/scrapedoctl/pkg/search"
)

type MockReader struct {
	lines []string
	index int
}

func (m *MockReader) Readline() (string, error) {
	if m.index >= len(m.lines) {
		return "", io.EOF
	}
	line := m.lines[m.index]
	m.index++
	return line, nil
}

func TestREPL_Run(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)
	s := repl.NewShell(client)

	t.Run("successful run with exit", func(t *testing.T) {
		reader := &MockReader{lines: []string{"help", "exit"}}
		s.SetReader(reader)
		err := s.Run(context.Background())
		require.NoError(t, err)
	})

	t.Run("empty lines and unknown command", func(t *testing.T) {
		reader := &MockReader{lines: []string{"", "unknown", "quit"}}
		s.SetReader(reader)
		err := s.Run(context.Background())
		require.NoError(t, err)
	})

	t.Run("readline failure", func(t *testing.T) {
		reader := &MockReader{lines: []string{}} // index 0 >= len 0 -> EOF
		s.SetReader(reader)
		err := s.Run(context.Background())
		require.Error(t, err)
	})

	t.Run("exit command", func(t *testing.T) {
		reader := &MockReader{lines: []string{"exit"}}
		s.SetReader(reader)
		err := s.Run(context.Background())
		assert.NoError(t, err)
	})

	t.Run("quit command", func(t *testing.T) {
		reader := &MockReader{lines: []string{"quit"}}
		s.SetReader(reader)
		err := s.Run(context.Background())
		assert.NoError(t, err)
	})
}

func TestREPL_ExecuteCommand(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mocked result"))
	}))
	defer ts.Close()

	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)
	client.SetBaseURL(ts.URL)
	s := repl.NewShell(client)

	ctx := context.Background()

	tests := []struct {
		name    string
		line    string
		wantErr error
	}{
		{"exit", "exit", repl.ErrExit},
		{"quit", "quit", repl.ErrExit},
		{"help", "help", nil},
		{"scrape", "scrape http://example.com", nil},
		{"scrape with params", "scrape http://example.com render=true super=true", nil},
		{"scrape invalid", "scrape", repl.ErrInvalidUsage},
		{"unknown", "unknown", repl.ErrUnknownCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ExecuteCommand(ctx, tt.line)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestREPL_ScrapeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)
	client.SetBaseURL(ts.URL)
	s := repl.NewShell(client)

	err = s.ExecuteCommand(context.Background(), "scrape http://example.com")
	require.Error(t, err)
}

// --- Mock search provider ---

type mockProvider struct {
	name    string
	engines []string
	resp    *search.Response
	err     error
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Engines() []string { return m.engines }
func (m *mockProvider) Search(_ context.Context, _ string, _ search.Options) (*search.Response, error) {
	return m.resp, m.err
}

// --- Mock cache store ---

type mockCacheStore struct {
	history []cache.ScrapeRecord
	stats   cache.Stats
	err     error
}

func (m *mockCacheStore) GetHistory(_ context.Context, _ string) ([]cache.ScrapeRecord, error) {
	return m.history, m.err
}

func (m *mockCacheStore) GetStats(_ context.Context) (cache.Stats, error) {
	return m.stats, m.err
}

func (m *mockCacheStore) Clear(_ context.Context) error {
	return m.err
}

// --- New command tests ---

func TestREPL_SearchCommand(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)

	router := search.NewRouter()
	router.Register(&mockProvider{
		name:    "mock",
		engines: []string{"google"},
		resp: &search.Response{
			Query:    "test",
			Engine:   "google",
			Provider: "mock",
			Results: []search.Result{
				{Position: 1, Title: "Test", URL: "http://example.com", Snippet: "snippet"},
			},
		},
	})

	s := repl.NewShell(client, repl.WithSearchRouter(router))
	ctx := context.Background()

	t.Run("basic search", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "search test query")
		require.NoError(t, err)
	})

	t.Run("search with engine param", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "search test engine=google")
		require.NoError(t, err)
	})

	t.Run("search no query", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "search")
		require.ErrorIs(t, err, repl.ErrInvalidUsage)
	})

	t.Run("search no router", func(t *testing.T) {
		s2 := repl.NewShell(client)
		err := s2.ExecuteCommand(ctx, "search test")
		require.ErrorIs(t, err, repl.ErrNoRouter)
	})
}

func TestREPL_HistoryCommand(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)

	mc := &mockCacheStore{
		history: []cache.ScrapeRecord{
			{ID: 1, URL: "http://example.com"},
		},
	}

	s := repl.NewShell(client, repl.WithCache(mc))
	ctx := context.Background()

	t.Run("history with results", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "history http://example.com")
		require.NoError(t, err)
	})

	t.Run("history no url", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "history")
		require.ErrorIs(t, err, repl.ErrInvalidUsage)
	})

	t.Run("history empty results", func(t *testing.T) {
		mc2 := &mockCacheStore{history: nil}
		s2 := repl.NewShell(client, repl.WithCache(mc2))
		err := s2.ExecuteCommand(ctx, "history http://nothing.com")
		require.NoError(t, err)
	})

	t.Run("history no cache", func(t *testing.T) {
		s2 := repl.NewShell(client)
		err := s2.ExecuteCommand(ctx, "history http://example.com")
		require.ErrorIs(t, err, repl.ErrNoCache)
	})
}

func TestREPL_CacheStatsCommand(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)

	mc := &mockCacheStore{
		stats: cache.Stats{TotalCount: 42, TotalSize: 1024},
	}

	s := repl.NewShell(client, repl.WithCache(mc))
	ctx := context.Background()

	err = s.ExecuteCommand(ctx, "cache stats")
	require.NoError(t, err)
}

func TestREPL_CacheClearCommand(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)

	mc := &mockCacheStore{}

	s := repl.NewShell(client, repl.WithCache(mc))
	ctx := context.Background()

	err = s.ExecuteCommand(ctx, "cache clear")
	require.NoError(t, err)
}

func TestREPL_ConfigListCommand(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)

	cfg := &config.Config{
		Global: config.GlobalConfig{
			Token:   "test-token",
			BaseURL: "https://api.scrape.do",
			Timeout: 60000,
		},
		Repl: config.ReplConfig{
			HistoryFile: "/tmp/history",
		},
	}

	s := repl.NewShell(client, repl.WithConfig(cfg))
	ctx := context.Background()

	err = s.ExecuteCommand(ctx, "config list")
	require.NoError(t, err)
}

func TestREPL_ConfigGetCommand(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)

	cfg := &config.Config{
		Global: config.GlobalConfig{
			Token:   "test-token",
			BaseURL: "https://api.scrape.do",
		},
	}

	s := repl.NewShell(client, repl.WithConfig(cfg))
	ctx := context.Background()

	t.Run("valid key", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "config get global.token")
		require.NoError(t, err)
	})

	t.Run("invalid key", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "config get unknown.key")
		require.ErrorIs(t, err, repl.ErrUnsupportedKey)
	})

	t.Run("no key", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "config get")
		require.ErrorIs(t, err, repl.ErrInvalidUsage)
	})
}

func TestREPL_HelpCommand(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)
	s := repl.NewShell(client)
	ctx := context.Background()

	t.Run("general help", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "help")
		require.NoError(t, err)
	})

	t.Run("help for specific command", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "help search")
		require.NoError(t, err)
	})
}

func TestREPL_HelpSpecificCommand(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)
	s := repl.NewShell(client)
	ctx := context.Background()

	err = s.ExecuteCommand(ctx, "help search")
	require.NoError(t, err)
}

func TestREPL_UnknownCommand(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)
	s := repl.NewShell(client)
	ctx := context.Background()

	err = s.ExecuteCommand(ctx, "foobar")
	require.ErrorIs(t, err, repl.ErrUnknownCmd)
}

func TestREPL_EmptyLine(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)
	s := repl.NewShell(client)
	ctx := context.Background()

	err = s.ExecuteCommand(ctx, "")
	require.NoError(t, err)
}

func TestREPL_CacheNoStore(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)
	s := repl.NewShell(client)
	ctx := context.Background()

	t.Run("stats no cache", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "cache stats")
		require.ErrorIs(t, err, repl.ErrNoCache)
	})

	t.Run("clear no cache", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "cache clear")
		require.ErrorIs(t, err, repl.ErrNoCache)
	})
}

func TestREPL_ConfigNoConfig(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)
	s := repl.NewShell(client)
	ctx := context.Background()

	t.Run("list no config", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "config list")
		require.ErrorIs(t, err, repl.ErrNoConfig)
	})

	t.Run("get no config", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "config get global.token")
		require.ErrorIs(t, err, repl.ErrNoConfig)
	})

	t.Run("set no config", func(t *testing.T) {
		err := s.ExecuteCommand(ctx, "config set global.token=foo")
		require.ErrorIs(t, err, repl.ErrNoConfig)
	})
}

func TestREPL_CacheSubcommandHelp(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)
	s := repl.NewShell(client)
	ctx := context.Background()

	// "cache" without subcommand shows help.
	err = s.ExecuteCommand(ctx, "cache")
	require.NoError(t, err)
}

func TestREPL_ConfigSubcommandHelp(t *testing.T) {
	client, err := scrapedo.NewClient("token")
	require.NoError(t, err)
	s := repl.NewShell(client)
	ctx := context.Background()

	// "config" without subcommand shows help.
	err = s.ExecuteCommand(ctx, "config")
	require.NoError(t, err)
}
