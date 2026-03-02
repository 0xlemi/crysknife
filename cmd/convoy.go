package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var convoyCmd = &cobra.Command{
	Use:   "convoy",
	Short: "Feature-level tracking",
}

var convoyCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a convoy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys convoy create: not implemented")
		return nil
	},
}

var convoyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all convoys",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys convoy list: not implemented")
		return nil
	},
}

var convoyStatusCmd = &cobra.Command{
	Use:   "status <name>",
	Short: "Show convoy task statuses",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys convoy status: not implemented")
		return nil
	},
}

func init() {
	convoyCreateCmd.Flags().StringSlice("tasks", nil, "Task IDs to include")
	convoyCmd.AddCommand(convoyCreateCmd, convoyListCmd, convoyStatusCmd)
	rootCmd.AddCommand(convoyCmd)
}
