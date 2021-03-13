package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/people"
)

var approveVerb bool

var approveCmd = &cobra.Command{
	Use:     "approve",
	Short:   "Approve the file records that are ready to go live",
	Aliases: []string{"a"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := database.Approve(approveVerb); err != nil {
			log.Fatal(err)
		}
		if err := database.Fix(); err != nil {
			log.Fatal(err)
		}
		if err := groups.Fix(simulate); err != nil {
			log.Fatal(err)
		}
		if err := people.Fix(simulate); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(approveCmd)
	approveCmd.Flags().BoolVarP(&approveVerb, "verbose", "v", false, "display all file records that qualify to go public")
	approveCmd.PersistentFlags().BoolVarP(&simulate, "dry-run", "d", false, "simulate the fixes and display the expected changes")
}
