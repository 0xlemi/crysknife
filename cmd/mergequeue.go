package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mergeQueueCmd = &cobra.Command{
	Use:   "merge-queue",
	Short: "Show branches ready to merge",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys merge-queue: not implemented")
		return nil
	},
}

func init() {
	mergeQueueCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(mergeQueueCmd)
}
