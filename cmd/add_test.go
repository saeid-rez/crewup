package cmd

import (
	"os"
	"strings"
	"testing"
)

// simulateNonTTY replaces stdin/stdout with pipes so isTTY() returns false.
// Returns a cleanup function that restores the originals.
func simulateNonTTY(t *testing.T) func() {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	origIn, origOut := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = r, w
	return func() {
		os.Stdin = origIn
		os.Stdout = origOut
		r.Close()
		w.Close()
	}
}

// TestAddMCPNoArgNonTTY verifies that `crewup add mcp` with no arg in non-TTY
// mode returns the required error message.
func TestAddMCPNoArgNonTTY(t *testing.T) {
	restore := simulateNonTTY(t)
	defer restore()

	err := addMCPCmd.RunE(addMCPCmd, []string{})
	if err == nil {
		t.Fatal("expected error for no-arg non-TTY, got nil")
	}
	want := "crewup add mcp requires a server name in non-interactive mode"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error to contain %q, got: %v", want, err)
	}
}

// TestAddMCPUnknownPreset verifies that an unknown preset ID returns a helpful
// error that does NOT reference the non-existent `crewup list mcp` command.
func TestAddMCPUnknownPreset(t *testing.T) {
	restore := simulateNonTTY(t)
	defer restore()

	err := addMCPCmd.RunE(addMCPCmd, []string{"nonexistent-preset-xyz"})
	if err == nil {
		t.Fatal("expected error for unknown preset, got nil")
	}
	if strings.Contains(err.Error(), "crewup list mcp") {
		t.Fatalf("error message must not reference non-existent `crewup list mcp`: %v", err)
	}
	if !strings.Contains(err.Error(), "nonexistent-preset-xyz") {
		t.Fatalf("expected error to mention the unknown ID, got: %v", err)
	}
}
