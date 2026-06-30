package cmd

import (
	"fmt"
	"os"

	"github.com/saeid-rez/crewup/internal/updater"
	"github.com/spf13/cobra"
)

// Version is injected at build time via ldflags
// e.g. go build -ldflags "-X cmd.Version=1.0.0"
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "crewup",
	Short: "crewup — scaffold AI agent workflows for any AI tool",
	Long: `
╔═══════════════════════════════════════╗
║              crewup                   ║
║   Setup AI agents, crews & MCP       ║
║   servers for any AI tool            ║
╚═══════════════════════════════════════╝

crewup helps you configure complete AI agent workflows
(planner, implementor, reviewer, etc.) for tools like
Claude CLI, Copilot, Opencode, and more.
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Check for updates on every run (non-blocking)
		updater.CheckForUpdate(Version)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	configureRootMenu()
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("crewup version %s\n", Version)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update crewup to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		return updaterCmdRun()
	},
}
