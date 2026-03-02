package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var myTaskCmd = &cobra.Command{
	Use:   "my-task <agent-id>",
	Short: "Print agent's current task summary",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys my-task: not implemented")
		return nil
	},
}

func init() { rootCmd.AddCommand(myTaskCmd) }
