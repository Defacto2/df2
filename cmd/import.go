package cmd

import (
	"log"
	"os"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/dizzer"
	"github.com/spf13/cobra"
)

var imrar arg.Import

var importCmd = &cobra.Command{
	Use:     "import (path to .rar)",
	Short:   "Import a .rar archive collection containing information NFO and text files.",
	Aliases: []string{"i"},
	GroupID: "group1",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Usage(); err != nil {
				logr.Fatal(err)
			}
			return
		}
		if err := dizzer.Run(os.Stdout, os.Stderr, args[0]); err != nil {
			log.Fatal(err)
		}
	},
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(importCmd)
	importCmd.PersistentFlags().BoolVarP(&imrar.Insert, "insert", "i", false,
		"insert the found text files metadata to the database")
}
