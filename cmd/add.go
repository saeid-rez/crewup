package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
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
  crewup add mcp filesystem
  crewup add mcp github`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverName := args[0]
		fmt.Printf("🔌 Adding MCP server: %s\n", serverName)

		// Load config to find which tools are configured
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("could not load config (run `crewup init` first): %w", err)
		}

		scope := mcp.ScopeGlobal
		if projectScope {
			scope = mcp.ScopeProject
		}

		var errs []string
		installed := false
		for _, tool := range cfg.Tools {
			if err := mcp.Install(serverName, tool.ID, scope); err != nil {
				errs = append(errs, fmt.Sprintf("  ⚠️  %s: %v", tool.Name, err))
			} else {
				installed = true
			}
		}

		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Println(e)
			}
		}

		// Update config.json to record the newly added MCP server.
		if installed {
			alreadyPresent := false
			for _, s := range cfg.MCPServers {
				if s == serverName {
					alreadyPresent = true
					break
				}
			}
			if !alreadyPresent {
				cfg.MCPServers = append(cfg.MCPServers, serverName)
				if saveErr := cfg.Save(); saveErr != nil {
					fmt.Printf("  ⚠️  Could not update config.json: %v\n", saveErr)
				}
			}
		}

		if !installed {
			fmt.Println("⚠️  No MCP servers were installed. Check the errors above.")
		} else if len(errs) > 0 {
			fmt.Printf("⚠️  Partially added %s (some tools failed — see above).\n", serverName)
		} else {
			fmt.Printf("✓ Done adding %s!\n", serverName)
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
	addMCPCmd.Flags().BoolVarP(&projectScope, "project", "p", false, "Write MCP config to project scope (opencode.json / .vscode/mcp.json)")
}
