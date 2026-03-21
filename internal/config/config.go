// Package config handles configuration loading and merging for scrapedoctl.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	// loadedPath is the path from which the config was loaded.
	loadedPath string
)
...
// Save writes the current global and repl config back to the configuration file.
func (c *Config) Save() error {
	k := koanf.New(".")

	// Set values for global and repl
	_ = k.Set("global", c.Global)
	_ = k.Set("repl", c.Repl)
	_ = k.Set("profiles", c.Profiles)

	out, err := k.Marshal(toml.Parser())
	if err != nil {
		return err
	}

	path := expandPath(loadedPath)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, out, 0644)
}
const DefaultConfigPath = "~/.scrapedoctl/conf.toml"

// Config represents the complete application configuration.
type Config struct {
	Global   GlobalConfig             `koanf:"global"`
	Repl     ReplConfig               `koanf:"repl"`
	Profiles map[string]ProfileConfig `koanf:"profiles"`

	// ActiveProfile is the name of the profile currently in use.
	ActiveProfile string
	// Resolved is the final merged configuration for the active request.
	Resolved ProfileConfig
}

// GlobalConfig holds core API settings.
type GlobalConfig struct {
	Token   string `koanf:"token"`
	BaseURL string `koanf:"base_url"`
	Timeout int    `koanf:"timeout"`
}

// ReplConfig holds interactive shell settings.
type ReplConfig struct {
	HistoryFile string `koanf:"history_file"`
}

// ProfileConfig holds scrapedo request parameters that can be customized per profile.
type ProfileConfig struct {
	Render  bool   `koanf:"render"`
	Super   bool   `koanf:"super"`
	GeoCode string `koanf:"geo_code"`
	Device  string `koanf:"device"`
	Session string `koanf:"session"`
}

var errProfileNotFound = errors.New("profile not found")

// Load reads and merges configuration from defaults, file, environment, and flags.
func Load(configPath, profileName string) (*Config, error) {
	loadedPath = configPath
	k := koanf.New(".")

	// 1. Load Defaults
	if err := k.Load(confmap.Provider(map[string]any{
		"global.base_url":   "https://api.scrape.do",
		"global.timeout":    60000,
		"repl.history_file": "~/.scrapedoctl/history",
	}, "."), nil); err != nil {
		return nil, fmt.Errorf("failed to load defaults: %w", err)
	}

	// 2. Load File
	path := expandPath(configPath)
	if _, err := os.Stat(path); err == nil {
		if err := k.Load(file.Provider(path), toml.Parser()); err != nil {
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
