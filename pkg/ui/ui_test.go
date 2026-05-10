package ui

import (
	"os"
	"testing"
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
