// nolint:gochecknoglobals
package cmd

import (
	"log"
	"sort"
	"strings"

	"github.com/Defacto2/df2/lib/assets"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/spf13/cobra"
)

type cleanFlags struct {
	delete   bool
	humanise bool
	makeDirs bool
	target   string
}

var (
	clf     cleanFlags
	targets = []string{"all", "download", "emulation", "image"}
)

// cleanCmd represents the clean command.
var cleanCmd = &cobra.Command{
	Use:     "clean",
	Short:   "Discover or clean orphan files",
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {
		directories.Init(clf.makeDirs)
		if err := assets.Clean(clf.target, clf.delete, clf.humanise); err != nil {
			log.Fatal(err)
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().StringVarP(&clf.target, "target", "t", "all",
		"what file section to clean"+options(targets...))
	cleanCmd.Flags().BoolVarP(&clf.delete, "delete", "x", false,
		"erase all discovered files to free up drive space")
	cleanCmd.Flags().BoolVar(&clf.humanise, "humanise", true,
		"humanise file sizes and date times")
	cleanCmd.Flags().BoolVar(&clf.makeDirs, "makedirs", false,
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
