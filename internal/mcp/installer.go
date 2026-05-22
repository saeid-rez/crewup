package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Scope controls whether MCP config is written globally or per-project.
type Scope string

const (
	ScopeGlobal  Scope = "global"
	ScopeProject Scope = "project"
)

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, path[1:]), nil
}

// hasSensitiveValues returns true when the preset has at least one Sensitive
// InputField for which the caller supplied a non-empty value.
func hasSensitiveValues(preset MCPPreset, values map[string]string) bool {
	for _, f := range preset.Inputs {
		if f.Sensitive && strings.TrimSpace(values[f.Key]) != "" {
			return true
		}
	}
	return false
}

// rejectSymlink returns an error if path exists and is a symlink.
// Only call this for project-scoped paths (not home-dir-resolved global paths).
func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		// File doesn't exist yet — nothing to check.
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to write MCP config: %s is a symlink — remove it and retry", path)
	}
	return nil
}

// --- VS Code / GitHub Copilot ---

// VSCodeMCPMerger writes MCP config to ~/.vscode/mcp.json (global) or .vscode/mcp.json (project).
type VSCodeMCPMerger struct{}

func (m *VSCodeMCPMerger) Merge(preset MCPPreset, scope Scope, opts InstallOptions) error {
	var path string
	var err error
	projectScoped := scope != ScopeGlobal

	if scope == ScopeGlobal {
		path, err = expandHome("~/.vscode/mcp.json")
	} else {
		path = filepath.Join(".vscode", "mcp.json")
	}
	if err != nil {
		return err
	}

	// Symlink guard — only for project-scoped paths.
	if projectScoped {
		if err := rejectSymlink(path); err != nil {
			return err
		}
	}

	existing, err := readJSONOrEmpty(path)
	if err != nil {
		return err
	}

	servers, err := getOrCreateMap(existing, "servers", path)
	if err != nil {
		return err
	}
	if _, exists := servers[preset.ID]; exists {
		fmt.Printf("  ℹ️  %s already configured for VS Code — skipping.\n", preset.Name)
		return nil
	}

	servers[preset.ID] = buildConfig(preset, "copilot", opts.Values)
	existing["servers"] = servers

	sensitive := hasSensitiveValues(preset, opts.Values)
	if err := writeJSONWithBackup(path, existing, sensitive); err != nil {
		return fmt.Errorf("writing VS Code MCP config: %w", err)
	}
	fmt.Printf("  ✅ %s added to VS Code (%s)\n", preset.Name, path)
	return nil
}

// --- OpenCode ---

// OpenCodeMCPMerger writes MCP config to ~/.config/opencode/opencode.json (global)
// or opencode.json (project).
type OpenCodeMCPMerger struct{}

func (m *OpenCodeMCPMerger) Merge(preset MCPPreset, scope Scope, opts InstallOptions) error {
	path := "opencode.json"

	// OpenCode always writes to a project-scoped path — apply symlink guard.
	if err := rejectSymlink(path); err != nil {
		return err
	}

	existing, err := readJSONOrEmpty(path)
	if err != nil {
		return err
	}

	mcpServers, err := getOrCreateMap(existing, "mcp", path)
	if err != nil {
		return err
	}
	if _, exists := mcpServers[preset.ID]; exists {
		fmt.Printf("  ℹ️  %s already configured for OpenCode — skipping.\n", preset.Name)
		return nil
	}

	mcpServers[preset.ID] = buildConfig(preset, "opencode", opts.Values)
	existing["mcp"] = mcpServers

	sensitive := hasSensitiveValues(preset, opts.Values)
	if err := writeJSONWithBackup(path, existing, sensitive); err != nil {
		return fmt.Errorf("writing OpenCode MCP config: %w", err)
	}
	fmt.Printf("  ✅ %s added to OpenCode (%s)\n", preset.Name, path)
	return nil
}
