package ui

import (
	"os"
	"testing"

	"github.com/saeid-rez/crewup/internal/mcp"
)

// Test isTTY fallback when stdin/stdout are not terminals.
func TestIsTTYNonTerminal(t *testing.T) {
	// Temporarily replace os.Stdin and os.Stdout with files.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	origIn := os.Stdin
	origOut := os.Stdout
	defer func() { os.Stdin = origIn; os.Stdout = origOut }()

	os.Stdin = r
	os.Stdout = w

	if isTTY() {
		t.Fatalf("expected isTTY to be false when using pipes")
	}
}

// TestPromptInputFieldsNonTTYNoRequired: non-TTY + no required fields → empty map, nil error.
func TestPromptInputFieldsNonTTYNoRequired(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	origIn := os.Stdin
	origOut := os.Stdout
	defer func() { os.Stdin = origIn; os.Stdout = origOut }()
	os.Stdin = r
	os.Stdout = w

	preset := mcp.MCPPreset{
		ID:   "test-preset",
		Name: "Test Preset",
		Inputs: []mcp.InputField{
			{Key: "OPTIONAL_KEY", Label: "Optional Key", Required: false, Sensitive: false},
		},
	}

	result, err := PromptInputFields(preset)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map, got: %v", result)
	}
}

// TestPromptInputFieldsNonTTYRequired: non-TTY + required field → non-nil error.
func TestPromptInputFieldsNonTTYRequired(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	origIn := os.Stdin
	origOut := os.Stdout
	defer func() { os.Stdin = origIn; os.Stdout = origOut }()
	os.Stdin = r
	os.Stdout = w

	preset := mcp.MCPPreset{
		ID:   "test-preset",
		Name: "Test Preset",
		Inputs: []mcp.InputField{
			{Key: "REQUIRED_KEY", Label: "Required Key", Required: true, Sensitive: false},
		},
	}

	_, err = PromptInputFields(preset)
	if err == nil {
		t.Fatal("expected non-nil error for required field in non-TTY mode")
	}
}
