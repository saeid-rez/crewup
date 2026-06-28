package cmd

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/saeid-rez/crewup/internal/agentdefs"
	"github.com/saeid-rez/crewup/internal/config"
	"github.com/saeid-rez/crewup/internal/mcp"
	"github.com/saeid-rez/crewup/internal/models"
	"github.com/saeid-rez/crewup/internal/tools"
	"github.com/saeid-rez/crewup/pkg/ui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an MCP server or agent role to your existing setup",
}

var projectScope bool

var addMCPCmd = &cobra.Command{
	Use:   "mcp [server-name]",
	Short: "Add a popular MCP server to your configured AI tools",
	Example: `  crewup add mcp context7
  crewup add mcp context7 --project
  crewup add mcp`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve presets before loading config so we can fail fast on bad input.
		var presets []mcp.MCPPreset

		if len(args) == 1 {
			// Arg provided: validate immediately
			serverName := args[0]
			preset, ok := mcp.FindByID(serverName)
			if !ok {
				ids := make([]string, len(mcp.Presets))
				for i, p := range mcp.Presets {
					ids[i] = p.ID
				}
				return fmt.Errorf("unknown MCP server %q. Valid presets: %s", serverName, strings.Join(ids, ", "))
			}
			presets = []mcp.MCPPreset{preset}
		} else {
			// No arg: non-TTY must return a clear error; TTY shows interactive picker
			if !ui.IsTTY() {
				return fmt.Errorf("crewup add mcp requires a server name in non-interactive mode. Run: crewup add mcp <name>")
			}
			selected, err := ui.SelectMCPPresets()
			if err != nil {
				if errors.Is(err, ui.ErrUserCancelled) {
					fmt.Println("Setup cancelled.")
					return nil
				}
				return err
			}
			if len(selected) == 0 {
				// TTY user selected nothing — exit cleanly
				return nil
			}
			presets = selected
		}

		// Load config to find which tools are configured
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("could not load config (run `crewup init` first): %w", err)
		}

		scope := mcp.ScopeGlobal
		if projectScope {
			scope = mcp.ScopeProject
		}
		_ = scope // per-tool scope is determined below

		for _, preset := range presets {
			fmt.Printf("🔌 Adding MCP server: %s\n", preset.Name)

			values, err := ui.PromptInputFields(preset)
			if err != nil {
				fmt.Printf("  ⚠️  %s: skipping (%v)\n", preset.Name, err)
				continue
			}

			targetTools, err := ui.SelectTargetTools(cfg.Tools, preset.Name)
			if err != nil {
				if errors.Is(err, ui.ErrUserCancelled) {
					fmt.Println("Setup cancelled.")
					return nil
				}
				fmt.Printf("  ⚠️  %s: skipping tool selection (%v)\n", preset.Name, err)
				continue
			}

			var errs []string
			successCount := 0
			for _, tool := range targetTools {
				toolScope := mcp.ScopeGlobal
				if projectScope {
					toolScope = mcp.ScopeProject
				}
				if tool.ID == "opencode" {
					toolScope = mcp.ScopeProject
				}
				if err := mcp.Install(preset.ID, tool.ID, toolScope, mcp.InstallOptions{Values: values}); err != nil {
					errs = append(errs, fmt.Sprintf("  ⚠️  %s: %v", tool.Name, err))
				} else {
					successCount++
				}
			}

			if len(errs) > 0 {
				for _, e := range errs {
					fmt.Println(e)
				}
			}

			// Only record preset ID after at least one successful install
			if successCount > 0 {
				alreadyPresent := false
				for _, s := range cfg.MCPServers {
					if s == preset.ID {
						alreadyPresent = true
						break
					}
				}
				if !alreadyPresent {
					cfg.MCPServers = append(cfg.MCPServers, preset.ID)
					if saveErr := cfg.Save(); saveErr != nil {
						fmt.Printf("  ⚠️  Could not update config.json: %v\n", saveErr)
					}
				}
			}

			if successCount == 0 {
				fmt.Println("⚠️  No MCP servers were installed. Check the errors above.")
			} else if len(errs) > 0 {
				fmt.Printf("⚠️  Partially added %s (some tools failed — see above).\n", preset.Name)
			} else {
				fmt.Printf("✓ Done adding %s!\n", preset.Name)
			}
		}
		return nil
	},
}

var addAgentCmd = &cobra.Command{
	Use:   "agent [role]",
	Short: "Add an agent role to your crew",
	Example: `  crewup add agent planner
  crewup add agent reviewer
  crewup add agent tester`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		roleID := args[0]

		// Step 1: Validate role ID
		def, ok := agentdefs.ByID(roleID)
		if !ok {
			defs, _ := agentdefs.All()
			ids := make([]string, len(defs))
			for i, d := range defs {
				ids[i] = d.ID
			}
			return fmt.Errorf("unknown agent role %q. Valid roles: %s", roleID, strings.Join(ids, ", "))
		}

		// Step 2: Require existing config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("could not load config (run `crewup init` first): %w", err)
		}

		// Step 3: Check if role already exists
		existingIdx := -1
		for i, r := range cfg.AgentRoles {
			if r.ID == roleID {
				existingIdx = i
				break
			}
		}

		if existingIdx >= 0 {
			replace := false
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Replace existing config for %s?", def.Name)).
						Value(&replace),
				),
			)
			if err := form.Run(); err != nil {
				return err
			}
			if !replace {
				fmt.Printf("Keeping existing config for %s.\n", def.Name)
				return nil
			}
		}

		// Step 4: Build base role from AgentDef
		role := config.AgentRole{
			ID:          def.ID,
			Name:        def.Name,
			Description: def.Description,
			Prompt:      def.Prompt,
			Mode:        def.Mode,
		}

		// Step 5: Run per-agent wizard
		configured, err := ui.ConfigureAgent(role, models.All())
		if err != nil {
			return err
		}

		// Step 6: Update config
		if existingIdx >= 0 {
			cfg.AgentRoles[existingIdx] = configured
		} else {
			cfg.AgentRoles = append(cfg.AgentRoles, configured)
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		// Step 7: Re-run WriteAgentConfig for all configured tools
		for _, tool := range cfg.Tools {
			writer, ok := tools.Writers[tool.ID]
			if !ok {
				continue
			}
			if err := writer.WriteAgentConfig(tool.ConfigPath, cfg.AgentRoles); err != nil {
				fmt.Printf("  ⚠️  %s: agent config skipped (%v)\n", tool.Name, err)
			}
		}

		fmt.Printf("✅ %s agent configured.\n", def.Name)
		return nil
	},
}

func init() {
	addCmd.AddCommand(addMCPCmd)
	addCmd.AddCommand(addAgentCmd)
	addMCPCmd.Flags().BoolVarP(&projectScope, "project", "p", false, "Write MCP config to project scope (.vscode/mcp.json). OpenCode always writes to opencode.json in the current folder")
}
