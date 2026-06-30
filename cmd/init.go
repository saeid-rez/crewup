package cmd

import (
	"errors"
	"fmt"

	"github.com/saeid-rez/crewup/internal/config"
	"github.com/saeid-rez/crewup/internal/mcp"
	"github.com/saeid-rez/crewup/internal/models"
	"github.com/saeid-rez/crewup/internal/tools"
	"github.com/saeid-rez/crewup/pkg/ui"
	"github.com/spf13/cobra"
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
	fmt.Println("\n🔍 Detecting installed AI tools on your machine...")
	fmt.Println()
	detected := tools.DetectInstalledTools()

	if len(detected) == 0 {
		fmt.Println("⚠️  No supported AI tools detected.")
		fmt.Println("   Supported: GitHub Copilot, OpenCode.")
		fmt.Println("   Install one and re-run: crewup init")
		fmt.Println()
		return nil
	}

	// Step 2: Let user pick which tools to configure
	selectedTools, err := ui.SelectTools(detected)
	if err != nil {
		if errors.Is(err, ui.ErrUserCancelled) {
			fmt.Println("Setup cancelled.")
			return nil
		}
		return err
	}

	// Step 3: Let user pick which agent roles to set up
	selectedRoles, err := ui.SelectAgentRoles()
	if err != nil {
		if errors.Is(err, ui.ErrUserCancelled) {
			fmt.Println("Setup cancelled.")
			return nil
		}
		return err
	}

	// Step 3b: Configure per-agent models
	configuredRoles, err := ui.ConfigureAgents(selectedRoles, models.All())
	if err != nil {
		if errors.Is(err, ui.ErrUserCancelled) {
			fmt.Println("Setup cancelled.")
			return nil
		}
		return err
	}

	// Convert []tools.AITool → []config.ToolInfo
	toolInfos := make([]config.ToolInfo, len(selectedTools))
	for i, t := range selectedTools {
		toolInfos[i] = config.ToolInfo{
			ID:         t.ID,
			Name:       t.Name,
			ConfigPath: t.ConfigPath,
		}
	}

	// Step 4: Write agent config and save crewup config (no MCP here)
	cfg := config.NewCrewConfig(toolInfos, configuredRoles, nil)

	writeAgentConfig := func(toolID, configPath string, roles []config.AgentRole) error {
		writer, ok := tools.Writers[toolID]
		if !ok {
			return fmt.Errorf("no writer registered for tool %q", toolID)
		}
		return writer.WriteAgentConfig(configPath, roles)
	}

	if err := cfg.Apply(writeAgentConfig); err != nil {
		// Non-fatal: partial failures are printed inline
		fmt.Printf("⚠️  Some configurations had errors: %v\n", err)
	}

	// Step 5: MCP preset selection and install
	selectedPresets, err := ui.SelectMCPPresets()
	if err != nil {
		if errors.Is(err, ui.ErrUserCancelled) {
			fmt.Println("Setup cancelled.")
			return nil
		}
		return err
	}

	var installedMCPIDs []string
	for _, preset := range selectedPresets {
		values, err := ui.PromptInputFields(preset)
		if err != nil {
			if errors.Is(err, ui.ErrUserCancelled) {
				fmt.Println("Setup cancelled.")
				return nil
			}
			fmt.Printf("  ⚠️  %s: skipping input prompts (%v)\n", preset.Name, err)
			continue
		}

		targetTools, err := ui.SelectTargetTools(toolInfos, preset.Name)
		if err != nil {
			if errors.Is(err, ui.ErrUserCancelled) {
				fmt.Println("Setup cancelled.")
				return nil
			}
			fmt.Printf("  ⚠️  %s: skipping tool selection (%v)\n", preset.Name, err)
			continue
		}

		successCount := 0
		for _, tool := range targetTools {
			scope := mcp.ScopeGlobal
			if tool.ID == "opencode" {
				scope = mcp.ScopeProject
			}
			if err := mcp.Install(preset.ID, tool.ID, scope, mcp.InstallOptions{Values: values}); err != nil {
				fmt.Printf("  ⚠️  %s: %v\n", tool.Name, err)
			} else {
				successCount++
			}
		}

		if successCount > 0 {
			installedMCPIDs = append(installedMCPIDs, preset.ID)
		}
	}

	// Update config with installed MCP IDs
	cfg.MCPServers = installedMCPIDs
	if err := cfg.Save(); err != nil {
		fmt.Printf("  ⚠️  Could not update config.json with MCP servers: %v\n", err)
	}

	ui.PrintSuccess(selectedTools, selectedRoles, installedMCPIDs)
	return nil
}
