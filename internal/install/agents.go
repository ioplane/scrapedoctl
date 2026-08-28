// Package install handles the installation and configuration of scrapedoctl.
package install

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/v2"

	"github.com/ioplane/scrapedoctl/internal/atomicfile"
)

// ErrReadNotImplemented is returned when the Read method is not implemented.
var (
	ErrReadNotImplemented      = errors.New("read not implemented")
	ErrUnsupportedAgent        = errors.New("unsupported agent")
	ErrUnsupportedConfigFormat = errors.New("unsupported agent config format")
	ErrConfigSymlink           = errors.New("existing config must not be a symlink")
	ErrConfigNotRegular        = errors.New("existing config is not a regular file")
	ErrInvalidMCPServers       = errors.New("existing JSON mcpServers value is not an object")
	ErrBackupChecksumMismatch  = errors.New("config backup checksum mismatch")
)

const (
	formatJSON = "json"
	formatTOML = "toml"
)

// MCPServerConfig represents the server definition for Scrape.do.
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// AgentConfigInfo holds the metadata for an AI agent.
type AgentConfigInfo struct {
	ID         string
	Name       string
	ConfigPath string
	Format     string // "json" or "toml"
}

// SupportedAgents contains the list of agents that can be configured.
var SupportedAgents = []AgentConfigInfo{
	{ID: "claude", Name: "Claude Code", ConfigPath: "~/.claude.json", Format: formatJSON},
	{ID: "junie", Name: "JetBrains Junie", ConfigPath: "~/.junie/mcp/mcp.json", Format: formatJSON},
	{ID: "gemini", Name: "Gemini CLI", ConfigPath: "~/.gemini/settings.json", Format: formatJSON},
	{ID: "opencode", Name: "OpenCode AI", ConfigPath: "~/.opencode.json", Format: formatJSON},
	{ID: "codex", Name: "Codex AI", ConfigPath: "~/.codex/config.toml", Format: formatTOML},
	{ID: "kimi", Name: "Kimi AI", ConfigPath: "~/.kimi/config.toml", Format: formatTOML},
}

// ConfigureAgents injects the scrapedoctl server definition into selected agents.
func ConfigureAgents(agentIDs []string, _ string) error {
	exe, err := os.Executable()
	if err != nil {
		exe = "scrapedoctl" // Fallback
	}

	serverDef := MCPServerConfig{
		Command: exe,
		Args:    []string{"mcp"},
		Env: map[string]string{ //nolint:gosec // Value is an environment-variable reference, not a credential.
			"SCRAPEDO_TOKEN": "${SCRAPEDO_TOKEN}",
		},
	}

	prepared := make([]preparedConfig, 0, len(agentIDs))
	for _, id := range agentIDs {
		info, ok := findAgent(id)
		if !ok {
			return fmt.Errorf("%w: %q", ErrUnsupportedAgent, id)
		}
		change, err := prepareConfig(info, serverDef)
		if err != nil {
			return fmt.Errorf("prepare %s config: %w", info.Name, err)
		}
		prepared = append(prepared, change)
	}
	if err := commitPrepared(prepared); err != nil {
		return err
	}
	for _, change := range prepared {
		fmt.Printf("Successfully configured %s\n", change.info.Name)
	}

	return nil
}

func injectConfig(info AgentConfigInfo, def MCPServerConfig) error {
	change, err := prepareConfig(info, def)
	if err != nil {
		return err
	}
	return commitPrepared([]preparedConfig{change})
}

func injectJSON(path string, def MCPServerConfig) error {
	change, err := preparePath(path, formatJSON, def)
	if err != nil {
		return err
	}
	return commitPrepared([]preparedConfig{change})
}

func injectTOML(path string, def MCPServerConfig) error {
	change, err := preparePath(path, formatTOML, def)
	if err != nil {
		return err
	}
	return commitPrepared([]preparedConfig{change})
}

type fileSnapshot struct {
	exists bool
	data   []byte
	mode   fs.FileMode
}

type preparedConfig struct {
	info     AgentConfigInfo
	path     string
	output   []byte
	original fileSnapshot
}

func findAgent(id string) (AgentConfigInfo, bool) {
	for _, info := range SupportedAgents {
		if info.ID == id {
			return info, true
		}
	}
	return AgentConfigInfo{}, false
}

func prepareConfig(info AgentConfigInfo, def MCPServerConfig) (preparedConfig, error) {
	change, err := preparePath(expandPath(info.ConfigPath), info.Format, def)
	change.info = info
	return change, err
}

func preparePath(path, format string, def MCPServerConfig) (preparedConfig, error) {
	original, err := snapshot(path)
	if err != nil {
		return preparedConfig{}, err
	}

	var output []byte
	switch format {
	case formatJSON:
		output, err = renderJSON(original, def)
	case formatTOML:
		output, err = renderTOML(original, def)
	default:
		err = fmt.Errorf("%w: %q", ErrUnsupportedConfigFormat, format)
	}
	if err != nil {
		return preparedConfig{}, err
	}
	return preparedConfig{path: path, output: output, original: original}, nil
}

