package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ioplane/scrapedoctl/internal/config"
)

func TestInit_Branches(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("mkdir error", func(t *testing.T) {
		// Create a file where a directory should be
		conflictFile := filepath.Join(tmpDir, "conflict")
		_ = os.WriteFile(conflictFile, []byte("not a dir"), 0644)

		cfg := config.LoggingConfig{
			Path:   filepath.Join(conflictFile, "log.txt"),
			Format: "text",
			Level:  "warn",
		}
		Init(cfg)
		// Should fall back to stderr and continue
	})

	t.Run("levels", func(t *testing.T) {
		levels := []string{"debug", "warn", "warning", "error", "info", "unknown"}
		for _, l := range levels {
			cfg := config.LoggingConfig{
				Level: l,
			}
			Init(cfg)
		}
	})

	t.Run("json format", func(t *testing.T) {
		cfg := config.LoggingConfig{
			Format: "json",
		}
		Init(cfg)
	})
}

func TestParseLevel_Branches(t *testing.T) {
	assert.Equal(t, slog.LevelDebug, parseLevel("debug"))
	assert.Equal(t, slog.LevelWarn, parseLevel("warn"))
	assert.Equal(t, slog.LevelWarn, parseLevel("warning"))
	assert.Equal(t, slog.LevelError, parseLevel("error"))
	assert.Equal(t, slog.LevelInfo, parseLevel("info"))
	assert.Equal(t, slog.LevelInfo, parseLevel("unknown"))
}
