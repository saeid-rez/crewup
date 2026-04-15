package ui

import (
	"fmt"

	"github.com/saeid-rez/crewup/internal/config"
	"github.com/saeid-rez/crewup/internal/mcp"
	"github.com/saeid-rez/crewup/internal/tools"
)

// PrintBanner prints the crewup welcome banner
func PrintBanner() {
	fmt.Println(`
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
// TODO: replace with bubbletea multi-select for a polished interactive UI.
func SelectTools(detected []tools.AITool) ([]tools.AITool, error) {
	fmt.Println("✅ Detected AI tools:\n")
	for i, t := range detected {
		fmt.Printf("  [%d] %s\n", i+1, t.Name)
	}

	fmt.Println("\nPress Enter to configure all, or type numbers separated by commas (e.g. 1,3):")
	// TODO: bubbletea checkbox list
	// For now, return all detected
	return detected, nil
}

// SelectAgentRoles lets user pick which agent roles to include in their crew.
// TODO: replace with bubbletea multi-select
func SelectAgentRoles() ([]config.AgentRole, error) {
	fmt.Println("\n🤖 Select agent roles for your crew:\n")
	for i, role := range config.DefaultRoles {
		fmt.Printf("  [%d] %-15s %s\n", i+1, role.Name, role.Description)
	}

	fmt.Println("\nPress Enter to include all roles, or type numbers (e.g. 1,2,3):")
	// TODO: bubbletea checkbox list
	return config.DefaultRoles, nil
}

// SelectMCPServers lets user pick which MCP servers to install.
// TODO: replace with bubbletea multi-select
func SelectMCPServers() ([]string, error) {
	fmt.Println("\n🔌 Popular MCP servers (optional):\n")
	for i, s := range mcp.Registry {
		fmt.Printf("  [%d] %-20s %s\n", i+1, s.Name, s.Description)
	}

	fmt.Println("\nPress Enter to skip, or type numbers (e.g. 1,2):")
	// TODO: bubbletea checkbox list
	return []string{}, nil
}

// PrintSuccess shows a summary of what was configured
func PrintSuccess(selectedTools []tools.AITool, roles []config.AgentRole, mcps []string) {
	fmt.Println("\n✅ Your crew is ready!\n")

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
	fmt.Println("Run `crewup add mcp <name>` to add more MCP servers later.\n")
}
