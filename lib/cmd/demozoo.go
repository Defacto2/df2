// nolint:gochecknoglobals
package cmd

import (
	"errors"

	"github.com/Defacto2/df2/lib/cmd/internal/arg"
	"github.com/Defacto2/df2/lib/cmd/internal/run"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
)

var dzf arg.Demozoo

// demozooCmd represents the demozoo command.
var demozooCmd = &cobra.Command{
	Use:     "demozoo",
	Short:   "Interact with Demozoo submissions.",
	Long:    "Manage upload submissions that rely on the API hosted on demozoo.org.",
	Aliases: []string{"d", "dz"},
	Example: `  df2 demozoo [--new|--all|--id] (--dry-run,--overwrite)
  df2 demozoo [--refresh|--sync|--ping|--download]`,
	Run: func(cmd *cobra.Command, args []string) {
		err := run.Demozoo(dzf)
		switch {
		case errors.Is(err, run.ErrDZFlag):
			if err := cmd.Usage(); err != nil {
				logs.Fatal(err)
			}
		case err != nil:
			logs.Fatal(err)
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(demozooCmd)
	demozooCmd.Flags().BoolVarP(&dzf.New, "new", "n", false,
		"scan for new demozoo submissions (recommended)")
	demozooCmd.Flags().BoolVar(&dzf.All, "all", false,
		"scan all files with demozoo links (SLOW)")
	demozooCmd.Flags().StringVarP(&dzf.ID, "id", "i", "",
		"update the empty data of a table id or uuid with its fetched demozoo data")
	demozooCmd.Flags().BoolVarP(&dzf.Simulate, "dry-run", "d", false,
		"simulate the fixes and display the expected changes")
	demozooCmd.Flags().BoolVar(&dzf.Overwrite, "overwrite", false,
		"rescan archives and overwrite all existing assets\n")
	demozooCmd.Flags().BoolVarP(&dzf.Refresh, "refresh", "r", false,
		"replace empty table data with fetched demozoo data (SLOW)\n"+
			"any demozoo ids with 404 are removed from the table")
	demozooCmd.Flags().BoolVar(&dzf.Pouet, "pouet", false,
		"scan the demozoo api for missing or deleted pouet ids [not implemented]")
	demozooCmd.Flags().BoolVarP(&dzf.Sync, "sync", "s", false,
		"scan the demozoo api for missing bbstros and cracktros (SLOW) [not implemented]\n")
	demozooCmd.Flags().UintVarP(&dzf.Ping, "ping", "p", 0,
		"fetch and display a production record from the demozoo API")
	demozooCmd.Flags().UintVarP(&dzf.Download, "download", "g", 0,
		"fetch and download a linked file from the demozoo API\n")
	demozooCmd.Flags().StringArrayVar(&dzf.Extract, "extract", make([]string, 0),
		`extracts and parses an archived file
requires two flags: --extract [filename] --extract [uuid]`)
	if err := demozooCmd.MarkFlagFilename("extract"); err != nil {
		logs.Fatal(err)
	}
	if err := demozooCmd.Flags().MarkHidden("extract"); err != nil {
		logs.Fatal(err)
	}
	demozooCmd.Flags().SortFlags = false
}
