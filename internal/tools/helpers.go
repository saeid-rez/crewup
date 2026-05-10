package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExpandHome replaces a leading ~ with the user's home directory.
func ExpandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, path[1:]), nil
}

// WriteFileWithBackup writes data to path, first backing up any existing file.
// If the write fails, it attempts to restore the backup.
// Preserves the existing file's permissions if the file already exists.
func WriteFileWithBackup(path string, data []byte, perm os.FileMode) error {
	backupPath := path + ".crewup.bak"

	// Backup existing file if it exists, and capture its permissions.
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
		if existing, err := os.ReadFile(path); err == nil {
			_ = os.WriteFile(backupPath, existing, perm)
		}
	}

	if err := os.WriteFile(path, data, perm); err != nil {
		// Attempt restore
		if backup, berr := os.ReadFile(backupPath); berr == nil {
			_ = os.WriteFile(path, backup, perm)
		}
		return err
	}
	return nil
}

// EnsureDir creates a directory (and parents) if it doesn't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// IndentPrompt indents each line of a prompt with two spaces (for YAML).
func IndentPrompt(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = "  " + l
	}
	return strings.Join(lines, "\n")
}

// RenderAgentMarkdown renders agent roles as a Markdown document.
// roles is passed as a slice of structs with ID, Name, Description, Prompt fields.
// We use interface{} to avoid circular imports — callers cast appropriately.
func RenderAgentMarkdown(roles []AgentRoleInfo) string {
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

// AgentRoleInfo is a minimal role descriptor used by writers to avoid import cycles.
type AgentRoleInfo struct {
	ID          string
	Name        string
	Description string
	Prompt      string
}
