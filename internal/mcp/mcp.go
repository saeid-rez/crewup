package mcp

// Server represents a known MCP server crewup can install
type Server struct {
	ID            string
	Name          string
	Description   string
	ConfigSnippet func(toolID string) map[string]interface{} // generates config per tool
}

func localNPMConfig(toolID string, pkg string, extraArgs ...string) map[string]interface{} {
	command := append([]string{"npx", "-y", pkg}, extraArgs...)

	if toolID == "opencode" {
		return map[string]interface{}{
			"type":    "local",
			"command": command,
			"enabled": true,
		}
	}

	return map[string]interface{}{
		"command": command[0],
		"args":    command[1:],
	}
}

// Registry is the master list of supported MCP servers
var Registry = []Server{
	{
		ID:          "context7",
		Name:        "Context7",
		Description: "Up-to-date docs for any library, directly in your AI tool",
		ConfigSnippet: func(toolID string) map[string]interface{} {
			return localNPMConfig(toolID, "@upstash/context7-mcp")
		},
	},
	{
		ID:          "filesystem",
		Name:        "Filesystem",
		Description: "Read/write local files safely from your AI tool",
		ConfigSnippet: func(toolID string) map[string]interface{} {
			return localNPMConfig(toolID, "@modelcontextprotocol/server-filesystem", ".")
		},
	},
	{
		ID:          "github",
		Name:        "GitHub",
		Description: "Interact with GitHub repos, PRs and issues",
		ConfigSnippet: func(toolID string) map[string]interface{} {
			return localNPMConfig(toolID, "@modelcontextprotocol/server-github")
		},
	},
	{
		ID:          "brave-search",
		Name:        "Brave Search",
		Description: "Web search via Brave from inside your AI tool",
		ConfigSnippet: func(toolID string) map[string]interface{} {
			return localNPMConfig(toolID, "@modelcontextprotocol/server-brave-search")
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
