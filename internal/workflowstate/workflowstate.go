package workflowstate

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	FileName      = "WORKFLOW_STATE.md"
	archivePrefix = "WORKFLOW_STATE_"
	archiveSuffix = ".md"
)

type ResetResult struct {
	ArchivedPath string
	CreatedPath  string
}

var archiveNameSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func FilePath(dir string) string {
	return filepath.Join(dir, FileName)
}

func ArchivePath(dir, name string) (string, error) {
	filename, err := ArchiveFilename(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}

func ArchiveFilename(name string) (string, error) {
	normalized := NormalizeArchiveName(name)
	if normalized == "" {
		return "", errors.New("archive name cannot be empty")
	}
	return archivePrefix + normalized + archiveSuffix, nil
}

func NormalizeArchiveName(name string) string {
	normalized := strings.TrimSpace(name)
	normalized = strings.TrimSuffix(normalized, archiveSuffix)
	normalized = strings.TrimPrefix(normalized, archivePrefix)
	normalized = strings.TrimPrefix(normalized, strings.TrimSuffix(FileName, archiveSuffix)+"_")
	normalized = strings.TrimPrefix(normalized, strings.TrimSuffix(FileName, archiveSuffix))
	normalized = strings.ToLower(normalized)
	normalized = archiveNameSanitizer.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	if len(normalized) > 64 {
		normalized = strings.Trim(normalized[:64], "-")
	}
	return normalized
}

func SuggestArchiveName(content []byte, now time.Time) string {
	body := string(content)
	for _, section := range []string{"## Commit Message Draft", "## Clarified Scope", "## Request"} {
		if name := normalizeSuggestedLine(firstMeaningfulLine(body, section)); name != "" {
			return name
		}
	}
	return fmt.Sprintf("workflow-%s", now.Format("2006-01-02"))
}

func NextAvailableArchiveName(dir, preferred string) (string, error) {
	base := NormalizeArchiveName(preferred)
	if base == "" {
		return "", errors.New("archive name cannot be empty")
	}

	for i := 1; ; i++ {
		candidate := base
		if i > 1 {
			candidate = fmt.Sprintf("%s-%d", base, i)
		}

		archivePath, err := ArchivePath(dir, candidate)
		if err != nil {
			return "", err
		}

		_, err = os.Stat(archivePath)
		if os.IsNotExist(err) {
			return candidate, nil
		}
		if err != nil {
			return "", fmt.Errorf("stat archive path: %w", err)
		}
	}
}

func ArchiveAndReset(dir, archiveName string) (ResetResult, error) {
	statePath := FilePath(dir)
	info, err := os.Stat(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(statePath, CleanTemplate(), 0644); err != nil {
				return ResetResult{}, fmt.Errorf("write %s: %w", FileName, err)
			}
			return ResetResult{CreatedPath: statePath}, nil
		}
		return ResetResult{}, fmt.Errorf("stat %s: %w", FileName, err)
	}

	archivePath, err := ArchivePath(dir, archiveName)
	if err != nil {
		return ResetResult{}, err
	}

	if _, err := os.Stat(archivePath); err == nil {
		return ResetResult{}, fmt.Errorf("archive already exists: %s", filepath.Base(archivePath))
	} else if !os.IsNotExist(err) {
		return ResetResult{}, fmt.Errorf("stat archive path: %w", err)
	}

	if err := copyFile(statePath, archivePath, info.Mode().Perm()); err != nil {
		return ResetResult{}, fmt.Errorf("archive %s: %w", FileName, err)
	}

	perm := info.Mode().Perm()
	if err := os.WriteFile(statePath, CleanTemplate(), perm); err != nil {
		if removeErr := os.Remove(archivePath); removeErr != nil {
			return ResetResult{}, fmt.Errorf("write clean %s: %w (cleanup failed: %v)", FileName, err, removeErr)
		}
		return ResetResult{}, fmt.Errorf("write clean %s: %w", FileName, err)
	}

	return ResetResult{ArchivedPath: archivePath, CreatedPath: statePath}, nil
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return err
	}

	if err := out.Close(); err != nil {
		_ = os.Remove(dst)
		return err
	}

	return nil

}

func CleanTemplate() []byte {
	return []byte(`# Workflow State

## Request
- Pending.

## Clarified Scope
- Pending.

## Acceptance Criteria
- Pending.

## Open Questions
- None yet.

## Constraints
- None yet.

## Current Status
- Ready for planning.

## Plan
- Pending.

## Debate Notes
- Pending.

## Files To Change
- Pending.

## Implementation Notes
- Pending.

## Review Findings
- Pending formal review.

## Security Findings
- Pending security review.

## Test Results
- Not run yet.

## Lint Results
- Not run yet.

## Commit Message Draft
- Pending.
`)
}

func firstMeaningfulLine(body, section string) string {
	lines := strings.Split(body, "\n")
	inSection := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == section {
			inSection = true
			continue
		}
		if inSection && strings.HasPrefix(trimmed, "## ") {
			return ""
		}
		if !inSection || trimmed == "" {
			continue
		}

		cleaned := cleanSummaryLine(trimmed)
		if cleaned == "" || isPlaceholderLine(cleaned) {
			continue
		}
		return cleaned
	}
	return ""
}

func cleanSummaryLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "- ")
	line = strings.TrimPrefix(line, "* ")
	line = strings.Trim(line, "`")
	return strings.TrimSpace(line)
}

func isPlaceholderLine(line string) bool {
	normalized := strings.ToLower(strings.TrimSpace(line))
	placeholders := []string{
		"pending",
		"pending.",
		"none yet",
		"none yet.",
		"ready for planning",
		"ready for planning.",
		"not run yet",
		"not run yet.",
		"pending formal review",
		"pending formal review.",
		"pending security review",
		"pending security review.",
	}
	for _, placeholder := range placeholders {
		if normalized == placeholder {
			return true
		}
	}
	return false
}

func normalizeSuggestedLine(line string) string {
	return NormalizeArchiveName(line)
}
