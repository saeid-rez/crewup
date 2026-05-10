package agentdefs

import (
	"strings"
	"testing"
)

func TestAll(t *testing.T) {
	defs, err := All()
	if err != nil {
		t.Fatalf("All() error: %v", err)
	}
	if len(defs) == 0 {
		t.Fatal("All() returned empty slice")
	}
	// Expect at least 8 agents
	if len(defs) < 8 {
		t.Errorf("expected at least 8 agents, got %d", len(defs))
	}
}

func TestByID(t *testing.T) {
	def, ok := ByID("planner")
	if !ok {
		t.Fatal("ByID('planner') returned false")
	}
	if def.ID != "planner" {
		t.Errorf("expected ID 'planner', got %q", def.ID)
	}
	if def.Name == "" {
		t.Error("Name is empty")
	}
	if def.DefaultModel == "" {
		t.Error("DefaultModel is empty")
	}

	_, ok = ByID("nonexistent-agent-xyz")
	if ok {
		t.Error("ByID('nonexistent') should return false")
	}
}

func TestFrontmatterParse(t *testing.T) {
	content := `---
id: test-agent
name: Test Agent
description: A test agent
mode: subagent
default_provider: Anthropic
default_model: claude-sonnet-4-5
temperature: 0.3
max_steps: 5
tools_allow: read,edit,bash
---

You are a test agent.
`
	def, err := parse(content)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if def.ID != "test-agent" {
		t.Errorf("ID: got %q, want 'test-agent'", def.ID)
	}
	if def.Name != "Test Agent" {
		t.Errorf("Name: got %q, want 'Test Agent'", def.Name)
	}
	if def.Mode != "subagent" {
		t.Errorf("Mode: got %q, want 'subagent'", def.Mode)
	}
	if def.Temperature != 0.3 {
		t.Errorf("Temperature: got %v, want 0.3", def.Temperature)
	}
	if def.MaxSteps != 5 {
		t.Errorf("MaxSteps: got %d, want 5", def.MaxSteps)
	}
	if len(def.ToolsAllow) != 3 {
		t.Errorf("ToolsAllow: got %v, want [read edit bash]", def.ToolsAllow)
	}
	if !strings.Contains(def.Prompt, "You are a test agent") {
		t.Errorf("Prompt missing expected content, got: %q", def.Prompt)
	}
}

func TestFrontmatterBodyWithDashes(t *testing.T) {
	// Body containing "---" should not break parsing
	content := `---
id: dash-agent
name: Dash Agent
description: Agent with dashes in body
default_provider: Anthropic
default_model: claude-sonnet-4-5
---

First paragraph.

---

Second paragraph after a horizontal rule.
`
	def, err := parse(content)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if !strings.Contains(def.Prompt, "First paragraph") {
		t.Errorf("Prompt missing 'First paragraph', got: %q", def.Prompt)
	}
	if !strings.Contains(def.Prompt, "Second paragraph") {
		t.Errorf("Prompt missing 'Second paragraph', got: %q", def.Prompt)
	}
}

func TestToolsAllowParsing(t *testing.T) {
	content := `---
id: tool-agent
name: Tool Agent
description: Agent with tools
default_provider: Anthropic
default_model: claude-sonnet-4-5
tools_allow: read, edit, search
---

Prompt.
`
	def, err := parse(content)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(def.ToolsAllow) != 3 {
		t.Errorf("expected 3 tools, got %d: %v", len(def.ToolsAllow), def.ToolsAllow)
	}
	// Spaces around commas should be trimmed
	for _, alias := range def.ToolsAllow {
		if alias != strings.TrimSpace(alias) {
			t.Errorf("alias %q has surrounding whitespace", alias)
		}
	}
}

func TestToolsAllowUnknownAlias(t *testing.T) {
	content := `---
id: bad-agent
name: Bad Agent
description: Agent with unknown tool
default_provider: Anthropic
default_model: claude-sonnet-4-5
tools_allow: read,unknown-tool
---

Prompt.
`
	_, err := parse(content)
	if err == nil {
		t.Fatal("expected error for unknown tool alias, got nil")
	}
	if !strings.Contains(err.Error(), "unknown tool alias") {
		t.Errorf("error should mention 'unknown tool alias', got: %v", err)
	}
}

func TestToolsAllowEmpty(t *testing.T) {
	content := `---
id: empty-tools-agent
name: Empty Tools Agent
description: Agent with empty tools_allow
default_provider: Anthropic
default_model: claude-sonnet-4-5
tools_allow:
---

Prompt.
`
	_, err := parse(content)
	if err == nil {
		t.Fatal("expected error for empty tools_allow, got nil")
	}
}

