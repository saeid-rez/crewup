package tools

import (
	"os/exec"
)

// AITool represents a supported AI tool crewup can configure
type AITool struct {
	ID         string // internal identifier e.g. "claude-cli"
	Name       string // display name e.g. "Claude CLI"
	ConfigPath string // where its config lives e.g. "~/.claude/config.json"
	CheckCmd   string // binary name to check if installed e.g. "claude"
}

// SupportedTools is the master list of tools crewup knows about
var SupportedTools = []AITool{
	{
		ID:         "copilot",
		Name:       "GitHub Copilot",
		ConfigPath: "~/.copilot/",
		CheckCmd:   "gh",
	},
	{
		ID:         "opencode",
		Name:       "OpenCode",
		ConfigPath: "~/.opencode/",
		CheckCmd:   "opencode",
	},
	// TODO: add Cursor, Windsurf, etc.
}

// DetectInstalledTools returns only tools that are installed on the machine
func DetectInstalledTools() []AITool {
	var detected []AITool

	for _, tool := range SupportedTools {
		if isInstalled(tool.CheckCmd) {
			detected = append(detected, tool)
		}
	}

	return detected
}

// isInstalled checks if a binary exists in PATH
func isInstalled(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
