package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start tmux session with all agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys start: not implemented")
		return nil
	},
}

func init() {
	startCmd.Flags().IntP("workers", "w", 4, "Number of worker agents")
	startCmd.Flags().String("agent", "", "Start only a specific agent")
	rootCmd.AddCommand(startCmd)
}
