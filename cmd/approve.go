//nolint:gochecknoglobals
package cmd

import (
	"os"

	"github.com/Defacto2/df2/cmd/internal/arg"
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
		w := os.Stdout
		if err := database.Approve(w, log, cfg.IncomingFiles, appr.Verbose); err != nil {
			log.Info(err)
		}
		if err := database.Fix(w, log); err != nil {
			log.Info(err)
		}
		if err := groups.Fix(w); err != nil {
			log.Info(err)
		}
		if err := people.Fix(w); err != nil {
			log.Info(err)
		}
	},
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(approveCmd)
	approveCmd.Flags().BoolVar(&appr.Verbose, "verbose", false,
		"display all file records that qualify to go public")
}
