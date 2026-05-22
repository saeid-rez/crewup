package mcp

import "fmt"

// InstallOptions carries user-supplied values for an MCP preset install.
type InstallOptions struct {
	Values map[string]string // InputField.Key → user-supplied string
}

// MCPMerger merges an MCP preset config into a specific tool's config file.
type MCPMerger interface {
	Merge(preset MCPPreset, scope Scope, opts InstallOptions) error
}

var mergers = map[string]MCPMerger{
	"copilot":  &VSCodeMCPMerger{},
	"opencode": &OpenCodeMCPMerger{},
}

// Install adds an MCP preset config to the specified AI tool.
func Install(serverID string, toolID string, scope Scope, opts InstallOptions) error {
	preset, ok := FindByID(serverID)
	if !ok {
		return fmt.Errorf("unknown MCP server: %s", serverID)
	}

	merger, ok := mergers[toolID]
	if !ok {
		return fmt.Errorf("MCP not supported for tool: %s", toolID)
	}

	return merger.Merge(preset, scope, opts)
}
