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

// MCPMerger merges an MCP server config into a specific tool's config file.
type MCPMerger interface {
	Merge(server Server, scope Scope, opts InstallOptions) error
}

var mergers = map[string]MCPMerger{
	"copilot":  &VSCodeMCPMerger{},
	"opencode": &OpenCodeMCPMerger{},
}

// Install adds an MCP server config to the specified AI tool.
func Install(serverID string, toolID string, scope Scope, opts InstallOptions) error {
	server, ok := FindByID(serverID)
	if !ok {
		return fmt.Errorf("unknown MCP server: %s", serverID)
	}

	merger, ok := mergers[toolID]
	if !ok {
		return fmt.Errorf("MCP not supported for tool: %s", toolID)
	}

	return merger.Merge(server, scope, opts)
}

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

// --- VS Code / GitHub Copilot ---

// VSCodeMCPMerger writes MCP config to ~/.vscode/mcp.json (global) or .vscode/mcp.json (project).
type VSCodeMCPMerger struct{}

func (m *VSCodeMCPMerger) Merge(server Server, scope Scope, opts InstallOptions) error {
	var path string
	var err error

	if scope == ScopeGlobal {
		path, err = expandHome("~/.vscode/mcp.json")
	} else {
		path = filepath.Join(".vscode", "mcp.json")
	}
	if err != nil {
		return err
	}

	existing, err := readJSONOrEmpty(path)
	if err != nil {
		return err
	}

	servers, err := getOrCreateMap(existing, "servers", path)
	if err != nil {
		return err
	}
	if _, exists := servers[server.ID]; exists {
		fmt.Printf("  ℹ️  %s already configured for VS Code — skipping.\n", server.Name)
		return nil
	}

	servers[server.ID] = server.ConfigSnippet("copilot")(opts)
	existing["servers"] = servers

	if err := writeJSONWithBackup(path, existing); err != nil {
		return fmt.Errorf("writing VS Code MCP config: %w", err)
	}
	fmt.Printf("  ✅ %s added to VS Code (%s)\n", server.Name, path)
	return nil
}

// --- OpenCode ---

// OpenCodeMCPMerger writes MCP config to ~/.config/opencode/opencode.json (global)
// or opencode.json (project).
type OpenCodeMCPMerger struct{}

func (m *OpenCodeMCPMerger) Merge(server Server, scope Scope, opts InstallOptions) error {
	path := "opencode.json"

	existing, err := readJSONOrEmpty(path)
	if err != nil {
		return err
	}

	mcpServers, err := getOrCreateMap(existing, "mcp", path)
	if err != nil {
		return err
	}
	if _, exists := mcpServers[server.ID]; exists {
		fmt.Printf("  ℹ️  %s already configured for OpenCode — skipping.\n", server.Name)
		return nil
	}

	mcpServers[server.ID] = server.ConfigSnippet("opencode")(opts)
	existing["mcp"] = mcpServers

	if err := writeJSONWithBackup(path, existing); err != nil {
		return fmt.Errorf("writing OpenCode MCP config: %w", err)
	}
	fmt.Printf("  ✅ %s added to OpenCode (%s)\n", server.Name, path)
	return nil
}
