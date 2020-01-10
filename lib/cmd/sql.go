package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// sqlCmd represents the sql command
var sqlCmd = &cobra.Command{
	Use:   "sql",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sql called")
	},
}

func init() {
	rootCmd.AddCommand(sqlCmd)
}
