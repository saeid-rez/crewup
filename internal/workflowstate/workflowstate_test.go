package workflowstate

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSuggestArchiveNamePrefersCommitMessageDraft(t *testing.T) {
	content := []byte(`# Workflow State

## Commit Message Draft
- ` + "`ui: add welcome menu and bubble tea navigation`" + `

## Request
- Replace the setup screens.
`)

	got := SuggestArchiveName(content, time.Date(2026, time.July, 12, 0, 0, 0, 0, time.UTC))
	want := "ui-add-welcome-menu-and-bubble-tea-navigation"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestNextAvailableArchiveNameAppendsCounter(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "WORKFLOW_STATE_feature.md"), []byte("old"), 0644); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	got, err := NextAvailableArchiveName(dir, "feature")
	if err != nil {
		t.Fatalf("next available archive name: %v", err)
	}
	if got != "feature-2" {
		t.Fatalf("expected feature-2, got %q", got)
	}
}

func TestArchiveAndResetArchivesExistingFileAndWritesTemplate(t *testing.T) {
	dir := t.TempDir()
	original := []byte("# Workflow State\n\n## Request\n- Ship a feature.\n")
	statePath := FilePath(dir)
	if err := os.WriteFile(statePath, original, 0600); err != nil {
		t.Fatalf("write state file: %v", err)
	}

	result, err := ArchiveAndReset(dir, "ship-feature")
	if err != nil {
		t.Fatalf("archive and reset: %v", err)
	}

	if filepath.Base(result.ArchivedPath) != "WORKFLOW_STATE_ship-feature.md" {
		t.Fatalf("unexpected archive path: %s", result.ArchivedPath)
	}

	archived, err := os.ReadFile(result.ArchivedPath)
	if err != nil {
		t.Fatalf("read archived file: %v", err)
	}
	if string(archived) != string(original) {
		t.Fatalf("archived content mismatch\nwant:\n%s\n\ngot:\n%s", original, archived)
	}

	originalAfter, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read rewritten state file: %v", err)
	}
	if string(originalAfter) == string(original) {
		t.Fatal("expected WORKFLOW_STATE.md to be rewritten to a clean template")
	}

	cleaned, err := os.ReadFile(result.CreatedPath)
	if err != nil {
		t.Fatalf("read cleaned state file: %v", err)
	}
	if string(cleaned) != string(CleanTemplate()) {
		t.Fatalf("clean template mismatch\nwant:\n%s\n\ngot:\n%s", CleanTemplate(), cleaned)
	}

	info, err := os.Stat(result.CreatedPath)
	if err != nil {
		t.Fatalf("stat cleaned state file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("expected permissions 0600, got %#o", info.Mode().Perm())
	}
}

func TestArchiveAndResetCreatesCleanFileWhenMissing(t *testing.T) {
	dir := t.TempDir()

	result, err := ArchiveAndReset(dir, "")
	if err != nil {
		t.Fatalf("archive and reset: %v", err)
	}
	if result.ArchivedPath != "" {
		t.Fatalf("expected no archived path, got %q", result.ArchivedPath)
	}

	cleaned, err := os.ReadFile(result.CreatedPath)
	if err != nil {
		t.Fatalf("read cleaned state file: %v", err)
	}
	if string(cleaned) != string(CleanTemplate()) {
		t.Fatalf("clean template mismatch\nwant:\n%s\n\ngot:\n%s", CleanTemplate(), cleaned)
	}
}
