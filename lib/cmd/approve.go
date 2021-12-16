// nolint:gochecknoglobals
package cmd

import (
	"log"

	"github.com/Defacto2/df2/lib/cmd/internal/arg"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/people"
	"github.com/spf13/cobra"
)

var appr arg.Approve

var approveCmd = &cobra.Command{
	Use:     "approve",
	Short:   "Approve the file records that are ready to go live",
	Aliases: []string{"a"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := database.Approve(appr.Verbose); err != nil {
			log.Fatal(err)
		}
		if err := database.Fix(); err != nil {
			log.Fatal(err)
		}
		if err := groups.Fix(gf.Simulate); err != nil {
			log.Fatal(err)
		}
		if err := people.Fix(gf.Simulate); err != nil {
			log.Fatal(err)
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(approveCmd)
	approveCmd.Flags().BoolVarP(&appr.Verbose, "verbose", "v", false,
		"display all file records that qualify to go public")
	approveCmd.PersistentFlags().BoolVarP(&gf.Simulate, "dry-run", "d", false,
		"simulate the fixes and display the expected changes")
}
