package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// shrinkCmd represents the compact command.
var shrinkCmd = &cobra.Command{
	Use:     "shrink",
	Short:   "Reduces the space used in directories",
	Aliases: []string{"s"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello world.")
	},
}

func init() {
	rootCmd.AddCommand(shrinkCmd)
}
