// Package install handles the installation and configuration of scrapedoctl.
package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/v2"
)

// AgentConfig represents the server definition for Scrape.do.
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// AgentConfigInfo holds the metadata for an AI agent.
type AgentConfigInfo struct {
	ID         string
	Name       string
	ConfigPath string
	Format     string // "json" or "toml"
}

var SupportedAgents = []AgentConfigInfo{
	{ID: "claude", Name: "Claude Code", ConfigPath: "~/.claude.json", Format: "json"},
	{ID: "junie", Name: "JetBrains Junie", ConfigPath: "~/.junie/mcp/mcp.json", Format: "json"},
	{ID: "gemini", Name: "Gemini CLI", ConfigPath: "~/.gemini/settings.json", Format: "json"},
	{ID: "opencode", Name: "OpenCode AI", ConfigPath: "~/.opencode.json", Format: "json"},
	{ID: "codex", Name: "Codex AI", ConfigPath: "~/.codex/config.toml", Format: "toml"},
	{ID: "kimi", Name: "Kimi AI", ConfigPath: "~/.kimi/config.toml", Format: "toml"},
}

// ConfigureAgents injects the scrapedoctl server definition into selected agents.
func ConfigureAgents(agentIDs []string, apiToken string) error {
	exe, err := os.Executable()
	if err != nil {
		exe = "scrapedoctl" // Fallback
	}

	serverDef := MCPServerConfig{
		Command: exe,
		Args:    []string{"mcp"},
		Env: map[string]string{
			"SCRAPEDO_TOKEN": apiToken,
		},
	}

	for _, id := range agentIDs {
		for _, info := range SupportedAgents {
			if info.ID == id {
				if err := injectConfig(info, serverDef); err != nil {
					fmt.Printf("Warning: Failed to configure %s: %v\n", info.Name, err)
				} else {
					fmt.Printf("Successfully configured %s\n", info.Name)
				}
			}
		}
	}

	return nil
}

func injectConfig(info AgentConfigInfo, def MCPServerConfig) error {
	path := expandPath(info.ConfigPath)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if info.Format == "json" {
		return injectJSON(path, def)
	}
	return injectTOML(path, def)
}

func injectJSON(path string, def MCPServerConfig) error {
	data, err := os.ReadFile(path)
	config := make(map[string]any)

	if err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			// If corrupted, we'll start fresh
			config = make(map[string]any)
		}
	}

	mcpServers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		mcpServers = make(map[string]any)
	}

	mcpServers["scrape-do"] = def
	config["mcpServers"] = mcpServers

	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, newData, 0644)
}

func injectTOML(path string, def MCPServerConfig) error {
	// For TOML, we'll use koanf to merge or just simple writing if it's new.
	// This is a bit simplified for now.
	k := koanf.New(".")
	_ = k.Load(fileProvider(path), toml.Parser())

	// Set values
	_ = k.Set("mcpServers.scrape-do.command", def.Command)
	_ = k.Set("mcpServers.scrape-do.args", def.Args)
	_ = k.Set("mcpServers.scrape-do.env.SCRAPEDO_TOKEN", def.Env["SCRAPEDO_TOKEN"])

	out, err := k.Marshal(toml.Parser())
	if err != nil {
		return err
	}

	return os.WriteFile(path, out, 0644)
}

// Helper to expand ~
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// Dummy file provider for koanf since I don't want to import it here if not needed
// but wait, I already have it in the project.
type dummyProvider struct {
	path string
}

func (d *dummyProvider) ReadBytes() ([]byte, error) { return os.ReadFile(d.path) }
func (d *dummyProvider) Read() (map[string]any, error) { return nil, nil }
func fileProvider(path string) koanf.Provider { return &dummyProvider{path: path} }
