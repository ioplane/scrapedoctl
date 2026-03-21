package repl

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
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
	client, _ := scrapedo.NewClient("token")
	s := NewShell(client)

	t.Run("successful run with exit", func(t *testing.T) {
		reader := &MockReader{lines: []string{"help", "exit"}}
		s.SetReader(reader)
		err := s.Run(context.Background())
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("empty lines and unknown command", func(t *testing.T) {
		reader := &MockReader{lines: []string{"", "unknown", "quit"}}
		s.SetReader(reader)
		err := s.Run(context.Background())
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("readline failure", func(t *testing.T) {
		reader := &MockReader{lines: []string{}} // index 0 >= len 0 -> EOF
		s.SetReader(reader)
		err := s.Run(context.Background())
		if err == nil {
			t.Error("Expected error on Readline failure (EOF)")
		}
	})
}

func TestREPL_ExecuteCommand(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mocked result"))
	}))
	defer ts.Close()

	client, _ := scrapedo.NewClient("token")
	client.SetBaseURL(ts.URL)
	s := NewShell(client)

	ctx := context.Background()

	tests := []struct {
		name    string
		line    string
		wantErr error
	}{
		{"exit", "exit", errExit},
		{"quit", "quit", errExit},
		{"help", "help", nil},
		{"scrape", "scrape http://example.com", nil},
		{"scrape with params", "scrape http://example.com render=true super=true", nil},
		{"scrape invalid", "scrape", errInvalidUsage},
		{"unknown", "unknown", errUnknownCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ExecuteCommand(ctx, tt.line)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestREPL_ScrapeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client, _ := scrapedo.NewClient("token")
	client.SetBaseURL(ts.URL)
	s := NewShell(client)

	err := s.ExecuteCommand(context.Background(), "scrape http://example.com")
	if err == nil {
		t.Error("Expected error from failed scrape")
	}
}
