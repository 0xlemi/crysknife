package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display agent status, queue, and convoys",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys status: not implemented")
		return nil
	},
}

func init() {
	statusCmd.Flags().Bool("json", false, "Output raw JSON")
	rootCmd.AddCommand(statusCmd)
}
