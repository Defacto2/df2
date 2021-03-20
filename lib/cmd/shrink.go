package cmd

import (
	"fmt"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
)

// shrinkCmd represents the compact command.
var shrinkCmd = &cobra.Command{
	Use:     "shrink",
	Short:   "Reduces the space used in directories",
	Aliases: []string{"s"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Shrinking directories")
		if err := directories.SQL(); err != nil {
			logs.Danger(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(shrinkCmd)
}
