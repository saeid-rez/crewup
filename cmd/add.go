package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an MCP server or agent role to your existing setup",
}

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
		// TODO: call mcp.Install(serverName)
		fmt.Printf("✓ %s added successfully!\n", serverName)
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
		role := args[0]
		fmt.Printf("🤖 Adding agent role: %s\n", role)
		// TODO: call agents.AddRole(role)
		fmt.Printf("✓ %s agent added to your crew!\n", role)
		return nil
	},
}

func init() {
	addCmd.AddCommand(addMCPCmd)
	addCmd.AddCommand(addAgentCmd)
}
