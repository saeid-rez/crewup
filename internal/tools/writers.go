package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saeid-rez/crewup/internal/agentdefs"
	"github.com/saeid-rez/crewup/internal/config"
)

// renderToolsForCopilot maps crewup tool aliases to GitHub Copilot tool names.
// webfetch is omitted (unconfirmed for Copilot).
func renderToolsForCopilot(aliases []string) []string {
	var result []string
	for _, alias := range aliases {
		switch alias {
		case "read":
			result = append(result, "read")
		case "edit":
			result = append(result, "edit")
		case "search":
			result = append(result, "search")
		case "bash":
			result = append(result, "execute")
		case "web":
			result = append(result, "web")
		case "task":
			result = append(result, "agent")
			// webfetch: omitted — unconfirmed for Copilot
		}
	}
	return result
}

// effectiveModel returns the model ID to use for a role.
// If role.Model is set (user customized), use that. Otherwise use the AgentDef default.
func effectiveModel(role config.AgentRole) string {
	if role.Model != nil && role.Model.ModelID != "" {
		return role.Model.ModelID
	}
	def, ok := agentdefs.ByID(role.ID)
	if ok && def.DefaultModel != "" {
		return def.DefaultModel
	}
	return ""
}

// effectiveToolsAllow returns the tools_allow aliases for a role.
// Falls back to the AgentDef's ToolsAllow if available.
func effectiveToolsAllow(role config.AgentRole) []string {
	def, ok := agentdefs.ByID(role.ID)
	if ok && def.ToolsAllow != nil {
		return def.ToolsAllow
	}
	return nil // nil = all tools
}

// effectiveMaxSteps returns the max_steps for a role from its AgentDef.
func effectiveMaxSteps(role config.AgentRole) int {
	def, ok := agentdefs.ByID(role.ID)
	if ok && def.MaxSteps > 0 {
		return def.MaxSteps
	}
	return 8
}

// CopilotWriter writes agent configs for GitHub Copilot.
// Agent files → ~/.copilot/agents/<id>.agent.md (user-level, always)
//
//	→ .github/agents/<id>.agent.md (repo-level, only if .git exists in cwd)
type CopilotWriter struct{}

func (w *CopilotWriter) WriteAgentConfig(configPath string, roles []config.AgentRole) error {
	// Determine user-level agents dir
	expanded, err := ExpandHome(configPath)
	if err != nil {
		return err
	}
	userAgentsDir := filepath.Join(expanded, "agents")
	if err := EnsureDir(userAgentsDir); err != nil {
		return fmt.Errorf("creating ~/.copilot/agents dir: %w", err)
	}

	// Check if we're inside a git repo (repo-level write)
	_, gitErr := os.Stat(".git")
	inGitRepo := gitErr == nil

	var repoAgentsDir string
	if inGitRepo {
		repoAgentsDir = filepath.Join(".github", "agents")
		if err := EnsureDir(repoAgentsDir); err != nil {
			// Non-fatal: log and skip repo-level
			fmt.Printf("  ⚠️  GitHub Copilot: could not create .github/agents: %v\n", err)
			inGitRepo = false
		}
	}

	var errs []error
	for _, role := range roles {
		content := renderCopilotAgentFile(role)
		filename := role.ID + ".agent.md"

		// Write user-level
		userPath := filepath.Join(userAgentsDir, filename)
		if err := WriteFileWithBackup(userPath, []byte(content), 0644); err != nil {
			errs = append(errs, fmt.Errorf("writing Copilot user-level agent file for %s: %w", role.ID, err))
		} else {
			fmt.Printf("  ✅ GitHub Copilot: agent %s written to %s\n", role.ID, userPath)
		}

		// Write repo-level (only if inside a git repo)
		if inGitRepo {
			repoPath := filepath.Join(repoAgentsDir, filename)
			if err := WriteFileWithBackup(repoPath, []byte(content), 0644); err != nil {
				errs = append(errs, fmt.Errorf("writing Copilot repo-level agent file for %s: %w", role.ID, err))
			} else {
				fmt.Printf("  ✅ GitHub Copilot: agent %s written to %s\n", role.ID, repoPath)
			}
		}
	}

	return joinErrors(errs)
}

