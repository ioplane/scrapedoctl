// Package config handles configuration loading, merging, and persistence for scrapedoctl.
// It uses koanf to merge defaults, configuration files (TOML/YAML/JSON),
// environment variables, and command-line flags.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	// loadedPath is the path from which the config was loaded.
	loadedPath string
)

// Save writes the current global, repl, logging, and cache config back to the configuration file.
// It ensures the parent directory exists and uses strict file permissions (0600).
func (c *Config) Save() error {
	k := koanf.New(".")

	profiles := make(map[string]any)
	for name, p := range c.Profiles {
		profiles[name] = map[string]any{
			"render":   p.Render,
			"super":    p.Super,
			"geo_code": p.GeoCode,
			"device":   p.Device,
			"session":  p.Session,
		}
	}

	// We need to load from a map to avoid "cannot convert to Tree" errors for structs
	data := map[string]any{
		"global": map[string]any{
			"token":    c.Global.Token,
			"base_url": c.Global.BaseURL,
			"timeout":  c.Global.Timeout,
		},
		"repl": map[string]any{
			"history_file": c.Repl.HistoryFile,
		},
		"logging": map[string]any{
			"level":       c.Logging.Level,
			"format":      c.Logging.Format,
			"path":        c.Logging.Path,
			"max_size":    c.Logging.MaxSize,
			"max_age":     c.Logging.MaxAge,
			"max_backups": c.Logging.MaxBackups,
			"compress":    c.Logging.Compress,
		},
		"cache": map[string]any{
			"enabled":        c.Cache.Enabled,
			"path":           c.Cache.Path,
			"ttl_days":       c.Cache.TTLDays,
			"keep_versions":  c.Cache.KeepVersions,
			"max_size_mb":    c.Cache.MaxSizeMB,
		},
		"profiles": profiles,
	}

	if err := k.Load(confmap.Provider(data, "."), nil); err != nil {
		return fmt.Errorf("failed to load data for save: %w", err)
	}

	out, err := k.Marshal(toml.Parser())
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	path := expandPath(loadedPath)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return os.WriteFile(path, out, 0o600)
}

// DefaultConfigPath is the default location for the configuration file.
const DefaultConfigPath = "~/.scrapedoctl/conf.toml"

// Config represents the complete application configuration.
type Config struct {
	// Global holds core API settings.
	Global GlobalConfig `koanf:"global"`
	// Repl holds interactive shell settings.
	Repl ReplConfig `koanf:"repl"`
	// Logging holds settings for the advanced logging system.
	Logging LoggingConfig `koanf:"logging"`
	// Cache holds settings for the persistent caching system.
	Cache CacheConfig `koanf:"cache"`
	// Profiles holds named configurations for quick switching.
	Profiles map[string]ProfileConfig `koanf:"profiles"`

	// ActiveProfile is the name of the profile currently in use.
	ActiveProfile string
	// Resolved is the final merged configuration for the active request.
	Resolved ProfileConfig
}

// GlobalConfig holds core API settings.
type GlobalConfig struct {
	// Token is the Scrape.do API key.
	Token string `koanf:"token"`
	// BaseURL is the Scrape.do API endpoint.
	BaseURL string `koanf:"base_url"`
	// Timeout is the request timeout in milliseconds.
	Timeout int `koanf:"timeout"`
}

// ReplConfig holds interactive shell settings.
type ReplConfig struct {
	// HistoryFile is the path to the REPL command history file.
	HistoryFile string `koanf:"history_file"`
}

// LoggingConfig holds settings for the advanced logging system.
type LoggingConfig struct {
	// Level defines the logging threshold (debug, info, warn, error).
	Level string `koanf:"level"`
	// Format defines the output format (json, text).
	Format string `koanf:"format"`
	// Path is the absolute path to the log file.
	Path string `koanf:"path"`
	// MaxSize is the size in megabytes before the log file is rotated.
	MaxSize int `koanf:"max_size"`
	// MaxAge is the maximum number of days to retain old log files.
	MaxAge int `koanf:"max_age"`
	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int `koanf:"max_backups"`
	// Compress determines if rotated logs should be gzipped.
	Compress bool `koanf:"compress"`
}

