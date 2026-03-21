package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ioplane/scrapedoctl/internal/config"
)

func TestInit(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := config.LoggingConfig{
		Level:      "debug",
		Format:     "json",
		Path:       logPath,
		MaxSize:    1,
		MaxAge:     1,
		MaxBackups: 1,
		Compress:   true,
	}

	t.Run("initialize json logger with file", func(t *testing.T) {
		Init(cfg)
		slog.Debug("test debug message")
		
		// Verify file exists
		assert.FileExists(t, logPath)
		
		data, err := os.ReadFile(logPath)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "test debug message")
		assert.Contains(t, string(data), "\"level\":\"DEBUG\"")
	})

	t.Run("initialize text logger with stderr fallback", func(t *testing.T) {
		// Use an invalid path to trigger fallback
		cfg.Path = filepath.Join(tmpDir, "nonexistent_dir", "test.log")
		// We can't easily capture stderr here without complex piping, 
		// but we can verify Init doesn't panic.
		cfg.Format = "text"
		Init(cfg)
		slog.Info("test info message")
	})
}

func TestParseLevel(t *testing.T) {
	assert.Equal(t, slog.LevelDebug, parseLevel("debug"))
	assert.Equal(t, slog.LevelInfo, parseLevel("info"))
	assert.Equal(t, slog.LevelWarn, parseLevel("warn"))
	assert.Equal(t, slog.LevelWarn, parseLevel("warning"))
	assert.Equal(t, slog.LevelError, parseLevel("error"))
	assert.Equal(t, slog.LevelInfo, parseLevel("unknown"))
}
