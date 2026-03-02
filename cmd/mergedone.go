package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mergeDoneCmd = &cobra.Command{
	Use:   "merge-done <branch>",
	Short: "Mark merge complete or failed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys merge-done: not implemented")
		return nil
	},
}

func init() {
	mergeDoneCmd.Flags().String("failed", "", "Mark as failed with reason")
	rootCmd.AddCommand(mergeDoneCmd)
}
