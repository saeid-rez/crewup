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

	preset, _ := FindByID("context7")

	// VSCode (project) — .vscode/mcp.json with servers as array
	vscodeDir := filepath.Join(tmp, ".vscode")
	if err := os.MkdirAll(vscodeDir, 0755); err != nil {
		t.Fatalf("mkdir .vscode: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vscodeDir, "mcp.json"), []byte(`{"servers": []}`), 0644); err != nil {
		t.Fatalf("write vscode mcp.json: %v", err)
	}
	if err := (&VSCodeMCPMerger{}).Merge(preset, ScopeProject, InstallOptions{}); err == nil {
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

	preset, _ := FindByID("context7")
	if err := (&OpenCodeMCPMerger{}).Merge(preset, ScopeProject, InstallOptions{Values: map[string]string{"CONTEXT7_API_KEY": "ctx7sk-test"}}); err != nil {
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

	preset, _ := FindByID("context7")
	if err := (&OpenCodeMCPMerger{}).Merge(preset, ScopeProject, InstallOptions{}); err == nil {
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

	preset, _ := FindByID("context7")
	if err := (&VSCodeMCPMerger{}).Merge(preset, ScopeProject, InstallOptions{Values: map[string]string{"CONTEXT7_API_KEY": "ctx7sk-test"}}); err != nil {
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

// TestSensitiveFilePermissions verifies that config files written with a sensitive
// value (Sensitive InputField + non-empty user value) get 0600 permissions.
func TestSensitiveFilePermissions(t *testing.T) {
	tmp, err := os.MkdirTemp("", "crewup-mcp-perms-")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	defer os.RemoveAll(tmp)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	preset, _ := FindByID("context7")

	// VSCode project-scoped with a sensitive value supplied.
	if err := (&VSCodeMCPMerger{}).Merge(preset, ScopeProject, InstallOptions{
		Values: map[string]string{"CONTEXT7_API_KEY": "secret"},
	}); err != nil {
		t.Fatalf("vscode merge: %v", err)
	}

	vscodePath := filepath.Join(tmp, ".vscode", "mcp.json")
	info, err := os.Stat(vscodePath)
	if err != nil {
		t.Fatalf("stat vscode mcp.json: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Errorf("expected 0600 permissions for sensitive vscode config, got %04o", got)
	}

	// OpenCode project-scoped with a sensitive value supplied.
	if err := (&OpenCodeMCPMerger{}).Merge(preset, ScopeProject, InstallOptions{
		Values: map[string]string{"CONTEXT7_API_KEY": "secret"},
	}); err != nil {
		t.Fatalf("opencode merge: %v", err)
	}

	opencodePath := filepath.Join(tmp, "opencode.json")
	info2, err := os.Stat(opencodePath)
	if err != nil {
		t.Fatalf("stat opencode.json: %v", err)
	}
	if got := info2.Mode().Perm(); got != 0600 {
		t.Errorf("expected 0600 permissions for sensitive opencode config, got %04o", got)
	}
}

// TestNonSensitiveFilePermissions verifies that config files written without
// sensitive values keep the default 0644 permissions.
func TestNonSensitiveFilePermissions(t *testing.T) {
	tmp, err := os.MkdirTemp("", "crewup-mcp-perms-nonsensitive-")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	defer os.RemoveAll(tmp)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	preset, _ := FindByID("context7")

	// No value supplied for the sensitive field → should use default 0644.
	if err := (&VSCodeMCPMerger{}).Merge(preset, ScopeProject, InstallOptions{
		Values: map[string]string{},
	}); err != nil {
		t.Fatalf("vscode merge: %v", err)
	}

	vscodePath := filepath.Join(tmp, ".vscode", "mcp.json")
	info, err := os.Stat(vscodePath)
	if err != nil {
		t.Fatalf("stat vscode mcp.json: %v", err)
	}
	if got := info.Mode().Perm(); got != 0644 {
		t.Errorf("expected 0644 permissions for non-sensitive vscode config, got %04o", got)
	}
}

// TestSymlinkRejectedVSCode verifies that a symlink at the project-scoped
// .vscode/mcp.json path is rejected before any write.
func TestSymlinkRejectedVSCode(t *testing.T) {
	tmp, err := os.MkdirTemp("", "crewup-mcp-symlink-vscode-")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	defer os.RemoveAll(tmp)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Create .vscode dir and place a symlink where mcp.json would go.
	vscodeDir := filepath.Join(tmp, ".vscode")
	if err := os.MkdirAll(vscodeDir, 0755); err != nil {
		t.Fatalf("mkdir .vscode: %v", err)
	}
	target := filepath.Join(tmp, "real_file.json")
	if err := os.WriteFile(target, []byte("{}"), 0644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	symlinkPath := filepath.Join(vscodeDir, "mcp.json")
	if err := os.Symlink(target, symlinkPath); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	preset, _ := FindByID("context7")
	err = (&VSCodeMCPMerger{}).Merge(preset, ScopeProject, InstallOptions{})
	if err == nil {
		t.Fatal("expected error for symlink target, got nil")
	}
	if !containsStr(err.Error(), "symlink") {
		t.Fatalf("expected error to mention symlink, got: %v", err)
	}
}

// TestSymlinkRejectedOpenCode verifies that a symlink at opencode.json is rejected.
func TestSymlinkRejectedOpenCode(t *testing.T) {
	tmp, err := os.MkdirTemp("", "crewup-mcp-symlink-opencode-")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	defer os.RemoveAll(tmp)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	target := filepath.Join(tmp, "real_opencode.json")
	if err := os.WriteFile(target, []byte("{}"), 0644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	symlinkPath := filepath.Join(tmp, "opencode.json")
	if err := os.Symlink(target, symlinkPath); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	preset, _ := FindByID("context7")
	err = (&OpenCodeMCPMerger{}).Merge(preset, ScopeProject, InstallOptions{})
	if err == nil {
		t.Fatal("expected error for symlink target, got nil")
	}
	if !containsStr(err.Error(), "symlink") {
		t.Fatalf("expected error to mention symlink, got: %v", err)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
