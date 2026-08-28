// Package config handles configuration loading, merging, and persistence for scrapedoctl.
// It uses koanf to merge defaults, configuration files (TOML/YAML/JSON),
// environment variables, and command-line flags.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/ioplane/scrapedoctl/internal/atomicfile"
)

// Save writes the current global, repl, logging, and cache config back to the configuration file.
// It ensures the parent directory exists and uses strict file permissions (0600).
func (c *Config) Save() error {
	if err := c.Validate(); err != nil {
		return err
	}

	k := koanf.New(".")

	data := c.buildSaveData()

	if err := k.Load(confmap.Provider(data, "."), nil); err != nil {
		return fmt.Errorf("failed to load data for save: %w", err)
	}

	out, err := marshalConfig(k, c.sourceFormat)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	path := c.sourcePath
	if path == "" {
		return ErrConfigSourceMissing
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := atomicfile.Replace(path, 0o600, out); err != nil {
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
			"render":   c.Global.Render,
			"super":    c.Global.Super,
			"geo_code": c.Global.GeoCode,
			"device":   c.Global.Device,
			"session":  c.Global.Session,
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

type configFormat uint8

const (
	configFormatTOML configFormat = iota
	configFormatJSON
	configFormatYAML
)

func configFormatForPath(path string) configFormat {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return configFormatJSON
	case ".yaml", ".yml":
		return configFormatYAML
	default:
		return configFormatTOML
	}
}

func marshalConfig(k *koanf.Koanf, format configFormat) ([]byte, error) {
	switch format {
	case configFormatJSON:
		return k.Marshal(json.Parser())
	case configFormatYAML:
		return k.Marshal(yaml.Parser())
	default:
		return k.Marshal(toml.Parser())
	}
}

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

	sourcePath   string
	sourceFormat configFormat
}

// Validate rejects unsafe or unusable runtime configuration before it reaches a client or store.
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("%w: configuration is nil", ErrInvalidConfig)
	}

	endpoint, err := url.ParseRequestURI(c.Global.BaseURL)
	if err != nil || endpoint.Scheme != "https" || endpoint.Host == "" || endpoint.User != nil {
		return fmt.Errorf("%w: global.base_url must be an HTTPS URL without user info", ErrInvalidConfig)
	}
	if c.Global.Timeout <= 0 {
		return fmt.Errorf("%w: global.timeout must be positive", ErrInvalidConfig)
	}
	if c.Cache.Enabled {
		switch {
		case c.Cache.TTLDays <= 0:
			return fmt.Errorf("%w: cache.ttl_days must be positive", ErrInvalidConfig)
		case c.Cache.KeepVersions <= 0:
			return fmt.Errorf("%w: cache.keep_versions must be positive", ErrInvalidConfig)
		case c.Cache.MaxSizeMB <= 0:
			return fmt.Errorf("%w: cache.max_size_mb must be positive", ErrInvalidConfig)
		}
	}

	return nil
}

// RedactedSecret returns a stable marker for a configured secret without revealing it.
func RedactedSecret(value string) string {
	if value == "" {
		return ""
	}
	return "***"
}

// GlobalConfig holds core API settings.
type GlobalConfig struct {
	// Token is the Scrape.do API key.
	Token string `koanf:"token"`
	// BaseURL is the Scrape.do API endpoint.
	BaseURL string `koanf:"base_url"`
	// Timeout is the request timeout in milliseconds.
	Timeout int `koanf:"timeout"`
	// Render enables JavaScript rendering by default.
	Render bool `koanf:"render"`
	// Super enables residential proxies by default.
	Super bool `koanf:"super"`
	// GeoCode is the default proxy country.
	GeoCode string `koanf:"geo_code"`
	// Device is the default emulated device.
	Device string `koanf:"device"`
	// Session is the default sticky-session identifier.
	Session string `koanf:"session"`
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
	// ErrInvalidConfig is returned when configuration values violate runtime constraints.
	ErrInvalidConfig = errors.New("invalid configuration")
	// ErrConfigSourceMissing is returned when an in-memory config has no save destination.
	ErrConfigSourceMissing = errors.New("configuration source path is missing")
)

