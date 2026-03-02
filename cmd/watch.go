package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Monitor agents with auto-nudge and auto-restart",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys watch: not implemented")
		return nil
	},
}

func init() {
	watchCmd.Flags().Int("interval", 30, "Check interval in seconds")
	watchCmd.Flags().Int("nudge-after", 120, "Nudge idle agents after N seconds")
	watchCmd.Flags().Int("restart-after", 300, "Restart dead agents after N seconds")
	watchCmd.Flags().Bool("no-nudge", false, "Disable auto-nudge")
	watchCmd.Flags().Bool("no-restart", false, "Disable auto-restart")
	rootCmd.AddCommand(watchCmd)
}
