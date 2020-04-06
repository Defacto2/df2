package cmd

import (
	"github.com/Defacto2/df2/lib/database"
	"github.com/spf13/cobra"
)

var approveVerb bool

var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve the file records that are ready to go live",
	Run: func(cmd *cobra.Command, args []string) {
		database.Approve(approveVerb)
	},
}

func init() {
	rootCmd.AddCommand(approveCmd)
	approveCmd.Flags().BoolVarP(&approveVerb, "verbose", "v", false, "display all file records that qualify to go public")
	approveCmd.PersistentFlags().BoolVarP(&simulate, "dry-run", "d", false, "simulate the fixes and display the expected changes")
}
