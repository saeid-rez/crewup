package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"charm.land/huh/v2"
	"github.com/saeid-rez/crewup/internal/workflowstate"
	"github.com/saeid-rez/crewup/pkg/ui"
	"github.com/spf13/cobra"
)

var workflowArchiveName string

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage WORKFLOW_STATE.md",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var workflowStashCmd = &cobra.Command{
	Use:   "stash",
	Short: "Archive WORKFLOW_STATE.md and create a clean one",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := runWorkflowStash()
		if errors.Is(err, ui.ErrUserCancelled) {
			fmt.Println("Workflow state stash cancelled.")
			return nil
		}
		return err
	},
}

func runWorkflowStash() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return runWorkflowStashInDir(dir)
}

func runWorkflowStashInDir(dir string) error {
	statePath := workflowstate.FilePath(dir)
	content, err := os.ReadFile(statePath)
	stateExists := err == nil
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", workflowstate.FileName, err)
	}

	archiveName := workflowArchiveName
	if stateExists && archiveName == "" {
		suggested, err := workflowstate.NextAvailableArchiveName(dir, workflowstate.SuggestArchiveName(content, time.Now()))
		if err != nil {
			return err
		}

		if ui.IsTTY() {
			archiveName, err = promptWorkflowArchiveName(dir, suggested)
			if err != nil {
				return err
			}
		} else {
			archiveName = suggested
		}
	}

	result, err := workflowstate.ArchiveAndReset(dir, archiveName)
	if err != nil {
		return err
	}

	if result.ArchivedPath != "" {
		stashRef, stashed, err := stashWorkflowArchive(dir, result.ArchivedPath, archiveName)
		if err != nil {
			fmt.Printf("Git is not available or stash failed. Kept %s on disk at %s.\n", workflowstate.FileName, filepath.Base(result.ArchivedPath))
			fmt.Printf("Stash failed: %v\n", err)
		} else if stashed {
			fmt.Printf("Stashed previous %s as %s using %s.\n", workflowstate.FileName, stashRef, filepath.Base(result.ArchivedPath))
		}
	} else {
		fmt.Printf("No %s found.\n", workflowstate.FileName)
	}
	fmt.Printf("Created a clean %s and left it in your working tree.\n", filepath.Base(result.CreatedPath))
	return nil
}

func stashWorkflowArchive(dir, archivePath, archiveName string) (string, bool, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return "", false, fmt.Errorf("git executable not found in PATH")
	}

	relativeArchivePath, err := filepath.Rel(dir, archivePath)
	if err != nil {
		return "", false, fmt.Errorf("compute archive path: %w", err)
	}

	message := fmt.Sprintf("workflow-state stash: %s", archiveName)
	cmd := exec.Command("git", "stash", "push", "-u", "-m", message, "--", relativeArchivePath)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", false, fmt.Errorf("git stash push failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	stashRef := strings.TrimSpace(gitOutputOrEmpty(dir, "stash", "list", "-1", "--format=%gd"))
	if stashRef == "" {
		stashRef = message
	}

	return stashRef, true, nil
}

func gitOutputOrEmpty(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return string(output)
}

func promptWorkflowArchiveName(dir, suggested string) (string, error) {
	archiveName := suggested
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Archive name for the current workflow state").
				Value(&archiveName).
				Validate(func(v string) error {
					archivePath, err := workflowstate.ArchivePath(dir, v)
					if err != nil {
						return err
					}
					if _, err := os.Stat(archivePath); err == nil {
						return fmt.Errorf("%s already exists", filepath.Base(archivePath))
					} else if !os.IsNotExist(err) {
						return err
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ui.ErrUserCancelled
		}
		return "", err
	}

	return workflowstate.NormalizeArchiveName(archiveName), nil
}

func init() {
	workflowCmd.AddCommand(workflowStashCmd)
	workflowStashCmd.Flags().StringVar(&workflowArchiveName, "name", "", "Archive name for WORKFLOW_STATE_<name>.md")
}
