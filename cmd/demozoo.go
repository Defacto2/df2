//nolint:gochecknoglobals
package cmd

import (
	"errors"
	"os"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/cmd/internal/run"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/spf13/cobra"
)

var zoo arg.Demozoo

// demozooCmd represents the demozoo command.
var demozooCmd = &cobra.Command{
	Use:   "demozoo",
	Short: "Interact with Demozoo submissions.",
	Long: `Manage upload submissions that rely on the API hosted on demozoo.org.
There are additional Demozoo commands found under the api command.`,
	Aliases: []string{"d", "dz"},
	GroupID: "group3",
	Example: `  df2 demozoo [--new|--all|--releases|--id] (--overwrite)
  df2 demozoo [--ping|--download]`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(confg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		err = run.Demozoo(db, os.Stdout, logr, confg, zoo)
		switch {
		case errors.Is(err, run.ErrNothing):
			if err := cmd.Usage(); err != nil {
				logr.Fatal(err)
			}
		case err != nil:
			logr.Error(err)
		}
	},
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(demozooCmd)
	demozooCmd.Flags().BoolVarP(&zoo.New, "new", "n", false,
		"scan for new demozoo submissions (recommended)")
	demozooCmd.Flags().BoolVar(&zoo.All, "all", false,
		"scan all local files entries with demozoo links (SLOW)")
	demozooCmd.Flags().UintVar(&zoo.Releaser, "releases", 0,
		"add to the local files all the productions of a demozoo scener")
	demozooCmd.Flags().StringVarP(&zoo.ID, "id", "i", "",
		"replace any empty data cells of a local file with linked demozoo data")
	demozooCmd.Flags().BoolVar(&zoo.Overwrite, "overwrite", false,
		"rescan archives and overwrite all existing assets\n")
	demozooCmd.Flags().UintVarP(&zoo.Ping, "ping", "p", 0,
		"fetch and display a production record from the demozoo API")
	demozooCmd.Flags().UintVarP(&zoo.Download, "download", "g", 0,
		"fetch and download a linked file from the demozoo API\n")
	demozooCmd.Flags().StringArrayVar(&zoo.Extract, "extract", make([]string, 0),
		`extracts and parses an archived file
requires two flags: --extract [filename] --extract [uuid]`)
	if err := demozooCmd.MarkFlagFilename("extract"); err != nil {
		logr.Error(err)
	}
	if err := demozooCmd.Flags().MarkHidden("extract"); err != nil {
		logr.Error(err)
	}
	demozooCmd.Flags().SortFlags = false
}
