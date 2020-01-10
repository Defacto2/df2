package cmd

import (
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
)

// demozooCmd represents the demozoo command
var demozooCmd = &cobra.Command{
	Use:   "demozoo",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		logs.Println("demozoo called")
	},
}

func init() {
	rootCmd.AddCommand(demozooCmd)
}
