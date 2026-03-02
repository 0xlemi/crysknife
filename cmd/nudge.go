package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var nudgeCmd = &cobra.Command{
	Use:   "nudge <agent-id>",
	Short: "Nudge an idle agent to resume work",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys nudge: not implemented")
		return nil
	},
}

func init() {
	nudgeCmd.Flags().Bool("all", false, "Nudge all idle agents")
	rootCmd.AddCommand(nudgeCmd)
}
