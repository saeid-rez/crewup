package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/saeid-rez/crewup/internal/tools"
)

// AgentRole represents a role in the crew workflow
type AgentRole struct {
	ID          string
	Name        string
	Description string
	Prompt      string // system prompt injected for this role
}

// DefaultRoles is the built-in set of agent roles
var DefaultRoles = []AgentRole{
	{
		ID:          "planner",
		Name:        "Planner",
		Description: "Breaks down tasks into clear, actionable steps",
		Prompt:      "You are a planning agent. Your job is to analyze the request and produce a clear, step-by-step plan before any code is written.",
	},
	{
		ID:          "implementor",
		Name:        "Implementor",
		Description: "Executes the plan and writes code",
		Prompt:      "You are an implementation agent. You receive a plan and execute it precisely, writing clean, well-structured code.",
	},
	{
		ID:          "reviewer",
		Name:        "Reviewer",
		Description: "Reviews output for quality, bugs and improvements",
		Prompt:      "You are a code review agent. Critically review the implementation for bugs, edge cases, and improvements.",
	},
	{
		ID:          "tester",
		Name:        "Tester",
		Description: "Writes and runs tests for the implementation",
		Prompt:      "You are a testing agent. Write comprehensive tests for the implementation, covering edge cases and failure modes.",
	},
	{
		ID:          "documenter",
		Name:        "Documenter",
		Description: "Writes clear documentation for the output",
		Prompt:      "You are a documentation agent. Write clear, concise documentation for the implementation.",
	},
}

// CrewConfig is the top-level config written to ~/.crewup/config.json
type CrewConfig struct {
	Version    string            `json:"version"`
	Tools      []tools.AITool   `json:"tools"`
	AgentRoles []AgentRole       `json:"agent_roles"`
	MCPServers []string          `json:"mcp_servers"`
}

// NewCrewConfig builds a config from user selections
func NewCrewConfig(selectedTools []tools.AITool, roles []AgentRole, mcpServers []string) *CrewConfig {
	return &CrewConfig{
		Version:    "1",
		Tools:      selectedTools,
		AgentRoles: roles,
		MCPServers: mcpServers,
	}
}

// Apply writes the config to disk and applies changes to each tool
func (c *CrewConfig) Apply() error {
	if err := c.save(); err != nil {
		return err
	}

	// TODO: for each selected tool, apply agent configs and MCP configs
	// e.g. write CLAUDE.md files, update mcp_servers.json, etc.

	return nil
}

// ConfigDir returns the crewup config directory (~/.crewup)
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".crewup"), nil
}

func (c *CrewConfig) save() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create config dir: %w", err)
	}

	path := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
