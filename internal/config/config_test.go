package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test that Load returns a wrapped error including the file path when JSON is malformed.
func TestLoadMalformedJSON(t *testing.T) {
	tmp, err := os.MkdirTemp("", "crewup-config-test-")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	defer os.RemoveAll(tmp)

	// Set HOME to tmp so ConfigDir() resolves to tmp/.crewup
	homeOrig := os.Getenv("HOME")
	defer os.Setenv("HOME", homeOrig)
	os.Setenv("HOME", tmp)

	cfgDir := filepath.Join(tmp, ".crewup")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatalf("mkdir cfg dir: %v", err)
	}
	bad := []byte("{ not-json }")
	path := filepath.Join(cfgDir, "config.json")
	if err := os.WriteFile(path, bad, 0644); err != nil {
		t.Fatalf("write bad config: %v", err)
	}

	_, err = Load()
	if err == nil {
		t.Fatalf("expected error loading malformed json, got nil")
	}
	if !contains(err.Error(), "config.json") {
		t.Fatalf("expected error message to mention config.json, got: %v", err)
	}
}

// TestModelSelectionJSONRoundtrip verifies that ModelSelection survives marshal/unmarshal.
func TestModelSelectionJSONRoundtrip(t *testing.T) {
	role := AgentRole{
		ID:          "planner",
		Name:        "Planner",
		Description: "Plans things",
		Prompt:      "You are a planner.",
		Mode:        "primary",
		Model: &ModelSelection{
			Provider:   "Anthropic",
			ModelID:    "claude-sonnet-4-5",
			Customized: true,
		},
	}

	data, err := json.Marshal(role)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded AgentRole
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.ID != role.ID {
		t.Errorf("ID: got %q, want %q", decoded.ID, role.ID)
	}
	if decoded.Mode != role.Mode {
		t.Errorf("Mode: got %q, want %q", decoded.Mode, role.Mode)
	}
	if decoded.Model == nil {
		t.Fatal("Model is nil after roundtrip")
	}
	if decoded.Model.Provider != role.Model.Provider {
		t.Errorf("Model.Provider: got %q, want %q", decoded.Model.Provider, role.Model.Provider)
	}
	if decoded.Model.ModelID != role.Model.ModelID {
		t.Errorf("Model.ModelID: got %q, want %q", decoded.Model.ModelID, role.Model.ModelID)
	}
	if decoded.Model.Customized != role.Model.Customized {
		t.Errorf("Model.Customized: got %v, want %v", decoded.Model.Customized, role.Model.Customized)
	}
}

// TestModelSelectionNilOmitted verifies that nil Model is omitted from JSON.
func TestModelSelectionNilOmitted(t *testing.T) {
	role := AgentRole{
		ID:   "planner",
		Name: "Planner",
	}

	data, err := json.Marshal(role)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if strings.Contains(string(data), `"model"`) {
		t.Error("nil Model should be omitted from JSON (omitempty)")
	}
}

// TestDefaultRolesBackwardCompat verifies that DefaultRoles is populated at init time
// and contains the expected agents.
func TestDefaultRolesBackwardCompat(t *testing.T) {
	if len(DefaultRoles) == 0 {
		t.Fatal("DefaultRoles is empty — init() failed to populate it")
	}

	// Check that known agents are present (documenter removed — not in .opencode/agents/)
	expected := []string{"planner", "implementor", "reviewer", "tester"}
	idSet := make(map[string]bool)
	for _, r := range DefaultRoles {
		idSet[r.ID] = true
	}
	for _, e := range expected {
		if !idSet[e] {
			t.Errorf("expected agent %q in DefaultRoles", e)
		}
	}
}

// TestLoadDefaultRolesReturnsCopies verifies that LoadDefaultRoles returns independent copies.
func TestLoadDefaultRolesReturnsCopies(t *testing.T) {
	roles1, err := LoadDefaultRoles()
	if err != nil {
		t.Fatalf("LoadDefaultRoles error: %v", err)
	}
	roles2, err := LoadDefaultRoles()
	if err != nil {
		t.Fatalf("LoadDefaultRoles error: %v", err)
	}

	// Mutating roles1 should not affect roles2
	if len(roles1) == 0 {
		t.Skip("no roles to test")
	}
	roles1[0].Name = "MUTATED"
	if roles2[0].Name == "MUTATED" {
		t.Error("LoadDefaultRoles returned shared slice — mutations affect other callers")
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
