//nolint:gochecknoglobals
package cmd

import (
	"os"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/assets"
	"github.com/Defacto2/df2/pkg/database/msql"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/spf13/cobra"
)

var clean arg.Clean

// cleanCmd represents the clean command.
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Discover or clean orphan files.",
	Long: `Discover or clean orphan files found on the web server.
Files are considered orphan when they do not match to a correlating record in the
database. These can include UUID named thumbnails, previews, textfile previews.`,
	Aliases: []string{"c"},
	GroupID: "group2",
	Run: func(cmd *cobra.Command, args []string) {
		directories.Init(cfg, clean.MakeDirs)
		db, err := msql.Connect(cfg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		c := assets.Clean{
			Name:   clean.Target,
			Remove: clean.Delete,
			Human:  clean.Humanise,
			Config: cfg,
		}
		if err := c.Walk(db, os.Stdout); err != nil {
			logr.Error(err)
		}
	},
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().StringVarP(&clean.Target, "target", "t", "all",
		"which file section to clean"+arg.CleanOpts(arg.Targets()...))
	cleanCmd.Flags().BoolVarP(&clean.Delete, "delete", "x", false,
		"erase all discovered files to free up drive space")
	cleanCmd.Flags().BoolVar(&clean.Humanise, "humanise", true,
		"humanise file sizes and date times")
	cleanCmd.Flags().BoolVar(&clean.MakeDirs, "makedirs", false,
		"generate uuid directories and placeholder files")
	cleanCmd.Flags().SortFlags = false
	if err := cleanCmd.Flags().MarkHidden("makedirs"); err != nil {
		logr.Error(err)
	}
}