func renderCopilotAgentFile(role config.AgentRole) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", role.ID))
	sb.WriteString(fmt.Sprintf("description: %s\n", role.Description))

	model := effectiveModel(role)
	// Strip "github-copilot/" prefix for Copilot model field
	copilotModel := model
	if strings.HasPrefix(copilotModel, "github-copilot/") {
		copilotModel = strings.TrimPrefix(copilotModel, "github-copilot/")
	}
	if copilotModel != "" {
		sb.WriteString(fmt.Sprintf("model: %s\n", copilotModel))
	}

	// Emit disable-model-invocation only when explicitly set in the AgentDef
	if def, ok := agentdefs.ByID(role.ID); ok && def.DisableModelInvocation != nil {
		sb.WriteString(fmt.Sprintf("disable-model-invocation: %v\n", *def.DisableModelInvocation))
	}

	aliases := effectiveToolsAllow(role)
	if aliases != nil {
		copilotTools := renderToolsForCopilot(aliases)
		if len(copilotTools) > 0 {
			quoted := make([]string, len(copilotTools))
			for i, t := range copilotTools {
				quoted[i] = fmt.Sprintf("%q", t)
			}
			sb.WriteString(fmt.Sprintf("tools: [%s]\n", strings.Join(quoted, ", ")))
		}
	}

	sb.WriteString("---\n")
	if role.Prompt != "" {
		sb.WriteString("\n")
		sb.WriteString(role.Prompt)
		sb.WriteString("\n")
	}
	return sb.String()
}

// OpenCodeWriter writes agent configs for OpenCode.
// Agent files → .opencode/agents/<id>.md (repo-level, relative to cwd)
type OpenCodeWriter struct{}

func (w *OpenCodeWriter) WriteAgentConfig(configPath string, roles []config.AgentRole) error {
	agentsDir := filepath.Join(".opencode", "agents")
	if err := EnsureDir(agentsDir); err != nil {
		return fmt.Errorf("creating .opencode/agents dir: %w", err)
	}

	var errs []error
	for _, role := range roles {
		content := renderOpenCodeAgentFile(role)
		path := filepath.Join(agentsDir, role.ID+".md")
		if err := WriteFileWithBackup(path, []byte(content), 0644); err != nil {
			errs = append(errs, fmt.Errorf("writing OpenCode agent file for %s: %w", role.ID, err))
			continue
		}
		fmt.Printf("  ✅ OpenCode: agent %s written to %s\n", role.ID, path)
	}

	return joinErrors(errs)
}

func renderOpenCodeAgentFile(role config.AgentRole) string {
	def, hasDef := agentdefs.ByID(role.ID)

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("description: %s\n", role.Description))

	// mode: primary for planner, subagent for all others
	mode := "subagent"
	if hasDef && def.Mode == "primary" {
		mode = "primary"
	}
	sb.WriteString(fmt.Sprintf("mode: %s\n", mode))

	// model: OpenCode uses the full provider-prefixed model ID
	model := effectiveModel(role)
	if model != "" {
		sb.WriteString(fmt.Sprintf("model: %s\n", model))
	}

	// temperature
	temp := 0.1
	if hasDef {
		temp = def.Temperature
	}
	sb.WriteString(fmt.Sprintf("temperature: %g\n", temp))

	// max_steps
	sb.WriteString(fmt.Sprintf("max_steps: %d\n", effectiveMaxSteps(role)))

	// permission block derived from tools_allow aliases
	aliases := effectiveToolsAllow(role)
	aliasSet := make(map[string]bool, len(aliases))
	for _, a := range aliases {
		aliasSet[a] = true
	}

	sb.WriteString("permission:\n")

	// edit: full allow when edit is in tools_allow (e.g. implementor);
	// otherwise restrict to ask-by-default but always allow WORKFLOW_STATE.md.
	if aliasSet["edit"] {
		sb.WriteString("  edit: allow\n")
	} else {
		sb.WriteString("  edit:\n")
		sb.WriteString("    \"*\": ask\n")
		sb.WriteString("    \"WORKFLOW_STATE.md\": allow\n")
	}

	// bash
	if aliasSet["bash"] {
		sb.WriteString("  bash: allow\n")
	} else {
		sb.WriteString("  bash: deny\n")
	}

	// webfetch
	if aliasSet["web"] || aliasSet["webfetch"] {
		sb.WriteString("  webfetch: allow\n")
	}

	// task
	if aliasSet["task"] {
		sb.WriteString("  task: allow\n")
	}

	sb.WriteString("---\n")
	if role.Prompt != "" {
		sb.WriteString("\n")
		sb.WriteString(role.Prompt)
		sb.WriteString("\n")
	}
	return sb.String()
}

// joinErrors combines multiple errors into one, or returns nil if all are nil.
func joinErrors(errs []error) error {
	var nonNil []error
	for _, e := range errs {
		if e != nil {
			nonNil = append(nonNil, e)
		}
	}
	if len(nonNil) == 0 {
		return nil
	}
	msg := ""
	for _, e := range nonNil {
		msg += e.Error() + "; "
	}
	return fmt.Errorf("%s", msg[:len(msg)-2])
}
