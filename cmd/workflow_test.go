package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWorkflowStashInDirWithExplicitName(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "WORKFLOW_STATE.md"), []byte("# Workflow State\n\n## Request\n- Add command.\n"), 0644); err != nil {
		t.Fatalf("write state file: %v", err)
	}

	original := workflowArchiveName
	workflowArchiveName = "ui-command"
	defer func() { workflowArchiveName = original }()

	if err := runWorkflowStashInDir(dir); err != nil {
		t.Fatalf("run workflow stash: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "WORKFLOW_STATE_ui-command.md")); err != nil && !os.IsNotExist(err) {
		t.Fatalf("stat archived file: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "WORKFLOW_STATE.md"))
	if err != nil {
		t.Fatalf("read clean workflow state: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("expected clean workflow state content")
	}
	if strings.Contains(string(content), "Add command") {
		t.Fatal("expected clean workflow state content, found original workflow state text")
	}

	stashList := gitOutput(t, dir, "stash", "list")
	if !strings.Contains(stashList, "workflow-state stash: ui-command") {
		t.Fatalf("expected workflow stash entry, got: %s", stashList)
	}
}

func TestRunWorkflowStashInDirCreatesCleanFileWhenMissing(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	original := workflowArchiveName
	workflowArchiveName = ""
	defer func() { workflowArchiveName = original }()

	if err := runWorkflowStashInDir(dir); err != nil {
		t.Fatalf("run workflow stash: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "WORKFLOW_STATE.md")); err != nil {
		t.Fatalf("expected clean workflow state file: %v", err)
	}

	stashList := gitOutput(t, dir, "stash", "list")
	if stashList != "" {
		t.Fatalf("expected no stash entries, got: %s", stashList)
	}
}

func TestStashWorkflowArchiveKeepsArchiveWhenGitUnavailable(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "WORKFLOW_STATE_ui-command.md")
	if err := os.WriteFile(archivePath, []byte("old"), 0644); err != nil {
		t.Fatalf("write archive file: %v", err)
	}

	originalPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", dir); err != nil {
		t.Fatalf("set PATH: %v", err)
	}
	defer func() {
		_ = os.Setenv("PATH", originalPath)
	}()

	_, stashed, err := stashWorkflowArchive(dir, archivePath, "ui-command")
	if err == nil {
		t.Fatal("expected stashWorkflowArchive to report git unavailable")
	}
	if stashed {
		t.Fatal("expected stashed=false when git is unavailable")
	}
	if _, statErr := os.Stat(archivePath); statErr != nil {
		t.Fatalf("expected archive file to remain on disk, got: %v", statErr)
	}
	if !strings.Contains(err.Error(), "git executable not found") {
		t.Fatalf("expected git unavailable error, got: %v", err)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitRun(t, dir, "init")
	if err := os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte{}, 0644); err != nil {
		t.Fatalf("write .gitkeep: %v", err)
	}
	gitRun(t, dir, "add", ".gitkeep")
	gitRun(t, dir, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "init")
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
	return strings.TrimSpace(string(output))
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	_ = gitOutput(t, dir, args...)
}
