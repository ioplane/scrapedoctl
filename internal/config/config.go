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

// loadedPath is the path from which the config was loaded.
var loadedPath string

// Save writes the current global, repl, logging, and cache config back to the configuration file.
// It ensures the parent directory exists and uses strict file permissions (0600).
func (c *Config) Save() error {
	k := koanf.New(".")

	data := c.buildSaveData()

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

	if err := os.WriteFile(path, out, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// buildSaveData converts the config struct into a map suitable for koanf serialization.
func (c *Config) buildSaveData() map[string]any {
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

	providers := make(map[string]any)
	for name, p := range c.Providers {
		providers[name] = map[string]any{
			"token":   p.Token,
			"type":    p.Type,
			"command": p.Command,
			"args":    p.Args,
			"engines": p.Engines,
		}
	}

	return map[string]any{
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
			"enabled":       c.Cache.Enabled,
			"path":          c.Cache.Path,
			"ttl_days":      c.Cache.TTLDays,
			"keep_versions": c.Cache.KeepVersions,
			"max_size_mb":   c.Cache.MaxSizeMB,
		},
		"search": map[string]any{
			"default_provider": c.Search.DefaultProvider,
			"default_engine":   c.Search.DefaultEngine,
			"default_limit":    c.Search.DefaultLimit,
		},
		"profiles":  profiles,
		"providers": providers,
	}
}

// DefaultConfigPath is the default location for the configuration file.
const DefaultConfigPath = "~/.scrapedoctl/conf.toml"

// SearchConfig holds defaults for the search subsystem.
type SearchConfig struct {
	// DefaultProvider is the preferred search provider name.
	DefaultProvider string `koanf:"default_provider"`
	// DefaultEngine is the default search engine (e.g. google, bing).
	DefaultEngine string `koanf:"default_engine"`
	// DefaultLimit is the default maximum number of results.
	DefaultLimit int `koanf:"default_limit"`
}

// ProviderConfig describes a single search provider entry.
type ProviderConfig struct {
	// Token is the API key for the provider.
	Token string `koanf:"token"`
	// Type is the provider type: "" (built-in) or "exec".
	Type string `koanf:"type"`
	// Command is the executable path for exec-type providers.
	Command string `koanf:"command"`
	// Args are extra command-line arguments for exec-type providers.
	Args []string `koanf:"args"`
	// Engines lists the search engines this provider supports.
	Engines []string `koanf:"engines"`
}

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
	// Search holds defaults for the search subsystem.
	Search SearchConfig `koanf:"search"`
	// Providers holds named search provider configurations.
	Providers map[string]ProviderConfig `koanf:"providers"`

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
	// ErrConfigPathIsDirectory is returned when the configuration path is a directory.
	ErrConfigPathIsDirectory = errors.New("config path is a directory")
)

// Load reads and merges configuration from defaults, file, environment, and flags.
func Load(configPath, profileName string) (*Config, error) {
	loadedPath = configPath
	k := koanf.New(".")

	if err := loadDefaults(k); err != nil {
		return nil, err
	}

	fileMissing, err := loadFile(k, configPath)
	if err != nil {
		return nil, err
	}

	if err := loadEnv(k); err != nil {
		return nil, err
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.resolveProfile(k, profileName); err != nil {
		return nil, err
	}

	cfg.Repl.HistoryFile = expandPath(cfg.Repl.HistoryFile)
	cfg.Logging.Path = expandPath(cfg.Logging.Path)
	cfg.Cache.Path = expandPath(cfg.Cache.Path)

	if fileMissing {
		return &cfg, ErrConfigNotFound
	}

	return &cfg, nil
}

func loadDefaults(k *koanf.Koanf) error {
	if err := k.Load(confmap.Provider(map[string]any{
		"global.base_url":         "https://api.scrape.do",
		"global.timeout":          60000,
		"repl.history_file":       "~/.scrapedoctl/history",
		"logging.level":           "info",
		"logging.format":          "json",
		"logging.path":            "/var/log/scrapedoctl/scrapedoctl.log",
		"logging.max_size":        10,
		"logging.max_age":         7,
		"logging.max_backups":     5,
		"logging.compress":        true,
		"cache.enabled":           true,
		"cache.path":              "~/.scrapedoctl/cache.db",
		"cache.ttl_days":          7,
		"cache.keep_versions":     5,
		"cache.max_size_mb":       100,
		"search.default_provider": "scrapedo",
		"search.default_engine":   "google",
		"search.default_limit":    10,
	}, "."), nil); err != nil {
		return fmt.Errorf("failed to load defaults: %w", err)
	}
	return nil
}

func loadFile(k *koanf.Koanf, configPath string) (bool, error) {
	path := expandPath(configPath)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, fmt.Errorf("failed to check config file: %w", err)
	}
	if info.IsDir() {
		return false, fmt.Errorf("%w: %s", ErrConfigPathIsDirectory, path)
	}

	var parser koanf.Parser
	switch filepath.Ext(path) {
	case ".toml":
		parser = toml.Parser()
	case ".yaml", ".yml":
		parser = yaml.Parser()
	case ".json":
		parser = json.Parser()
	default:
		parser = toml.Parser()
	}

	if err := k.Load(file.Provider(path), parser); err != nil {
		return false, fmt.Errorf("failed to load config file: %w", err)
	}
	return false, nil
}

func loadEnv(k *koanf.Koanf) error {
	if err := k.Load(env.Provider("SCRAPEDO_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "SCRAPEDO_")), "_", ".")
	}), nil); err != nil {
		return fmt.Errorf("failed to load env: %w", err)
	}
	return nil
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
