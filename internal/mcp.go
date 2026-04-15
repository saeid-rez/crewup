package mcp

// Server represents a known MCP server crewup can install
type Server struct {
	ID          string
	Name        string
	Description string
	ConfigSnippet func(toolID string) map[string]interface{} // generates config per tool
}

// Registry is the master list of supported MCP servers
var Registry = []Server{
	{
		ID:          "context7",
		Name:        "Context7",
		Description: "Up-to-date docs for any library, directly in your AI tool",
		ConfigSnippet: func(toolID string) map[string]interface{} {
			return map[string]interface{}{
				"command": "npx",
				"args":    []string{"-y", "@upstash/context7-mcp"},
			}
		},
	},
	{
		ID:          "filesystem",
		Name:        "Filesystem",
		Description: "Read/write local files safely from your AI tool",
		ConfigSnippet: func(toolID string) map[string]interface{} {
			return map[string]interface{}{
				"command": "npx",
				"args":    []string{"-y", "@modelcontextprotocol/server-filesystem", "."},
			}
		},
	},
	{
		ID:          "github",
		Name:        "GitHub",
		Description: "Interact with GitHub repos, PRs and issues",
		ConfigSnippet: func(toolID string) map[string]interface{} {
			return map[string]interface{}{
				"command": "npx",
				"args":    []string{"-y", "@modelcontextprotocol/server-github"},
			}
		},
	},
	{
		ID:          "brave-search",
		Name:        "Brave Search",
		Description: "Web search via Brave from inside your AI tool",
		ConfigSnippet: func(toolID string) map[string]interface{} {
			return map[string]interface{}{
				"command": "npx",
				"args":    []string{"-y", "@modelcontextprotocol/server-brave-search"},
			}
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

// Install adds an MCP server config to the specified AI tool
func Install(serverID string, toolID string) error {
	// TODO:
	// 1. Find server in registry
	// 2. Load tool's current config file
	// 3. Inject ConfigSnippet into the right location
	// 4. Write config back to disk
	return nil
}
