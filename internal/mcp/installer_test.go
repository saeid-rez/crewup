package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Test wrong-type error for VSCodeMCPMerger (project scope)
func TestMergersWrongTypeError(t *testing.T) {
	tmp, err := os.MkdirTemp("", "crewup-mcp-wrongtype-")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	defer os.RemoveAll(tmp)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	srv, _ := FindByID("context7")

	// VSCode (project) — .vscode/mcp.json with servers as array
	vscodeDir := filepath.Join(tmp, ".vscode")
	if err := os.MkdirAll(vscodeDir, 0755); err != nil {
		t.Fatalf("mkdir .vscode: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vscodeDir, "mcp.json"), []byte(`{"servers": []}`), 0644); err != nil {
		t.Fatalf("write vscode mcp.json: %v", err)
	}
	if err := (&VSCodeMCPMerger{}).Merge(srv, ScopeProject, InstallOptions{}); err == nil {
		t.Fatalf("expected error for wrong-type vscode, got nil")
	}
}

func TestOpenCodeMergeProject(t *testing.T) {
	tmp, err := os.MkdirTemp("", "crewup-mcp-opencode-project-")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	defer os.RemoveAll(tmp)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	srv, _ := FindByID("context7")
	if err := (&OpenCodeMCPMerger{}).Merge(srv, ScopeProject, InstallOptions{Context7APIKey: "ctx7sk-test"}); err != nil {
		t.Fatalf("merge opencode project: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, "opencode.json"))
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal opencode.json: %v", err)
	}

	mcpMap, ok := cfg["mcp"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected mcp object, got %T", cfg["mcp"])
	}

	context7, ok := mcpMap["context7"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected context7 object, got %T", mcpMap["context7"])
	}

	if got, want := context7["type"], "local"; got != want {
		t.Fatalf("expected type %q, got %#v", want, got)
	}
	if got, want := context7["enabled"], true; got != want {
		t.Fatalf("expected enabled %v, got %#v", want, got)
	}

	environment, ok := context7["environment"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected environment object, got %T", context7["environment"])
	}
	if got, want := environment["CONTEXT7_API_KEY"], "ctx7sk-test"; got != want {
		t.Fatalf("expected CONTEXT7_API_KEY %q, got %#v", want, got)
	}

	command, ok := context7["command"].([]interface{})
	if !ok {
		t.Fatalf("expected command array, got %T", context7["command"])
	}
	if len(command) != 3 || command[0] != "npx" || command[1] != "-y" || command[2] != "@upstash/context7-mcp" {
		t.Fatalf("unexpected command array: %#v", command)
	}
}

func TestOpenCodeMergeWrongTypeError(t *testing.T) {
	tmp, err := os.MkdirTemp("", "crewup-mcp-opencode-wrongtype-")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	defer os.RemoveAll(tmp)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmp, "opencode.json"), []byte(`{"mcp": []}`), 0644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	srv, _ := FindByID("context7")
	if err := (&OpenCodeMCPMerger{}).Merge(srv, ScopeProject, InstallOptions{}); err == nil {
		t.Fatalf("expected error for wrong-type opencode mcp, got nil")
	}
}

func TestVSCodeMergeIncludesContext7Env(t *testing.T) {
	tmp, err := os.MkdirTemp("", "crewup-mcp-vscode-env-")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	defer os.RemoveAll(tmp)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	srv, _ := FindByID("context7")
	if err := (&VSCodeMCPMerger{}).Merge(srv, ScopeProject, InstallOptions{Context7APIKey: "ctx7sk-test"}); err != nil {
		t.Fatalf("merge vscode project: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, ".vscode", "mcp.json"))
	if err != nil {
		t.Fatalf("read .vscode/mcp.json: %v", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal mcp.json: %v", err)
	}

	servers, ok := cfg["servers"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected servers object, got %T", cfg["servers"])
	}

	context7, ok := servers["context7"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected context7 object, got %T", servers["context7"])
	}

	env, ok := context7["env"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected env object, got %T", context7["env"])
	}
	if got, want := env["CONTEXT7_API_KEY"], "ctx7sk-test"; got != want {
		t.Fatalf("expected CONTEXT7_API_KEY %q, got %#v", want, got)
	}
}
