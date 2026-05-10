package ui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/saeid-rez/crewup/internal/config"
	"github.com/saeid-rez/crewup/internal/mcp"
	"github.com/saeid-rez/crewup/internal/tools"
)

// isTTY returns true if both stdin and stdout are terminals (not piped/CI).
// Forms read from stdin, so we must check both to avoid hangs.
func isTTY() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}

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

	opts := make([]huh.Option[string], len(detected))
	for i, t := range detected {
		opts[i] = huh.NewOption(t.Name+" ("+t.ConfigPath+")", t.ID).Selected(true)
	}

	var selectedIDs []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("🔍 Select AI tools to configure").
				Description("Space to toggle, Enter to confirm").
				Options(opts...).
				Value(&selectedIDs),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Println("Setup cancelled.")
			os.Exit(0)
		}
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

	opts := make([]huh.Option[string], len(config.DefaultRoles))
	for i, r := range config.DefaultRoles {
		opts[i] = huh.NewOption(r.Name+" — "+r.Description, r.ID).Selected(true)
	}

	var selectedIDs []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("🤖 Select agent roles for your crew").
				Description("Space to toggle, Enter to confirm").
				Options(opts...).
				Value(&selectedIDs),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Println("Setup cancelled.")
			os.Exit(0)
		}
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

// SelectMCPServers lets user pick which MCP servers to install.
func SelectMCPServers() ([]string, error) {
	// Non-TTY: skip MCP selection
	if !isTTY() {
		fmt.Println("Non-interactive mode: skipping MCP server selection.")
		return []string{}, nil
	}

	opts := make([]huh.Option[string], len(mcp.Registry))
	for i, s := range mcp.Registry {
		opts[i] = huh.NewOption(s.Name+" — "+s.Description, s.ID)
	}

	var selectedIDs []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("🔌 Popular MCP servers (optional)").
				Description("Space to toggle, Enter to confirm (or skip)").
				Options(opts...).
				Value(&selectedIDs),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Println("Setup cancelled.")
			os.Exit(0)
		}
		return nil, err
	}

	return selectedIDs, nil
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
