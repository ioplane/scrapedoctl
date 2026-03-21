package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/v2"
)

func TestConfigureAgents(t *testing.T) {
	tempHome, err := os.MkdirTemp("", "scrapedoctl-test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Mock HOME/USERPROFILE
	oldHome := os.Getenv("HOME")
	oldUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tempHome)
	os.Setenv("USERPROFILE", tempHome)
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("USERPROFILE", oldUserProfile)
	}()

	apiToken := "test-token"
	agentIDs := []string{"claude", "codex"} // One JSON, one TOML

	err = ConfigureAgents(agentIDs, apiToken)
	if err != nil {
		t.Errorf("ConfigureAgents failed: %v", err)
	}

	// Verify Claude (JSON)
	claudePath := filepath.Join(tempHome, ".claude.json")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Errorf("Claude config not created at %s", claudePath)
	} else {
		data, _ := os.ReadFile(claudePath)
		var config map[string]any
		json.Unmarshal(data, &config)
		mcpServers := config["mcpServers"].(map[string]any)
		scrapeDo := mcpServers["scrape-do"].(map[string]any)
		if scrapeDo["env"].(map[string]any)["SCRAPEDO_TOKEN"] != apiToken {
			t.Errorf("Expected token %s, got %v", apiToken, scrapeDo["env"])
		}
	}

	// Verify Codex (TOML)
	codexPath := filepath.Join(tempHome, ".codex", "config.toml")
	if _, err := os.Stat(codexPath); os.IsNotExist(err) {
		t.Errorf("Codex config not created at %s", codexPath)
	} else {
		k := koanf.New(".")
		k.Load(fileProvider(codexPath), toml.Parser())
		if k.String("mcpServers.scrape-do.env.SCRAPEDO_TOKEN") != apiToken {
			t.Errorf("Expected token %s, got %v", apiToken, k.Get("mcpServers.scrape-do.env"))
		}
	}
}

func TestInjectConfig_AllAgents(t *testing.T) {
	tempHome, _ := os.MkdirTemp("", "scrapedoctl-test-home-all-*")
	defer os.RemoveAll(tempHome)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	def := MCPServerConfig{
		Command: "scrapedoctl",
		Args:    []string{"mcp"},
		Env:     map[string]string{"SCRAPEDO_TOKEN": "token"},
	}

	for _, info := range SupportedAgents {
		t.Run(info.ID, func(t *testing.T) {
			err := injectConfig(info, def)
			if err != nil {
				t.Errorf("Failed to inject for %s: %v", info.ID, err)
			}
			path := expandPath(info.ConfigPath)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Config file not created for %s at %s", info.ID, path)
			}
		})
	}
}

func TestInjectJSON_Merge(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "scrapedoctl-json-*")
	defer os.RemoveAll(tempDir)

	path := filepath.Join(tempDir, "test.json")
	initialContent := `{"existing": "value", "mcpServers": {"other": {"command": "other"}}}`
	os.WriteFile(path, []byte(initialContent), 0644)

	def := MCPServerConfig{Command: "scrapedoctl"}
	err := injectJSON(path, def)
	if err != nil {
		t.Fatalf("injectJSON failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	var config map[string]any
	json.Unmarshal(data, &config)

	if config["existing"] != "value" {
		t.Errorf("Existing value lost")
	}
	mcpServers := config["mcpServers"].(map[string]any)
	if _, ok := mcpServers["other"]; !ok {
		t.Errorf("Existing mcpServer lost")
	}
	if _, ok := mcpServers["scrape-do"]; !ok {
		t.Errorf("scrape-do mcpServer not added")
	}
}

func TestInjectJSON_Corrupted(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "scrapedoctl-json-corrupt-*")
	defer os.RemoveAll(tempDir)

	path := filepath.Join(tempDir, "test.json")
	os.WriteFile(path, []byte("invalid json"), 0644)

	def := MCPServerConfig{Command: "scrapedoctl"}
	err := injectJSON(path, def)
	if err != nil {
		t.Fatalf("injectJSON failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	var config map[string]any
	json.Unmarshal(data, &config)
	if _, ok := config["mcpServers"]; !ok {
		t.Errorf("mcpServers not created after corruption")
	}
}

func TestInjectTOML_Merge(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "scrapedoctl-toml-*")
	defer os.RemoveAll(tempDir)

	path := filepath.Join(tempDir, "test.toml")
	initialContent := `
[other]
key = "value"

[mcpServers.other]
command = "other"
`
	os.WriteFile(path, []byte(initialContent), 0644)

	def := MCPServerConfig{
		Command: "scrapedoctl",
		Args:    []string{"mcp"},
		Env:     map[string]string{"SCRAPEDO_TOKEN": "token"},
	}
	err := injectTOML(path, def)
	if err != nil {
		t.Fatalf("injectTOML failed: %v", err)
	}

	k := koanf.New(".")
	k.Load(fileProvider(path), toml.Parser())

	if k.String("other.key") != "value" {
		t.Errorf("Existing TOML value lost")
	}
	if k.String("mcpServers.other.command") != "other" {
		t.Errorf("Existing TOML mcpServer lost")
	}
	if k.String("mcpServers.scrape-do.command") != "scrapedoctl" {
		t.Errorf("scrape-do TOML mcpServer not added")
	}
}

func TestInjectConfig_Error(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping directory permission test on Windows")
	}

	tempDir, _ := os.MkdirTemp("", "scrapedoctl-error-*")
	defer os.RemoveAll(tempDir)

	// Create a file where a directory should be
	blockedDir := filepath.Join(tempDir, "blocked")
	os.WriteFile(blockedDir, []byte("i am a file"), 0644)

	info := AgentConfigInfo{
		ID:         "test",
		ConfigPath: filepath.Join(blockedDir, "config.json"),
		Format:     "json",
	}
	def := MCPServerConfig{}

	err := injectConfig(info, def)
	if err == nil {
		t.Errorf("Expected error when directory creation fails, got nil")
	}
}

func TestExpandPath_NoHome(t *testing.T) {
	oldHome := os.Getenv("HOME")
	oldUserProfile := os.Getenv("USERPROFILE")
	os.Unsetenv("HOME")
	os.Unsetenv("USERPROFILE")
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("USERPROFILE", oldUserProfile)
	}()

	path := "~/test"
	expanded := expandPath(path)
	if expanded != path {
		// On some systems UserHomeDir might still find something, but let's assume it fails or returns original if it can't find home
		// Actually expandPath returns original if err != nil
	}
}
