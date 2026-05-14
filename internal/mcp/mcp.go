package mcp

import "strings"

// Server represents a known MCP server crewup can install
type Server struct {
	ID            string
	Name          string
	Description   string
	ConfigSnippet func(toolID string) func(opts InstallOptions) map[string]interface{} // generates config per tool
}

func localNPMConfig(toolID string, pkg string, extraArgs ...string) func(opts InstallOptions) map[string]interface{} {
	return func(opts InstallOptions) map[string]interface{} {
		command := append([]string{"npx", "-y", pkg}, extraArgs...)

		if toolID == "opencode" {
			cfg := map[string]interface{}{
				"type":    "local",
				"command": command,
				"enabled": true,
			}
			if apiKey := strings.TrimSpace(opts.Context7APIKey); apiKey != "" {
				cfg["environment"] = map[string]interface{}{
					"CONTEXT7_API_KEY": apiKey,
				}
			}
			return cfg
		}

		cfg := map[string]interface{}{
			"command": command[0],
			"args":    command[1:],
		}
		if apiKey := strings.TrimSpace(opts.Context7APIKey); apiKey != "" {
			cfg["env"] = map[string]interface{}{
				"CONTEXT7_API_KEY": apiKey,
			}
		}
		return cfg
	}
}

type InstallOptions struct {
	Context7APIKey string
}

// Registry is the master list of supported MCP servers
var Registry = []Server{
	{
		ID:          "context7",
		Name:        "Context7",
		Description: "Up-to-date docs for any library, directly in your AI tool",
		ConfigSnippet: func(toolID string) func(opts InstallOptions) map[string]interface{} {
			return localNPMConfig(toolID, "@upstash/context7-mcp")
		},
	},
	// TODO: add Postgres, Slack, Notion, Linear, etc.
}

// FindByID returns a server from the registry by ID
func FindByID(id string) (Server, bool) {
	for _, s := range Registry {
		if s.ID == id {
			return s, true
		}
	}
	return Server{}, false
}
