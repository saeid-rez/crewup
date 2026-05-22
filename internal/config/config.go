package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saeid-rez/crewup/internal/agentdefs"
)

// ModelSelection holds the user's model choice for an agent.
// nil Model on AgentRole means "use the agent's default model".
type ModelSelection struct {
	Provider   string `json:"provider"`
	ModelID    string `json:"model_id"`
	Customized bool   `json:"customized"`
}

// AgentRole represents a role in the crew workflow
type AgentRole struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Prompt      string          `json:"prompt"` // system prompt injected for this role
	Mode        string          `json:"mode"`
	Model       *ModelSelection `json:"model,omitempty"`
}

// ToolInfo is a minimal tool descriptor stored in config (avoids import cycle with tools pkg).
type ToolInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ConfigPath string `json:"config_path"`
}

// DefaultRoles is the built-in set of agent roles, populated at init time.
// It is kept for backward compatibility with existing call sites.
// Use LoadDefaultRoles() for fresh copies.
var DefaultRoles []AgentRole

func init() {
	DefaultRoles = mustLoadDefaultRoles()
}

// mustLoadDefaultRoles loads default roles from embedded templates.
// It panics at startup if templates cannot be parsed (fail-fast).
func mustLoadDefaultRoles() []AgentRole {
	roles, err := LoadDefaultRoles()
	if err != nil {
		panic(fmt.Sprintf("crewup: failed to load agent templates: %v", err))
	}
	return roles
}

// LoadDefaultRoles parses agent templates and returns the default roles.
// Returns copies so callers cannot mutate shared state.
func LoadDefaultRoles() ([]AgentRole, error) {
	defs, err := agentdefs.All()
	if err != nil {
		return nil, err
	}
	roles := make([]AgentRole, len(defs))
	for i, d := range defs {
		roles[i] = AgentRole{
			ID:          d.ID,
			Name:        d.Name,
			Description: d.Description,
			Prompt:      d.Prompt,
			Mode:        d.Mode,
		}
	}
	return roles, nil
}

// CrewConfig is the top-level config written to ~/.crewup/config.json
type CrewConfig struct {
	Version    string      `json:"version"`
	Tools      []ToolInfo  `json:"tools"`
	AgentRoles []AgentRole `json:"agent_roles"`
	MCPServers []string    `json:"mcp_servers"`
}

// NewCrewConfig builds a config from user selections
func NewCrewConfig(selectedTools []ToolInfo, roles []AgentRole, mcpServers []string) *CrewConfig {
	return &CrewConfig{
		Version:    "1",
		Tools:      selectedTools,
		AgentRoles: roles,
		MCPServers: mcpServers,
	}
}

// Apply writes the config to disk and applies agent config changes to each tool.
// MCP installs are handled explicitly by the caller (cmd/init.go, cmd/add.go).
// Partial failures are aggregated — all tools are attempted even if one fails.
func (c *CrewConfig) Apply(
	writeAgentConfig func(toolID, configPath string, roles []AgentRole) error,
) error {
	if err := c.save(); err != nil {
		return err
	}

	var errs []error

	for _, tool := range c.Tools {
		if err := writeAgentConfig(tool.ID, tool.ConfigPath, c.AgentRoles); err != nil {
			fmt.Printf("  ⚠️  %s: agent config skipped (%v)\n", tool.Name, err)
			errs = append(errs, fmt.Errorf("%s agent config: %w", tool.Name, err))
		} else {
			fmt.Printf("  ✅ %s configured\n", tool.Name)
		}
	}

	return errors.Join(errs...)
}

// Load reads the config from ~/.crewup/config.json
func Load() (*CrewConfig, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg CrewConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &cfg, nil
}

// ConfigDir returns the crewup config directory (~/.crewup)
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".crewup"), nil
}

// Save writes the config to disk without applying any tool configurations.
// Use this when you only need to persist config.json changes (e.g. after add mcp).
func (c *CrewConfig) Save() error {
	return c.save()
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

	// Preserve existing file permissions if the file already exists.
	perm := os.FileMode(0644)
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
	}

	return os.WriteFile(path, data, perm)
}

// RenderAgentMarkdown renders agent roles as a Markdown document.
func RenderAgentMarkdown(roles []AgentRole) string {
	var sb strings.Builder
	sb.WriteString("# Agent Roles (managed by crewup)\n\n")
	sb.WriteString("> This file is managed by crewup. Run `crewup init` to regenerate.\n\n")
	for _, role := range roles {
		sb.WriteString(fmt.Sprintf("## %s\n\n", role.Name))
		sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", role.Description))
		sb.WriteString(fmt.Sprintf("**System Prompt:**\n\n```\n%s\n```\n\n", role.Prompt))
	}
	return sb.String()
}
