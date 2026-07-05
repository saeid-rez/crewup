package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saeid-rez/crewup/internal/config"
)

// testRole returns a sample AgentRole for testing.
func testRole() config.AgentRole {
	return config.AgentRole{
		ID:          "planner",
		Name:        "Planner",
		Description: "Clarifies the request first, then creates a plan",
		Prompt:      "You are the planner agent.",
		Mode:        "primary",
	}
}

func testRoleWithModel() config.AgentRole {
	r := testRole()
	r.Model = &config.ModelSelection{
		Provider:   "Anthropic",
		ModelID:    "claude-sonnet-4-5",
		Customized: true,
	}
	return r
}

// --- renderToolsForCopilot ---

func TestRenderToolsForCopilot(t *testing.T) {
	aliases := []string{"read", "edit", "search", "bash", "webfetch", "task"}
	result := renderToolsForCopilot(aliases)

	// webfetch is omitted for Copilot
	expected := []string{"read", "edit", "search", "execute", "agent"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d tools, got %d: %v", len(expected), len(result), result)
	}
	for i, e := range expected {
		if result[i] != e {
			t.Errorf("index %d: expected %q, got %q", i, e, result[i])
		}
	}
}

func TestRenderToolsForCopilotNoWebfetch(t *testing.T) {
	// webfetch should be silently omitted
	aliases := []string{"webfetch"}
	result := renderToolsForCopilot(aliases)
	if len(result) != 0 {
		t.Errorf("expected empty result for webfetch-only, got %v", result)
	}
}

// --- Copilot writer golden test ---

func TestCopilotWriterGolden(t *testing.T) {
	role := testRoleWithModel()
	content := renderCopilotAgentFile(role)

	if !strings.HasPrefix(content, "---\n") {
		t.Error("Copilot agent file must start with '---'")
	}
	if !strings.Contains(content, "name: planner") {
		t.Error("missing 'name: planner'")
	}
	if !strings.Contains(content, "description:") {
		t.Error("missing description")
	}
	if !strings.Contains(content, "model:") {
		t.Error("missing model field")
	}
	if !strings.Contains(content, "You are the planner agent") {
		t.Error("missing prompt body")
	}
}

func TestCopilotWriterFilenameAndPath(t *testing.T) {
	// Verify that WriteAgentConfig writes to ~/.copilot/agents/<id>.agent.md
	// We use a temp dir as the configPath to avoid touching the real home dir.
	tmpDir := t.TempDir()

	// Create a fake .git dir so repo-level write is also exercised
	if err := os.MkdirAll(filepath.Join(tmpDir, ".git"), 0755); err != nil {
		t.Fatalf("creating fake .git: %v", err)
	}

	// Change working directory to tmpDir so .git detection works
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Use tmpDir as the configPath (simulates ~/.copilot/)
	w := &CopilotWriter{}
	roles := []config.AgentRole{testRole()}
	if err := w.WriteAgentConfig(tmpDir, roles); err != nil {
		t.Fatalf("WriteAgentConfig: %v", err)
	}

	// Verify user-level file: <configPath>/agents/planner.agent.md
	userFile := filepath.Join(tmpDir, "agents", "planner.agent.md")
	if _, err := os.Stat(userFile); err != nil {
		t.Errorf("user-level file not found at %s: %v", userFile, err)
	}

	// Verify filename uses .agent.md extension (not .md)
	if !strings.HasSuffix(userFile, ".agent.md") {
		t.Errorf("expected .agent.md extension, got %s", userFile)
	}

	// Verify repo-level file: .github/agents/planner.agent.md
	repoFile := filepath.Join(tmpDir, ".github", "agents", "planner.agent.md")
	if _, err := os.Stat(repoFile); err != nil {
		t.Errorf("repo-level file not found at %s: %v", repoFile, err)
	}
}

func TestCopilotWriterModelUsesDisplayName(t *testing.T) {
	role := testRole()
	role.Model = &config.ModelSelection{
		Provider:   "GitHub Copilot",
		ModelID:    "github-copilot/claude-sonnet-4.6",
		Customized: true,
	}
	content := renderCopilotAgentFile(role)
	if strings.Contains(content, "github-copilot/") {
		t.Error("Copilot model field should not contain 'github-copilot/' prefix")
	}
	if !strings.Contains(content, "Claude Sonnet 4.6 (Copilot)") {
		t.Error("Copilot model field should contain the model display name")
	}
}

func TestCopilotWriterUsesDefaultDisplayNameForUncustomizedRole(t *testing.T) {
	content := renderCopilotAgentFile(testRole())
	if !strings.Contains(content, "model: Claude Sonnet 4.6 (Copilot)") {
		t.Errorf("expected default Copilot display name in output, got:\n%s", content)
	}
}

// --- web alias tests ---

func TestRenderToolsForCopilotWebAlias(t *testing.T) {
	aliases := []string{"web"}
	result := renderToolsForCopilot(aliases)
	if len(result) != 1 || result[0] != "web" {
		t.Errorf("expected [\"web\"] for Copilot, got %v", result)
	}
}

// --- disable-model-invocation tests ---

func TestCopilotWriterEmitsDisableModelInvocationFalse(t *testing.T) {
	// security-reviewer template has disable_model_invocation: false
	role := config.AgentRole{
		ID:          "security-reviewer",
		Name:        "Security Reviewer",
		Description: "Performs security-focused code review",
		Prompt:      "You are the security reviewer.",
		Mode:        "subagent",
	}
	content := renderCopilotAgentFile(role)
	if !strings.Contains(content, "disable-model-invocation: false") {
		t.Errorf("expected 'disable-model-invocation: false' in Copilot output, got:\n%s", content)
	}
}

