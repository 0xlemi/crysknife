package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var slingCmd = &cobra.Command{
	Use:   "sling <worker-id>",
	Short: "Assign work to a worker agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys sling: not implemented")
		return nil
	},
}

func init() {
	slingCmd.Flags().String("task", "", "Task name")
	slingCmd.Flags().String("tier", "standard", "Workflow tier: full, standard, quick")
	slingCmd.Flags().String("area", "", "Filesystem area restriction")
	slingCmd.Flags().String("branch", "", "Git branch name")
	slingCmd.Flags().Bool("from-queue", false, "Pick next task from queue")
	rootCmd.AddCommand(slingCmd)
}
