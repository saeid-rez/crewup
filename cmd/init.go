package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/saeid-rez/crewup/internal/config"
	"github.com/saeid-rez/crewup/internal/tools"
	"github.com/saeid-rez/crewup/pkg/ui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive setup wizard — detect AI tools and configure your crew",
	Long: `Runs an interactive wizard that:
  1. Detects installed AI tools on your machine
  2. Asks which ones you want to configure
  3. Sets up your agent crew (planner, implementor, reviewer, etc.)
  4. Optionally installs popular MCP servers`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

func runInit() error {
	ui.PrintBanner()

	// Step 1: Detect installed AI tools
	fmt.Println("\n🔍 Detecting installed AI tools on your machine...\n")
	detected := tools.DetectInstalledTools()

	if len(detected) == 0 {
		fmt.Println("⚠️  No supported AI tools detected.")
		fmt.Println("   Supported: Claude CLI, GitHub Copilot, Ollama, MinMax, and more.")
		fmt.Println("   Install one and re-run: crewup init\n")
		return nil
	}

	// Step 2: Let user pick which tools to configure
	selectedTools, err := ui.SelectTools(detected)
	if err != nil {
		return err
	}

	// Step 3: Let user pick which agent roles to set up
	selectedRoles, err := ui.SelectAgentRoles()
	if err != nil {
		return err
	}

	// Step 4: Ask about MCP servers
	selectedMCPs, err := ui.SelectMCPServers()
	if err != nil {
		return err
	}

	// Step 5: Write config and apply
	cfg := config.NewCrewConfig(selectedTools, selectedRoles, selectedMCPs)
	if err := cfg.Apply(); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	ui.PrintSuccess(selectedTools, selectedRoles, selectedMCPs)
	return nil
}
