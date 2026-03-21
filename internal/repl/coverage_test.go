package repl

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

type MockErrorReader struct {
	err error
}

func (m *MockErrorReader) Readline() (string, error) {
	return "", m.err
}

func TestREPL_Run_Errors(t *testing.T) {
	client, _ := scrapedo.NewClient("token")
	s := NewShell(client)

	t.Run("custom error", func(t *testing.T) {
		s.SetReader(&MockErrorReader{err: errors.New("custom failure")})
		err := s.Run(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "custom failure")
	})

	// Note: We avoid testing s.reader == nil because readline.NewShell()
	// might hang or fail in headless CI environments.
}
