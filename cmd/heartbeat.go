package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var heartbeatCmd = &cobra.Command{
	Use:   "heartbeat <agent-id>",
	Short: "Update agent last_activity timestamp",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys heartbeat: not implemented")
		return nil
	},
}

func init() { rootCmd.AddCommand(heartbeatCmd) }