// Load reads and merges configuration from defaults, file, environment, and flags.
func Load(configPath, profileName string) (*Config, error) {
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
	cfg.sourcePath = expandPath(configPath)
	cfg.sourceFormat = configFormatForPath(configPath)
	if err := cfg.Validate(); err != nil {
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
	values := make(map[string]any)
	for environmentName, configKey := range legacyEnvironmentKeys {
		if value, ok := os.LookupEnv(environmentName); ok {
			values[configKey] = value
		}
	}
	for _, entry := range os.Environ() {
		name, value, ok := strings.Cut(entry, "=")
		if !ok || !strings.HasPrefix(name, "SCRAPEDO_") {
			continue
		}
		rawKey := strings.TrimPrefix(name, "SCRAPEDO_")
		if !strings.Contains(rawKey, "__") {
			continue
		}
		configKey := strings.ToLower(strings.ReplaceAll(rawKey, "__", "."))
		if isSupportedEnvironmentKey(configKey) {
			values[configKey] = value
		}
	}

	if err := k.Load(confmap.Provider(values, "."), nil); err != nil {
		return fmt.Errorf("failed to load env: %w", err)
	}
	return nil
}

var legacyEnvironmentKeys = map[string]string{
	"SCRAPEDO_GLOBAL_TOKEN":            "global.token",
	"SCRAPEDO_GLOBAL_BASE_URL":         "global.base_url",
	"SCRAPEDO_GLOBAL_TIMEOUT":          "global.timeout",
	"SCRAPEDO_GLOBAL_RENDER":           "global.render",
	"SCRAPEDO_GLOBAL_SUPER":            "global.super",
	"SCRAPEDO_GLOBAL_GEO_CODE":         "global.geo_code",
	"SCRAPEDO_GLOBAL_DEVICE":           "global.device",
	"SCRAPEDO_GLOBAL_SESSION":          "global.session",
	"SCRAPEDO_REPL_HISTORY_FILE":       "repl.history_file",
	"SCRAPEDO_LOGGING_LEVEL":           "logging.level",
	"SCRAPEDO_LOGGING_FORMAT":          "logging.format",
	"SCRAPEDO_LOGGING_PATH":            "logging.path",
	"SCRAPEDO_LOGGING_MAX_SIZE":        "logging.max_size",
	"SCRAPEDO_LOGGING_MAX_AGE":         "logging.max_age",
	"SCRAPEDO_LOGGING_MAX_BACKUPS":     "logging.max_backups",
	"SCRAPEDO_LOGGING_COMPRESS":        "logging.compress",
	"SCRAPEDO_CACHE_ENABLED":           "cache.enabled",
	"SCRAPEDO_CACHE_PATH":              "cache.path",
	"SCRAPEDO_CACHE_TTL_DAYS":          "cache.ttl_days",
	"SCRAPEDO_CACHE_KEEP_VERSIONS":     "cache.keep_versions",
	"SCRAPEDO_CACHE_MAX_SIZE_MB":       "cache.max_size_mb",
	"SCRAPEDO_SEARCH_DEFAULT_PROVIDER": "search.default_provider",
	"SCRAPEDO_SEARCH_DEFAULT_ENGINE":   "search.default_engine",
	"SCRAPEDO_SEARCH_DEFAULT_LIMIT":    "search.default_limit",
}

func isSupportedEnvironmentKey(key string) bool {
	for _, supportedKey := range legacyEnvironmentKeys {
		if key == supportedKey {
			return true
		}
	}

	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return false
	}
	switch parts[0] {
	case "profiles":
		switch parts[2] {
		case "render", "super", "geo_code", "device", "session":
			return parts[1] != ""
		}
	case "providers":
		switch parts[2] {
		case "token", "type", "command":
			return parts[1] != ""
		}
	}
	return false
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
