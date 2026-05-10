package cmd

import (
	"fmt"
	"os"

	"github.com/saeid-rez/crewup/internal/config"
	"github.com/saeid-rez/crewup/pkg/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show your current crewup configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList()
	},
}

func runList() error {
	cfg, err := config.Load()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No configuration found. Run `crewup init` to get started.")
			return nil
		}
		return fmt.Errorf("could not load config: %w", err)
	}
	ui.PrintConfig(cfg)
	return nil
}
