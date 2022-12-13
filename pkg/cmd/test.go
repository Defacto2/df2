package cmd

import (
	"log"

	"github.com/Defacto2/df2/pkg/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/cmd/internal/run"
	"github.com/Defacto2/df2/pkg/sitemap"
	"github.com/spf13/cobra"
)

// todo: random people

var tests arg.TestSite

var testCmd = &cobra.Command{
	Use:     "test",
	Short:   "Test the website by pinging or downloading a large, select number of URLs.",
	Aliases: []string{"t"},
	GroupID: "group3",
	Run: func(cmd *cobra.Command, args []string) {
		base := sitemap.Base
		if tests.LocalHost {
			base = sitemap.LocalBase
		}
		if err := run.TestSite(base); err != nil {
			log.Fatal(err)
		}
	},
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().BoolVarP(&tests.LocalHost, "localhost", "l", true,
		"run the tests to target "+sitemap.LocalBase)
}
