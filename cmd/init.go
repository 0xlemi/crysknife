package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Crysknife in the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("crys init: not implemented")
		return nil
	},
}

func init() { rootCmd.AddCommand(initCmd) }
