package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [agent-id]",
	Short: "Stop agents and kill tmux panes",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys stop: not implemented")
		return nil
	},
}

func init() { rootCmd.AddCommand(stopCmd) }
