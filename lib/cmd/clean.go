// nolint:gochecknoglobals
package cmd

import (
	"log"
	"sort"
	"strings"

	"github.com/Defacto2/df2/lib/assets"
	"github.com/Defacto2/df2/lib/cmd/internal/arg"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/spf13/cobra"
)

var (
	clf     arg.Clean
	targets = []string{"all", "download", "emulation", "image"}
)

// cleanCmd represents the clean command.
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Discover or clean orphan files.",
	Long: `Discover or clean orphan files found on the web server.
Files are considered orphan when they do not match to a correlating record in the
database. These can include UUID named thumbnails, previews, textfile previews.`,
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {
		directories.Init(clf.MakeDirs)
		if err := assets.Clean(clf.Target, clf.Delete, clf.Humanise); err != nil {
			log.Fatal(err)
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().StringVarP(&clf.Target, "target", "t", "all",
		"what file section to clean"+options(targets...))
	cleanCmd.Flags().BoolVarP(&clf.Delete, "delete", "x", false,
		"erase all discovered files to free up drive space")
	cleanCmd.Flags().BoolVar(&clf.Humanise, "humanise", true,
		"humanise file sizes and date times")
	cleanCmd.Flags().BoolVar(&clf.MakeDirs, "makedirs", false,
		"generate uuid directories and placeholder files")
	cleanCmd.Flags().SortFlags = false
	if err := cleanCmd.Flags().MarkHidden("makedirs"); err != nil {
		log.Fatal(err)
	}
}

func options(a ...string) string {
	sort.Strings(a)
	return "\noptions: " + strings.Join(a, ",")
}