func TestToolsAllowAbsent(t *testing.T) {
	// Missing tools_allow should mean nil (all tools)
	content := `---
id: all-tools-agent
name: All Tools Agent
description: Agent without tools_allow
default_provider: Anthropic
default_model: claude-sonnet-4-5
---

Prompt.
`
	def, err := parse(content)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if def.ToolsAllow != nil {
		t.Errorf("expected nil ToolsAllow (all tools), got %v", def.ToolsAllow)
	}
}

func TestMissingIDDefaultsToFilenameStem(t *testing.T) {
	// When id is absent from frontmatter, parse() should succeed with empty ID,
	// and the caller (All / test code) applies the filename-stem default.
	content := `---
name: No ID Agent
description: Agent without an id field
default_provider: Anthropic
default_model: claude-sonnet-4-5
---

Prompt without id.
`
	def, err := parse(content)
	if err != nil {
		t.Fatalf("parse() should succeed when id is absent, got error: %v", err)
	}
	if def.ID != "" {
		t.Errorf("expected empty ID from parse() when id absent, got %q", def.ID)
	}

	// Simulate what All() does: apply filename-stem default
	stem := "no-id-agent"
	if def.ID == "" {
		def.ID = stem
	}
	if def.ID != stem {
		t.Errorf("after stem default, expected ID %q, got %q", stem, def.ID)
	}
}

func TestMissingRequiredField(t *testing.T) {
	content := `---
id: incomplete-agent
name: Incomplete Agent
---

Prompt.
`
	_, err := parse(content)
	if err == nil {
		t.Fatal("expected error for missing required fields, got nil")
	}
}

func TestWebToolAlias(t *testing.T) {
	content := `---
id: web-agent
name: Web Agent
description: Agent with web tool
default_provider: GitHub Copilot
default_model: claude-sonnet-4-5
tools_allow: read,search,web
---

Prompt.
`
	def, err := parse(content)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(def.ToolsAllow) != 3 {
		t.Errorf("expected 3 tools, got %d: %v", len(def.ToolsAllow), def.ToolsAllow)
	}
	found := false
	for _, alias := range def.ToolsAllow {
		if alias == "web" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'web' in ToolsAllow, got %v", def.ToolsAllow)
	}
}

func TestDisableModelInvocationFalse(t *testing.T) {
	content := `---
id: dmi-agent
name: DMI Agent
description: Agent with disable_model_invocation set to false
default_provider: GitHub Copilot
default_model: claude-sonnet-4-5
disable_model_invocation: false
---

Prompt.
`
	def, err := parse(content)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if def.DisableModelInvocation == nil {
		t.Fatal("expected DisableModelInvocation to be non-nil (explicitly set to false)")
	}
	if *def.DisableModelInvocation != false {
		t.Errorf("expected DisableModelInvocation = false, got %v", *def.DisableModelInvocation)
	}
}

func TestDisableModelInvocationTrue(t *testing.T) {
	content := `---
id: dmi-true-agent
name: DMI True Agent
description: Agent with disable_model_invocation set to true
default_provider: GitHub Copilot
default_model: claude-sonnet-4-5
disable_model_invocation: true
---

Prompt.
`
	def, err := parse(content)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if def.DisableModelInvocation == nil {
		t.Fatal("expected DisableModelInvocation to be non-nil (explicitly set to true)")
	}
	if *def.DisableModelInvocation != true {
		t.Errorf("expected DisableModelInvocation = true, got %v", *def.DisableModelInvocation)
	}
}

func TestDisableModelInvocationAbsent(t *testing.T) {
	content := `---
id: no-dmi-agent
name: No DMI Agent
description: Agent without disable_model_invocation
default_provider: GitHub Copilot
default_model: claude-sonnet-4-5
---

Prompt.
`
	def, err := parse(content)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if def.DisableModelInvocation != nil {
		t.Errorf("expected DisableModelInvocation to be nil when absent, got %v", *def.DisableModelInvocation)
	}
}

func TestSecurityReviewerTemplate(t *testing.T) {
	def, ok := ByID("security-reviewer")
	if !ok {
		t.Fatal("ByID('security-reviewer') returned false")
	}
	if def.ID != "security-reviewer" {
		t.Errorf("expected ID 'security-reviewer', got %q", def.ID)
	}
	if def.DisableModelInvocation == nil {
		t.Fatal("security-reviewer should have DisableModelInvocation set")
	}
	if *def.DisableModelInvocation != false {
		t.Errorf("security-reviewer should have DisableModelInvocation = false")
	}
	found := false
	for _, alias := range def.ToolsAllow {
		if alias == "web" {
			found = true
		}
	}
	if !found {
		t.Errorf("security-reviewer should have 'web' in ToolsAllow, got %v", def.ToolsAllow)
	}
}
