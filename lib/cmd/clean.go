package cmd

import (
	"sort"
	"strings"

	"github.com/Defacto2/df2/lib/assets"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Discover or clean orphan files",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		directories.Init(makeDirs)
		assets.Clean(target, delete, humanize)
	},
}

var (
	delete   bool
	humanize bool
	makeDirs bool
	target   string
	targets  []string = []string{"all", "download", "emulation", "image"}
)

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().StringVarP(&target, "target", "t", "all", "what file section to clean"+options(targets))
	cleanCmd.Flags().BoolVarP(&delete, "delete", "d", false, "erase all discovered files to free up drive space")
	cleanCmd.Flags().BoolVar(&humanize, "humanize", true, "humanize file sizes and date times")
	cleanCmd.Flags().BoolVar(&makeDirs, "makedirs", false, "generate uuid directories and placeholder files")
	cleanCmd.Flags().SortFlags = false
	_ = cleanCmd.Flags().MarkHidden("makedirs")
}

func options(a []string) string {
	sort.Strings(a)
	return "\noptions: " + strings.Join(a, ", ")
}

func valid(a []string, x string) bool {
	for _, b := range a {
		if b == x {
			return true
		}
	}
	return false
}
