package ui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/mattn/go-isatty"
	"github.com/saeid-rez/crewup/internal/config"
	"github.com/saeid-rez/crewup/internal/mcp"
	"github.com/saeid-rez/crewup/internal/tools"
)

func normalizePromptError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, huh.ErrUserAborted) {
		return ErrUserCancelled
	}
	return err
}

// isTTY returns true if both stdin and stdout are terminals (not piped/CI).
// Forms read from stdin, so we must check both to avoid hangs.
func isTTY() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}

// IsTTY is the exported version of isTTY for use by cmd packages.
func IsTTY() bool { return isTTY() }

// PrintBanner prints the crewup welcome banner
func PrintBanner() {
	fmt.Print(`
 ██████╗██████╗ ███████╗██╗    ██╗██╗   ██╗██████╗ 
██╔════╝██╔══██╗██╔════╝██║    ██║██║   ██║██╔══██╗
██║     ██████╔╝█████╗  ██║ █╗ ██║██║   ██║██████╔╝
██║     ██╔══██╗██╔══╝  ██║███╗██║██║   ██║██╔═══╝ 
╚██████╗██║  ██║███████╗╚███╔███╔╝╚██████╔╝██║     
 ╚═════╝╚═╝  ╚═╝╚══════╝ ╚══╝╚══╝  ╚═════╝ ╚═╝     

  Setup AI agent crews for any AI tool on your machine.
`)
}

// SelectTools shows detected tools and lets user pick which to configure.
func SelectTools(detected []tools.AITool) ([]tools.AITool, error) {
	if len(detected) == 1 {
		fmt.Printf("✅ Found: %s — auto-selected.\n", detected[0].Name)
		return detected, nil
	}

	// Non-TTY: auto-select all
	if !isTTY() {
		fmt.Println("Non-interactive mode: selecting all detected tools.")
		return detected, nil
	}

	items := make([]item, len(detected))
	for i, t := range detected {
		items[i] = item{id: t.ID, title: t.Name, desc: t.ConfigPath, selected: true}
	}

	selectedIDs, err := RunMultiSelect("🔍 Select AI tools to configure", items)
	if err != nil {
		return nil, err
	}

	// Filter detected by selectedIDs
	idSet := make(map[string]bool, len(selectedIDs))
	for _, id := range selectedIDs {
		idSet[id] = true
	}
	var result []tools.AITool
	for _, t := range detected {
		if idSet[t.ID] {
			result = append(result, t)
		}
	}
	return result, nil
}

// SelectAgentRoles lets user pick which agent roles to include in their crew.
func SelectAgentRoles() ([]config.AgentRole, error) {
	// Non-TTY: auto-select all
	if !isTTY() {
		fmt.Println("Non-interactive mode: selecting all agent roles.")
		return config.DefaultRoles, nil
	}

	items := make([]item, len(config.DefaultRoles))
	for i, r := range config.DefaultRoles {
		items[i] = item{id: r.ID, title: r.Name, desc: r.Description, selected: true}
	}

	selectedIDs, err := RunMultiSelect("🤖 Select agent roles for your crew", items)
	if err != nil {
		return nil, err
	}

	idSet := make(map[string]bool, len(selectedIDs))
	for _, id := range selectedIDs {
		idSet[id] = true
	}
	var result []config.AgentRole
	for _, r := range config.DefaultRoles {
		if idSet[r.ID] {
			result = append(result, r)
		}
	}
	return result, nil
}

// SelectMCPPresets lets user pick which MCP presets to install.
// Returns full MCPPreset structs so callers have access to Inputs.
// Non-TTY: returns empty slice (skip MCP).
func SelectMCPPresets() ([]mcp.MCPPreset, error) {
	// Non-TTY: skip MCP selection
	if !isTTY() {
		fmt.Println("Non-interactive mode: skipping MCP server selection.")
		return []mcp.MCPPreset{}, nil
	}

	items := make([]item, len(mcp.Presets))
	for i, p := range mcp.Presets {
		items[i] = item{id: p.ID, title: p.Name, desc: p.Description}
	}

	selectedIDs, err := RunMultiSelect("🔌 Popular MCP servers (optional)", items)
	if err != nil {
		return nil, err
	}

	idSet := make(map[string]bool, len(selectedIDs))
	for _, id := range selectedIDs {
		idSet[id] = true
	}
	var result []mcp.MCPPreset
	for _, p := range mcp.Presets {
		if idSet[p.ID] {
			result = append(result, p)
		}
	}
	return result, nil
}