func snapshot(path string) (fileSnapshot, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return fileSnapshot{}, nil
	}
	if err != nil {
		return fileSnapshot{}, fmt.Errorf("inspect existing config: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fileSnapshot{}, fmt.Errorf("%w: %s", ErrConfigSymlink, path)
	}
	if !info.Mode().IsRegular() {
		return fileSnapshot{}, fmt.Errorf("%w: %s", ErrConfigNotRegular, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fileSnapshot{}, fmt.Errorf("read existing config: %w", err)
	}
	return fileSnapshot{exists: true, data: data, mode: info.Mode().Perm()}, nil
}

func renderJSON(original fileSnapshot, def MCPServerConfig) ([]byte, error) {
	config := make(map[string]any)
	if original.exists {
		if err := json.Unmarshal(original.data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse existing JSON config: %w", err)
		}
	}

	mcpServers, ok := config["mcpServers"].(map[string]any)
	if !ok && config["mcpServers"] != nil {
		return nil, ErrInvalidMCPServers
	}
	if mcpServers == nil {
		mcpServers = make(map[string]any)
	}

	mcpServers["scrape-do"] = def
	config["mcpServers"] = mcpServers

	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return newData, nil
}

func renderTOML(original fileSnapshot, def MCPServerConfig) ([]byte, error) {
	k := koanf.New(".")
	if original.exists {
		if err := k.Load(&memoryProvider{data: original.data}, toml.Parser()); err != nil {
			return nil, fmt.Errorf("failed to parse existing TOML config: %w", err)
		}
	}

	// Set values
	if err := k.Set("mcpServers.scrape-do.command", def.Command); err != nil {
		return nil, fmt.Errorf("failed to set command: %w", err)
	}
	if err := k.Set("mcpServers.scrape-do.args", def.Args); err != nil {
		return nil, fmt.Errorf("failed to set args: %w", err)
	}
	if tokenReference, ok := def.Env["SCRAPEDO_TOKEN"]; ok {
		if err := k.Set("mcpServers.scrape-do.env.SCRAPEDO_TOKEN", tokenReference); err != nil {
			return nil, fmt.Errorf("failed to set env: %w", err)
		}
	}

	out, err := k.Marshal(toml.Parser())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal TOML: %w", err)
	}
	return out, nil
}

func commitPrepared(changes []preparedConfig) error {
	committed := make([]preparedConfig, 0, len(changes))
	for _, change := range changes {
		if err := os.MkdirAll(filepath.Dir(change.path), 0o700); err != nil {
			return errors.Join(fmt.Errorf("create config directory: %w", err), rollbackPrepared(committed))
		}
		if change.original.exists {
			if err := writeVerifiedBackup(change); err != nil {
				return errors.Join(err, rollbackPrepared(committed))
			}
		}
		if err := atomicfile.Replace(change.path, 0o600, change.output); err != nil {
			return errors.Join(
				fmt.Errorf("write agent config: %w", err),
				rollbackPrepared(append(committed, change)),
			)
		}
		committed = append(committed, change)
	}
	return nil
}

func writeVerifiedBackup(change preparedConfig) error {
	stamp := time.Now().UTC().Format("20060102T150405.000000000Z")
	backupPath := change.path + ".bak." + stamp
	if err := atomicfile.Replace(backupPath, 0o600, change.original.data); err != nil {
		return fmt.Errorf("write config backup: %w", err)
	}
	backup, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("verify config backup: %w", err)
	}
	if sha256.Sum256(backup) != sha256.Sum256(change.original.data) {
		return fmt.Errorf("verify config backup: %w", ErrBackupChecksumMismatch)
	}
	return nil
}

func rollbackPrepared(changes []preparedConfig) error {
	var rollbackErrors []error
	for _, change := range slices.Backward(changes) {
		if change.original.exists {
			if err := atomicfile.Replace(change.path, change.original.mode, change.original.data); err != nil {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("restore %s: %w", change.path, err))
			}
		} else {
			if err := os.Remove(change.path); err != nil && !errors.Is(err, os.ErrNotExist) {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("remove %s: %w", change.path, err))
			}
		}
	}
	return errors.Join(rollbackErrors...)
}

type memoryProvider struct {
	data []byte
}

func (p *memoryProvider) ReadBytes() ([]byte, error) {
	return bytes.Clone(p.data), nil
}

func (p *memoryProvider) Read() (map[string]any, error) {
	return nil, ErrReadNotImplemented
}

// Helper to expand ~.
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

func (d *dummyProvider) ReadBytes() ([]byte, error) {
	data, err := os.ReadFile(d.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

func (d *dummyProvider) Read() (map[string]any, error) {
	return nil, ErrReadNotImplemented
}

func fileProvider(path string) koanf.Provider {
	return &dummyProvider{path: path}
}

// ProjectFile defines a project-level integration file to generate.
type ProjectFile struct {
	Name    string
	Content string
}

// projectFiles returns the list of project integration files to generate.
func projectFiles() []ProjectFile {
	return []ProjectFile{
		{Name: ".mcp.json", Content: mcpJSONContent},
		{Name: "CLAUDE.md", Content: claudeMDContent},
		{Name: "AGENTS.md", Content: agentsMDContent},
		{Name: "GEMINI.md", Content: geminiMDContent},
	}
}

// GenerateProjectFiles creates project-level integration files in the given directory.
// Existing files are skipped to avoid overwriting user customizations.
func GenerateProjectFiles(projectDir string) error {
	for _, pf := range projectFiles() {
		if err := writeProjectFile(projectDir, pf); err != nil {
			return err
		}
	}
	return nil
}

func writeProjectFile(dir string, pf ProjectFile) error {
	path := filepath.Join(dir, pf.Name)

	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Skipped %s (already exists)\n", pf.Name)
		return nil
	}

	if err := atomicfile.Replace(path, 0o600, []byte(pf.Content)); err != nil {
		return fmt.Errorf("failed to write %s: %w", pf.Name, err)
	}

	fmt.Printf("Generated %s\n", pf.Name)
	return nil
}
