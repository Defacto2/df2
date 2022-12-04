//nolint:gochecknoglobals
package cmd

import (
	"log"

	"github.com/Defacto2/df2/pkg/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/Defacto2/df2/pkg/people"
	"github.com/spf13/cobra"
)

var appr arg.Approve

var approveCmd = &cobra.Command{
	Use:     "approve",
	Short:   "Approve the records that are ready to go live.",
	Aliases: []string{"a"},
	GroupID: "group1",
	Run: func(cmd *cobra.Command, args []string) {
		if err := database.Approve(appr.Verbose); err != nil {
			log.Print(err)
		}
		if err := database.Fix(); err != nil {
			log.Print(err)
		}
		if err := groups.Fix(); err != nil {
			log.Print(err)
		}
		if err := people.Fix(); err != nil {
			log.Print(err)
		}
	},
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(approveCmd)
	approveCmd.Flags().BoolVar(&appr.Verbose, "verbose", false,
		"display all file records that qualify to go public")
}