// PromptInputFields prompts the user for each InputField in the preset.
// Non-TTY + any Required field → returns a descriptive error.
// Non-TTY + all optional → returns empty map, nil.
func PromptInputFields(preset mcp.MCPPreset) (map[string]string, error) {
	if len(preset.Inputs) == 0 {
		return map[string]string{}, nil
	}

	if !isTTY() {
		for _, f := range preset.Inputs {
			if f.Required {
				return nil, fmt.Errorf("preset %q requires field %q (%s) but running in non-interactive mode", preset.Name, f.Key, f.Label)
			}
		}
		return map[string]string{}, nil
	}

	ptrs := make([]*string, len(preset.Inputs))
	fields := make([]huh.Field, len(preset.Inputs))
	for i, f := range preset.Inputs {
		s := new(string)
		ptrs[i] = s
		inp := huh.NewInput().
			Title(f.Label).
			Value(s)
		if f.Sensitive {
			inp = inp.EchoMode(huh.EchoModePassword)
		}
		if f.Required {
			label := f.Label
			inp = inp.Validate(func(v string) error {
				if strings.TrimSpace(v) == "" {
					return fmt.Errorf("%s is required", label)
				}
				return nil
			})
		}
		fields[i] = inp
	}

	form := huh.NewForm(huh.NewGroup(fields...))
	if err := form.Run(); err != nil {
		return nil, normalizePromptError(err)
	}

	result := make(map[string]string, len(preset.Inputs))
	for i, f := range preset.Inputs {
		result[f.Key] = strings.TrimSpace(*ptrs[i])
	}
	return result, nil
}

// SelectTargetTools lets user pick which configured tools to install a preset for.
// Non-TTY: returns all tools.
func SelectTargetTools(configured []config.ToolInfo, presetName string) ([]config.ToolInfo, error) {
	if !isTTY() {
		return configured, nil
	}

	items := make([]item, len(configured))
	for i, t := range configured {
		items[i] = item{id: t.ID, title: t.Name, selected: true}
	}

	selectedIDs, err := RunMultiSelect(fmt.Sprintf("🛠  Install %s for which tools?", presetName), items)
	if err != nil {
		return nil, err
	}

	idSet := make(map[string]bool, len(selectedIDs))
	for _, id := range selectedIDs {
		idSet[id] = true
	}
	var result []config.ToolInfo
	for _, t := range configured {
		if idSet[t.ID] {
			result = append(result, t)
		}
	}
	return result, nil
}

// PrintSuccess shows a summary of what was configured
func PrintSuccess(selectedTools []tools.AITool, roles []config.AgentRole, mcps []string) {
	fmt.Println("\n✅ Your crew is ready!")
	fmt.Println()

	fmt.Println("🤖 Agents configured:")
	for _, r := range roles {
		fmt.Printf("  • %s\n", r.Name)
	}

	fmt.Println("\n🛠  Configured for:")
	for _, t := range selectedTools {
		fmt.Printf("  • %s\n", t.Name)
	}

	if len(mcps) > 0 {
		fmt.Println("\n🔌 MCP servers added:")
		for _, m := range mcps {
			fmt.Printf("  • %s\n", m)
		}
	}

	fmt.Println("\nRun `crewup list` to see your setup anytime.")
	fmt.Println("Run `crewup add mcp <name>` to add more MCP servers later.")
	fmt.Println()
}

// PrintConfig renders the current crewup config as a formatted table.
func PrintConfig(cfg *config.CrewConfig) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	fmt.Println()
	fmt.Println(titleStyle.Render("  crewup — Your Crew Setup"))
	fmt.Println()

	// Tools
	fmt.Print(labelStyle.Render("  🛠  Tools:    "))
	if len(cfg.Tools) == 0 {
		fmt.Println(dimStyle.Render("none (run `crewup init` to configure tools)"))
	} else {
		names := make([]string, len(cfg.Tools))
		for i, t := range cfg.Tools {
			names[i] = t.Name
		}
		fmt.Println(valueStyle.Render(strings.Join(names, ", ")))
	}

	// Agent Roles
	fmt.Print(labelStyle.Render("  🤖 Agents:   "))
	if len(cfg.AgentRoles) == 0 {
		fmt.Println(dimStyle.Render("none"))
	} else {
		names := make([]string, len(cfg.AgentRoles))
		for i, r := range cfg.AgentRoles {
			names[i] = r.Name
		}
		fmt.Println(valueStyle.Render(strings.Join(names, ", ")))
	}

	// MCP Servers
	fmt.Print(labelStyle.Render("  🔌 MCP:       "))
	if len(cfg.MCPServers) == 0 {
		fmt.Println(dimStyle.Render("none"))
	} else {
		fmt.Println(valueStyle.Render(strings.Join(cfg.MCPServers, ", ")))
	}

	fmt.Println()
}
