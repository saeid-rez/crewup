package cmd

import (
	"errors"
	"fmt"

	"github.com/saeid-rez/crewup/internal/updater"
	"github.com/saeid-rez/crewup/pkg/ui"
	"github.com/spf13/cobra"
)

func runMainMenu() error {
	for {
		ui.PrintBanner()

		choice, err := ui.RunMenu(
			"Welcome To crewup",
			"Choose a command to get started. You can always come back here with cancel/back during interactive flows.",
			[]ui.MenuOption{
				{ID: "init", Title: "Init", Description: "Detect AI tools and set up your crew from scratch."},
				{ID: "add", Title: "Add", Description: "Add an MCP server or an agent role to an existing setup."},
				{ID: "list", Title: "List", Description: "Show your current crewup configuration."},
				{ID: "version", Title: "Version", Description: "Show the installed crewup version."},
				{ID: "update", Title: "Update", Description: "Update crewup to the latest release."},
				{ID: "help", Title: "Help", Description: "Show the CLI help and available commands."},
				{ID: "quit", Title: "Quit", Description: "Exit crewup."},
			},
		)
		if err != nil {
			if errors.Is(err, ui.ErrUserCancelled) {
				return nil
			}
			return err
		}

		var runErr error
		switch choice {
		case "init":
			runErr = runInit()
		case "add":
			runErr = runAddMenu()
		case "list":
			runErr = runList()
		case "version":
			fmt.Printf("crewup version %s\n", Version)
		case "update":
			runErr = updaterCmdRun()
		case "help":
			runErr = rootCmd.Help()
		case "quit":
			return nil
		default:
			runErr = fmt.Errorf("unknown main menu option %q", choice)
		}

		if runErr != nil {
			if errors.Is(runErr, ui.ErrUserCancelled) {
				fmt.Println("Returning to main menu.")
				fmt.Println()
				continue
			}
			return runErr
		}

		if choice == "help" || choice == "version" || choice == "list" {
			fmt.Println()
		}

		if !ui.IsTTY() {
			return nil
		}
	}
}

func updaterCmdRun() error {
	return updater.Update(Version)
}

func configureRootMenu() {
	rootCmd.Args = cobra.NoArgs
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if !ui.IsTTY() {
			fmt.Println("Interactive home screen requires a terminal (TTY).")
			fmt.Println("Run `crewup` in a normal terminal, or use a direct command like `crewup init`.")
			fmt.Println()
			return cmd.Help()
		}
		return runMainMenu()
	}
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
}
