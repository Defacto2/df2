// nolint:gochecknoglobals
package cmd

import (
	"errors"

	"github.com/Defacto2/df2/pkg/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/cmd/internal/run"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/spf13/cobra"
)

var apis arg.Apis

// apisCmd represents the demozoo command.
var apisCmd = &cobra.Command{
	Use:   "apis",
	Short: "Batch data synchronization with remote APIs.",
	Long: `Run batch data synchronizations with the remote APIs hosted
on demozoo.org and pouet.net. All these commands are SLOW and
require the parsing of 10,000s of records.`,
	Aliases: []string{"ap", "api"},
	Example: `  df2 apis [--refresh|--pouet|--msdos|--windows]`,
	Run: func(cmd *cobra.Command, args []string) {
		err := run.Apis(apis)
		switch {
		case errors.Is(err, run.ErrArgFlag):
			if err := cmd.Usage(); err != nil {
				logs.Fatal(err)
			}
		case err != nil:
			logs.Fatal(err)
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(apisCmd)
	apisCmd.Flags().BoolVarP(&apis.Refresh, "refresh", "r", false,
		"replace any empty data cells with fetched demozoo data\n"+
			"demozoo ids that return 404 not found errors are removed")
	apisCmd.Flags().BoolVarP(&apis.Pouet, "pouet", "p", false,
		"sync local files with pouet ids linked on demozoo")
	apisCmd.Flags().BoolVarP(&apis.SyncDos, "msdos", "d", false,
		"scan demozoo for missing local msdos bbstros and cracktros")
	apisCmd.Flags().BoolVarP(&apis.SyncWin, "windows", "w", false,
		"scan demozoo for missing local windows bbstros and cracktros")
	apisCmd.Flags().SortFlags = false
}
