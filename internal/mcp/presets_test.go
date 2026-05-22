package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPresetsNotEmpty(t *testing.T) {
	if len(Presets) == 0 {
		t.Fatal("Presets is empty — expected at least one entry")
	}
}

func TestFindByID(t *testing.T) {
	p, ok := FindByID("context7")
	if !ok {
		t.Fatal("expected to find context7 preset")
	}
	if p.ID != "context7" {
		t.Fatalf("expected ID %q, got %q", "context7", p.ID)
	}

	_, ok = FindByID("nonexistent-preset-xyz")
	if ok {
		t.Fatal("expected FindByID to return false for unknown ID")
	}

	p, ok = FindByID("serena")
	if !ok {
		t.Fatal("expected to find serena preset")
	}
	if p.ID != "serena" {
		t.Fatalf("expected ID %q, got %q", "serena", p.ID)
	}
}

func TestContext7InputFields(t *testing.T) {
	p, ok := FindByID("context7")
	if !ok {
		t.Fatal("context7 preset not found")
	}
	if len(p.Inputs) != 1 {
		t.Fatalf("expected 1 InputField, got %d", len(p.Inputs))
	}
	f := p.Inputs[0]
	if f.Key != "CONTEXT7_API_KEY" {
		t.Errorf("expected Key %q, got %q", "CONTEXT7_API_KEY", f.Key)
	}
	if !f.Sensitive {
		t.Error("expected Sensitive == true")
	}
	if f.Required {
		t.Error("expected Required == false")
	}
}

func TestBuildConfigCopilot(t *testing.T) {
	p, _ := FindByID("context7")
	values := map[string]string{"CONTEXT7_API_KEY": "test-key"}
	cfg := buildConfig(p, "copilot", values)

	if cfg["command"] != "npx" {
		t.Errorf("expected command %q, got %v", "npx", cfg["command"])
	}
	args, ok := cfg["args"].([]interface{})
	if !ok {
		t.Fatalf("expected args []interface{}, got %T", cfg["args"])
	}
	if len(args) < 2 || args[0] != "-y" || args[1] != "@upstash/context7-mcp" {
		t.Errorf("unexpected args: %v", args)
	}
	env, ok := cfg["env"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected env map, got %T", cfg["env"])
	}
	if env["CONTEXT7_API_KEY"] != "test-key" {
		t.Errorf("expected CONTEXT7_API_KEY %q, got %v", "test-key", env["CONTEXT7_API_KEY"])
	}
}

func TestBuildConfigOpencode(t *testing.T) {
	p, _ := FindByID("context7")
	values := map[string]string{"CONTEXT7_API_KEY": "test-key"}
	cfg := buildConfig(p, "opencode", values)

	if cfg["type"] != "local" {
		t.Errorf("expected type %q, got %v", "local", cfg["type"])
	}
	if cfg["enabled"] != true {
		t.Errorf("expected enabled true, got %v", cfg["enabled"])
	}
	command, ok := cfg["command"].([]interface{})
	if !ok {
		t.Fatalf("expected command []interface{}, got %T", cfg["command"])
	}
	if len(command) < 3 || command[0] != "npx" || command[1] != "-y" || command[2] != "@upstash/context7-mcp" {
		t.Errorf("unexpected command: %v", command)
	}
	env, ok := cfg["environment"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected environment map, got %T", cfg["environment"])
	}
	if env["CONTEXT7_API_KEY"] != "test-key" {
		t.Errorf("expected CONTEXT7_API_KEY %q, got %v", "test-key", env["CONTEXT7_API_KEY"])
	}
}

func TestBuildConfigNoValues(t *testing.T) {
	p, _ := FindByID("context7")
	cfg := buildConfig(p, "copilot", map[string]string{})

	if _, hasEnv := cfg["env"]; hasEnv {
		t.Error("expected no env key when values are empty")
	}

	cfg2 := buildConfig(p, "opencode", map[string]string{})
	if _, hasEnv := cfg2["environment"]; hasEnv {
		t.Error("expected no environment key when values are empty")
	}
}

func TestBuildConfigSerenaOpencode(t *testing.T) {
	p, ok := FindByID("serena")
	if !ok {
		t.Fatal("serena preset not found")
	}

	cfg := buildConfig(p, "opencode", map[string]string{})

	if cfg["type"] != "local" {
		t.Errorf("expected type %q, got %v", "local", cfg["type"])
	}
	if cfg["enabled"] != true {
		t.Errorf("expected enabled true, got %v", cfg["enabled"])
	}
	command, ok := cfg["command"].([]interface{})
	if !ok {
		t.Fatalf("expected command []interface{}, got %T", cfg["command"])
	}
	want := []interface{}{
		"uvx",
		"--from",
		"git+https://github.com/oraios/serena",
		"serena",
		"start-mcp-server",
		"--context",
		"ide",
		"--project-from-cwd",
	}
	if len(command) != len(want) {
		t.Fatalf("expected command length %d, got %d: %v", len(want), len(command), command)
	}
	for i := range want {
		if command[i] != want[i] {
			t.Fatalf("expected command[%d] = %v, got %v", i, want[i], command[i])
		}
	}
	if _, hasEnv := cfg["environment"]; hasEnv {
		t.Error("expected no environment key for serena without values")
	}
}

func TestBuildConfigSerenaCopilot(t *testing.T) {
	p, ok := FindByID("serena")
	if !ok {
		t.Fatal("serena preset not found")
	}

	cfg := buildConfig(p, "copilot", map[string]string{})

	if cfg["command"] != "uvx" {
		t.Errorf("expected command %q, got %v", "uvx", cfg["command"])
	}
	args, ok := cfg["args"].([]interface{})
	if !ok {
		t.Fatalf("expected args []interface{}, got %T", cfg["args"])
	}
	want := []interface{}{
		"--from",
		"git+https://github.com/oraios/serena",
		"serena",
		"start-mcp-server",
		"--context",
		"ide",
		"--project-from-cwd",
	}
	if len(args) != len(want) {
		t.Fatalf("expected args length %d, got %d: %v", len(want), len(args), args)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("expected args[%d] = %v, got %v", i, want[i], args[i])
		}
	}
	if _, hasEnv := cfg["env"]; hasEnv {
		t.Error("expected no env key for serena without values")
	}
}

func TestInstallTargeting(t *testing.T) {
	tmp, err := os.MkdirTemp("", "crewup-mcp-targeting-")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	defer os.RemoveAll(tmp)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Install only for copilot (project scope)
	if err := Install("context7", "copilot", ScopeProject, InstallOptions{Values: map[string]string{}}); err != nil {
		t.Fatalf("install copilot: %v", err)
	}

	// Assert copilot config was written
	vscodePath := filepath.Join(tmp, ".vscode", "mcp.json")
	if _, err := os.Stat(vscodePath); os.IsNotExist(err) {
		t.Fatalf("expected .vscode/mcp.json to exist after copilot install")
	}

	// Assert opencode config was NOT written
	opencodePath := filepath.Join(tmp, "opencode.json")
	if _, err := os.Stat(opencodePath); !os.IsNotExist(err) {
		t.Fatalf("expected opencode.json to NOT exist when only copilot was targeted")
	}
}
