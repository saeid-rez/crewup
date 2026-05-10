package ui

import (
	"testing"

	"github.com/saeid-rez/crewup/internal/config"
	"github.com/saeid-rez/crewup/internal/models"
)

// TestConfigureAgentsNonTTY verifies that ConfigureAgents returns all roles
// unchanged (Model=nil) when not running in a TTY.
// In test environments, isTTY() returns false (stdin/stdout are pipes).
func TestConfigureAgentsNonTTY(t *testing.T) {
	roles := []config.AgentRole{
		{ID: "planner", Name: "Planner", Description: "Plans things"},
		{ID: "implementor", Name: "Implementor", Description: "Implements things"},
	}
	allModels := models.All()

	result, err := ConfigureAgents(roles, allModels)
	if err != nil {
		t.Fatalf("ConfigureAgents error: %v", err)
	}
	if len(result) != len(roles) {
		t.Fatalf("expected %d roles, got %d", len(roles), len(result))
	}
	for i, r := range result {
		if r.Model != nil {
			t.Errorf("role[%d] %q: expected Model=nil in non-TTY mode, got %+v", i, r.ID, r.Model)
		}
	}
}

// TestConfigureAgentNonTTY verifies that ConfigureAgent returns the role
// unchanged when not running in a TTY.
func TestConfigureAgentNonTTY(t *testing.T) {
	role := config.AgentRole{
		ID:          "planner",
		Name:        "Planner",
		Description: "Plans things",
	}
	allModels := models.All()

	result, err := ConfigureAgent(role, allModels)
	if err != nil {
		t.Fatalf("ConfigureAgent error: %v", err)
	}
	if result.Model != nil {
		t.Errorf("expected Model=nil in non-TTY mode, got %+v", result.Model)
	}
	if result.ID != role.ID {
		t.Errorf("expected ID %q, got %q", role.ID, result.ID)
	}
}
