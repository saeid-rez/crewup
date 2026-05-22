package mcp

import "strings"

// InputField describes a user-supplied value required by a preset.
type InputField struct {
	Key       string // env var name, e.g. "CONTEXT7_API_KEY"
	Label     string // shown in UI, e.g. "Context7 API Key (optional)"
	Required  bool
	Sensitive bool // mask input (password mode)
}

// MCPPreset is a fully declarative, installable MCP server configuration.
type MCPPreset struct {
	ID          string
	Name        string
	Description string
	Command     string            // e.g. "npx"
	Args        []string          // e.g. ["-y", "@upstash/context7-mcp"]
	Env         map[string]string // static env vars (not user-supplied)
	Inputs      []InputField      // user-supplied values (prompted at install time)
}

// Presets is the master list of supported MCP server presets.
var Presets = []MCPPreset{
	{
		ID:          "context7",
		Name:        "Context7",
		Description: "Up-to-date docs for any library, directly in your AI tool",
		Command:     "npx",
		Args:        []string{"-y", "@upstash/context7-mcp"},
		Inputs: []InputField{
			{Key: "CONTEXT7_API_KEY", Label: "Context7 API Key (optional)", Required: false, Sensitive: true},
		},
	},
	// TODO: add Postgres, Slack, Notion, Linear, etc.
}

// FindByID returns a preset by ID.
func FindByID(id string) (MCPPreset, bool) {
	for _, p := range Presets {
		if p.ID == id {
			return p, true
		}
	}
	return MCPPreset{}, false
}

// buildConfig converts a preset + user values into tool-specific JSON.
// toolID: "copilot" | "opencode"
// values: map of InputField.Key → user-supplied string
func buildConfig(preset MCPPreset, toolID string, values map[string]string) map[string]interface{} {
	// Build the full command slice: [command, args...]
	command := append([]string{preset.Command}, preset.Args...)

	// Merge static env + user-supplied values
	envMerged := map[string]string{}
	for k, v := range preset.Env {
		envMerged[k] = v
	}
	for k, v := range values {
		if strings.TrimSpace(v) != "" {
			envMerged[k] = v
		}
	}

	if toolID == "opencode" {
		cfg := map[string]interface{}{
			"type":    "local",
			"command": toInterfaceSlice(command),
			"enabled": true,
		}
		if len(envMerged) > 0 {
			env := make(map[string]interface{}, len(envMerged))
			for k, v := range envMerged {
				env[k] = v
			}
			cfg["environment"] = env
		}
		return cfg
	}

	// Default (copilot / vscode) format
	cfg := map[string]interface{}{
		"command": command[0],
		"args":    toInterfaceSlice(command[1:]),
	}
	if len(envMerged) > 0 {
		env := make(map[string]interface{}, len(envMerged))
		for k, v := range envMerged {
			env[k] = v
		}
		cfg["env"] = env
	}
	return cfg
}

func toInterfaceSlice(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}
