package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Manage the work queue",
}

var queueAddCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a task to the queue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys queue add: not implemented")
		return nil
	},
}

var queueListCmd = &cobra.Command{
	Use:   "list",
	Short: "List queued tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys queue list: not implemented")
		return nil
	},
}

var queueRemoveCmd = &cobra.Command{
	Use:   "remove <task-id>",
	Short: "Remove a task from the queue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys queue remove: not implemented")
		return nil
	},
}

func init() {
	queueAddCmd.Flags().String("tier", "standard", "Workflow tier")
	queueAddCmd.Flags().String("area", "", "Filesystem area")
	queueCmd.AddCommand(queueAddCmd, queueListCmd, queueRemoveCmd)
	rootCmd.AddCommand(queueCmd)
}
