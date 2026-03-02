package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done <worker-id>",
	Short: "Mark worker as done, add branch to merge queue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys done: not implemented")
		return nil
	},
}

func init() { rootCmd.AddCommand(doneCmd) }
