package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// demozooCmd represents the demozoo command
var demozooCmd = &cobra.Command{
	Use:   "demozoo",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("demozoo called")
	},
}

func init() {
	rootCmd.AddCommand(demozooCmd)
}
