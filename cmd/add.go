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
	RunE: func(cmd *cobra.Command, args []string) error {
		err := runAddMenu()
		if errors.Is(err, ui.ErrUserCancelled) {
			fmt.Println("Returning to main menu.")
			return nil
		}
		return err
	},
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
		err := runAddMCP(args)
		if errors.Is(err, ui.ErrUserCancelled) {
			fmt.Println("Setup cancelled.")
			return nil
		}
		return err
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
		err := runAddAgent(args[0])
		if errors.Is(err, ui.ErrUserCancelled) {
			fmt.Println("Setup cancelled.")
			return nil
		}
		return err
	},
}

func runAddMenu() error {
	if !ui.IsTTY() {
		fmt.Println("Add commands:")
		fmt.Println("  crewup add mcp <name>     Add a popular MCP server")
		fmt.Println("  crewup add agent <role>   Add or customize an agent role")
		return nil
	}

	choice, err := ui.RunMenu(
		"Add To Your Crew",
		"Choose what you want to add to an existing crewup setup.",
		[]ui.MenuOption{
			{ID: "mcp", Title: "Add MCP server", Description: "Install a popular MCP server for your configured AI tools."},
			{ID: "agent", Title: "Add agent role", Description: "Add or customize one agent role in your crew."},
			{ID: "back", Title: "Back to main menu", Description: "Return without making changes."},
		},
	)
	if err != nil {
		return err
	}

	switch choice {
	case "mcp":
		return runAddMCP(nil)
	case "agent":
		roleID, err := selectAgentRoleForAdd()
		if err != nil {
			return err
		}
		return runAddAgent(roleID)
	case "back":
		return ui.ErrUserCancelled
	default:
		return fmt.Errorf("unknown add menu option %q", choice)
	}
}

func runAddMCP(args []string) error {
	// Resolve presets before loading config so we can fail fast on bad input.
	var presets []mcp.MCPPreset

	if len(args) == 1 {
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
		if !ui.IsTTY() {
			return fmt.Errorf("crewup add mcp requires a server name in non-interactive mode. Run: crewup add mcp <name>")
		}
		selected, err := ui.SelectMCPPresets()
		if err != nil {
			return err
		}
		if len(selected) == 0 {
			return nil
		}
		presets = selected
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("could not load config (run `crewup init` first): %w", err)
	}

	for _, preset := range presets {
		fmt.Printf("🔌 Adding MCP server: %s\n", preset.Name)

		values, err := ui.PromptInputFields(preset)
		if err != nil {
			if errors.Is(err, ui.ErrUserCancelled) {
				return err
			}
			fmt.Printf("  ⚠️  %s: skipping (%v)\n", preset.Name, err)
			continue
		}

		targetTools, err := ui.SelectTargetTools(cfg.Tools, preset.Name)
		if err != nil {
			return err
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
}

func runAddAgent(roleID string) error {
	def, ok := agentdefs.ByID(roleID)
	if !ok {
		defs, _ := agentdefs.All()
		ids := make([]string, len(defs))
		for i, d := range defs {
			ids[i] = d.ID
		}
		return fmt.Errorf("unknown agent role %q. Valid roles: %s", roleID, strings.Join(ids, ", "))
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("could not load config (run `crewup init` first): %w", err)
	}

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
			if errors.Is(err, huh.ErrUserAborted) {
				return ui.ErrUserCancelled
			}
			return err
		}
		if !replace {
			fmt.Printf("Keeping existing config for %s.\n", def.Name)
			return nil
		}
	}

	role := config.AgentRole{
		ID:          def.ID,
		Name:        def.Name,
		Description: def.Description,
		Prompt:      def.Prompt,
		Mode:        def.Mode,
	}

	configured, err := ui.ConfigureAgent(role, models.All())
	if err != nil {
		return err
	}

	if existingIdx >= 0 {
		cfg.AgentRoles[existingIdx] = configured
	} else {
		cfg.AgentRoles = append(cfg.AgentRoles, configured)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

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
}

func selectAgentRoleForAdd() (string, error) {
	defs, err := agentdefs.All()
	if err != nil {
		return "", fmt.Errorf("loading agent definitions: %w", err)
	}

	options := make([]ui.MenuOption, 0, len(defs)+1)
	for _, def := range defs {
		options = append(options, ui.MenuOption{ID: def.ID, Title: def.Name, Description: def.Description})
	}
	options = append(options, ui.MenuOption{ID: "back", Title: "Back to main menu", Description: "Return without changing agent roles."})

	choice, err := ui.RunMenu("Choose Agent Role", "Select the agent role you want to add or customize.", options)
	if err != nil {
		return "", err
	}
	if choice == "back" {
		return "", ui.ErrUserCancelled
	}
	return choice, nil
}

func init() {
	addCmd.AddCommand(addMCPCmd)
	addCmd.AddCommand(addAgentCmd)
	addMCPCmd.Flags().BoolVarP(&projectScope, "project", "p", false, "Write MCP config to project scope (.vscode/mcp.json). OpenCode always writes to opencode.json in the current folder")
}
