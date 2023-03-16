//nolint:gochecknoglobals
package cmd

import (
	"os"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/cmd/internal/run"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/Defacto2/df2/pkg/sitemap"
	"github.com/spf13/cobra"
)

var tests arg.TestSite

var testCmd = &cobra.Command{
	Use:     "test",
	Short:   "Test various features of the website or database that cannot be fixed with automation.",
	Aliases: []string{"t"},
	GroupID: "group3",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Usage(); err != nil {
			log.Fatal(err)
		}
	},
}

var testGroupNames = &cobra.Command{
	Use:     "names",
	Short:   "Scans over the various group names and attempts to match possible misnamed duplicates.",
	Aliases: []string{"n"},
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		if err := groups.MatchStdOut(db, os.Stdout); err != nil {
			log.Fatal(err)
		}
	},
}

var testURLsCmd = &cobra.Command{
	Use:     "urls",
	Short:   "Test the website by pinging or downloading a large, select number of URLs.",
	Aliases: []string{"u"},
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		base := sitemap.Base
		if tests.LocalHost {
			base = sitemap.LocalBase
		}
		if err := run.TestSite(db, os.Stdout, base); err != nil {
			log.Fatal(err)
		}
	},
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(testCmd)
	testCmd.AddCommand(testGroupNames)
	testCmd.AddCommand(testURLsCmd)
	testURLsCmd.Flags().BoolVarP(&tests.LocalHost, "localhost", "l", true,
		"run the tests to target "+sitemap.LocalBase)
}