func TestCopilotWriterNoDisableModelInvocationWhenAbsent(t *testing.T) {
	// A role with an unknown ID (no AgentDef) should not emit disable-model-invocation
	role := config.AgentRole{
		ID:          "unknown-agent",
		Name:        "Unknown",
		Description: "An agent with no AgentDef",
		Prompt:      "You are an unknown agent.",
		Mode:        "subagent",
	}
	content := renderCopilotAgentFile(role)
	if strings.Contains(content, "disable-model-invocation") {
		t.Errorf("expected no 'disable-model-invocation' field when not set in AgentDef, got:\n%s", content)
	}
}

func TestCopilotWriterWebToolRendering(t *testing.T) {
	// security-reviewer uses web in tools_allow
	role := config.AgentRole{
		ID:          "security-reviewer",
		Name:        "Security Reviewer",
		Description: "Performs security-focused code review",
		Prompt:      "You are the security reviewer.",
		Mode:        "subagent",
	}
	content := renderCopilotAgentFile(role)
	if !strings.Contains(content, `"web"`) {
		t.Errorf("expected '\"web\"' in Copilot tools output for security-reviewer, got:\n%s", content)
	}
}

// --- OpenCode writer tests ---

func TestRenderOpenCodeAgentFileStructure(t *testing.T) {
	role := testRole() // planner — mode: primary
	content := renderOpenCodeAgentFile(role)

	checks := []string{
		"---\n",
		"description:",
		"mode: primary",
		"temperature:",
		"max_steps:",
		"permission:",
		"edit:",
		"bash:",
		"---\n",
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("OpenCode output missing %q:\n%s", want, content)
		}
	}
}

func TestRenderOpenCodeAgentFileSubagent(t *testing.T) {
	role := config.AgentRole{
		ID:          "reviewer",
		Name:        "Reviewer",
		Description: "Reviews the implementation",
		Prompt:      "You are the reviewer.",
	}
	content := renderOpenCodeAgentFile(role)
	if !strings.Contains(content, "mode: subagent") {
		t.Errorf("expected 'mode: subagent' for reviewer, got:\n%s", content)
	}
}

func TestRenderOpenCodeEditAllow(t *testing.T) {
	// implementor has edit in tools_allow → should get "edit: allow"
	role := config.AgentRole{
		ID:          "implementor",
		Name:        "Implementor",
		Description: "Implements the plan",
		Prompt:      "You are the implementor.",
	}
	content := renderOpenCodeAgentFile(role)
	if !strings.Contains(content, "  edit: allow") {
		t.Errorf("expected 'edit: allow' for implementor (has edit in tools_allow), got:\n%s", content)
	}
}

func TestRenderOpenCodeEditAsk(t *testing.T) {
	// reviewer has no edit in tools_allow (read,search,edit actually...)
	// Use security-reviewer which has read,search,web (no edit) → should get edit: ask map
	role := config.AgentRole{
		ID:          "security-reviewer",
		Name:        "Security Reviewer",
		Description: "Security review",
		Prompt:      "You are the security reviewer.",
	}
	content := renderOpenCodeAgentFile(role)
	if !strings.Contains(content, "\"*\": ask") {
		t.Errorf("expected edit ask map for security-reviewer, got:\n%s", content)
	}
	if !strings.Contains(content, "\"WORKFLOW_STATE.md\": allow") {
		t.Errorf("expected WORKFLOW_STATE.md allow for security-reviewer, got:\n%s", content)
	}
}

func TestRenderOpenCodeWebfetchPermission(t *testing.T) {
	// planner has web in tools_allow → should get webfetch: allow
	role := config.AgentRole{
		ID:          "planner",
		Name:        "Planner",
		Description: "Plans the work",
		Prompt:      "You are the planner.",
	}
	content := renderOpenCodeAgentFile(role)
	if !strings.Contains(content, "  webfetch: allow") {
		t.Errorf("expected 'webfetch: allow' for planner (has web in tools_allow), got:\n%s", content)
	}
}

func TestRenderOpenCodeBashDeny(t *testing.T) {
	// reviewer has no bash in tools_allow → bash: deny
	role := config.AgentRole{
		ID:          "reviewer",
		Name:        "Reviewer",
		Description: "Reviews the implementation",
		Prompt:      "You are the reviewer.",
	}
	content := renderOpenCodeAgentFile(role)
	if !strings.Contains(content, "  bash: deny") {
		t.Errorf("expected 'bash: deny' for reviewer (no bash in tools_allow), got:\n%s", content)
	}
}

func TestOpenCodeWriterCreatesFile(t *testing.T) {
	dir := t.TempDir()
	// Change to temp dir so .opencode/agents is created there
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	w := &OpenCodeWriter{}
	roles := []config.AgentRole{
		{ID: "planner", Name: "Planner", Description: "Plans the work", Prompt: "You are the planner."},
	}
	if err := w.WriteAgentConfig("~/.opencode/", roles); err != nil {
		t.Fatalf("WriteAgentConfig failed: %v", err)
	}

	path := filepath.Join(dir, ".opencode", "agents", "planner.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected file at %s: %v", path, err)
	}
	if !strings.Contains(string(data), "mode: primary") {
		t.Errorf("generated file missing 'mode: primary':\n%s", string(data))
	}
}