// CacheConfig holds settings for the persistent caching system.
type CacheConfig struct {
	// Enabled determines if the caching layer is active.
	Enabled bool `koanf:"enabled"`
	// Path is the absolute path to the SQLite database file.
	Path string `koanf:"path"`
	// TTLDays is the number of days a cached result is considered valid.
	TTLDays int `koanf:"ttl_days"`
	// KeepVersions is the maximum number of historical versions to keep per URL.
	KeepVersions int `koanf:"keep_versions"`
	// MaxSizeMB is the maximum total size of the cache database in megabytes.
	MaxSizeMB int `koanf:"max_size_mb"`
}

// ProfileConfig holds scrapedo request parameters that can be customized per profile.
type ProfileConfig struct {
	// Render enables JavaScript rendering.
	Render bool `koanf:"render"`
	// Super enables residential proxies.
	Super bool `koanf:"super"`
	// GeoCode routes requests through a specific country.
	GeoCode string `koanf:"geo_code"`
	// Device emulates a specific browser device.
	Device string `koanf:"device"`
	// Session maintains a sticky session ID.
	Session string `koanf:"session"`
}

var (
	errProfileNotFound = errors.New("profile not found")
	// ErrConfigNotFound is returned when the configuration file does not exist.
	ErrConfigNotFound = errors.New("config file not found")
)

// Load reads and merges configuration from defaults, file, environment, and flags.
func Load(configPath, profileName string) (*Config, error) {
	loadedPath = configPath
	k := koanf.New(".")

	// 1. Load Defaults
	if err := k.Load(confmap.Provider(map[string]any{
		"global.base_url":     "https://api.scrape.do",
		"global.timeout":      60000,
		"repl.history_file":   "~/.scrapedoctl/history",
		"logging.level":       "info",
		"logging.format":      "json",
		"logging.path":        "/var/log/scrapedoctl/scrapedoctl.log",
		"logging.max_size":    10,
		"logging.max_age":     7,
		"logging.max_backups": 5,
		"logging.compress":    true,
		"cache.enabled":       true,
		"cache.path":          "~/.scrapedoctl/cache.db",
		"cache.ttl_days":      7,
		"cache.keep_versions": 5,
		"cache.max_size_mb":   100,
	}, "."), nil); err != nil {
		return nil, fmt.Errorf("failed to load defaults: %w", err)
	}

	// 2. Load File (Optional)
	var fileMissing bool
	path := expandPath(configPath)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			fileMissing = true
		} else {
			return nil, fmt.Errorf("failed to check config file: %w", err)
		}
	} else if info.IsDir() {
		return nil, fmt.Errorf("config path is a directory: %s", path)
	} else {
		var parser koanf.Parser
		switch filepath.Ext(path) {
		case ".toml":
			parser = toml.Parser()
		case ".yaml", ".yml":
			parser = yaml.Parser()
		case ".json":
			parser = json.Parser()
		default:
			parser = toml.Parser() // Default to TOML
		}

		if err := k.Load(file.Provider(path), parser); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// 3. Load Environment (SCRAPEDO_*)
	if err := k.Load(env.Provider("SCRAPEDO_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "SCRAPEDO_")), "_", ".")
	}), nil); err != nil {
		return nil, fmt.Errorf("failed to load env: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 4. Resolve Profile
	if err := cfg.resolveProfile(k, profileName); err != nil {
		return nil, err
	}

	// Expand paths in config
	cfg.Repl.HistoryFile = expandPath(cfg.Repl.HistoryFile)
	cfg.Logging.Path = expandPath(cfg.Logging.Path)
	cfg.Cache.Path = expandPath(cfg.Cache.Path)

	if fileMissing {
		return &cfg, ErrConfigNotFound
	}

	return &cfg, nil
}

func (c *Config) resolveProfile(k *koanf.Koanf, profileName string) error {
	c.ActiveProfile = profileName
	c.Resolved = ProfileConfig{
		Render:  k.Bool("global.render"),
		Super:   k.Bool("global.super"),
		GeoCode: k.String("global.geo_code"),
		Device:  k.String("global.device"),
		Session: k.String("global.session"),
	}

	if profileName == "" {
		return nil
	}

	p, ok := c.Profiles[profileName]
	if !ok {
		return fmt.Errorf("%w: %q", errProfileNotFound, profileName)
	}

	if k.Exists("profiles." + profileName + ".render") {
		c.Resolved.Render = p.Render
	}
	if k.Exists("profiles." + profileName + ".super") {
		c.Resolved.Super = p.Super
	}
	if p.GeoCode != "" {
		c.Resolved.GeoCode = p.GeoCode
	}
	if p.Device != "" {
		c.Resolved.Device = p.Device
	}
	if p.Session != "" {
		c.Resolved.Session = p.Session
	}

	return nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
