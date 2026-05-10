package tools

import "github.com/saeid-rez/crewup/internal/config"

// AgentConfigWriter writes agent role configs for a specific AI tool.
type AgentConfigWriter interface {
	WriteAgentConfig(configPath string, roles []config.AgentRole) error
}

// MCPConfigWriter writes MCP server configs for a specific AI tool.
// Not all tools support MCP — check with type assertion.
type MCPConfigWriter interface {
	WriteMCPConfig(configPath string, serverIDs []string) error
}

// Writers maps tool ID → writer (implements AgentConfigWriter and optionally MCPConfigWriter).
var Writers = map[string]AgentConfigWriter{
	"copilot":  &CopilotWriter{},
	"opencode": &OpenCodeWriter{},
}
