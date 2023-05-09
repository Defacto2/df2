//nolint:gochecknoglobals,gochecknoinits
package cmd

import (
	"os"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/importer"
	"github.com/spf13/cobra"
)

var imrar arg.Import

var importCmd = &cobra.Command{
	Use:     "import (path to .rar)",
	Short:   "Import a .rar archive collection containing information NFO and text files.",
	Aliases: []string{"im"},
	GroupID: "group1",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Usage(); err != nil {
				logr.Fatal(err)
			}
			return
		}
		im := importer.Importer{
			RARFile: args[0],
			Insert:  imrar.Insert,
			Limit:   imrar.Limit,
			Config:  confg,
			Logger:  logr,
		}
		db, err := database.Connect(confg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		w := os.Stdout
		if err := im.Import(db, w); err != nil {
			logr.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.PersistentFlags().BoolVarP(&imrar.Insert, "insert", "i", false,
		"insert the found text files metadata to the database")
	importCmd.PersistentFlags().UintVarP(&imrar.Limit, "limit", "l", 0,
		"limit the total number of found text files to process")
}
