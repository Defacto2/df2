package cmd

import (
	"github.com/Defacto2/df2/lib/groups"
	"github.com/spf13/cobra"
)

type groupFlags struct {
	counts   bool
	cronjob  bool
	filter   string
	format   string
	init     bool
	progress bool
}

var gf groupFlags

// groupCmd represents the html command
var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "A HTML snippet generator to list groups",
	Run: func(cmd *cobra.Command, args []string) {
		if gf.cronjob {
			groups.Cronjob()
			return
		}
		filterFlag(groups.Wheres(), gf.filter)
		var req groups.Request
		if filterFlag(fmtflags, gf.format); gf.format != "" {
			req = groups.Request{Filter: gf.filter, Counts: gf.counts, Initialisms: gf.init, Progress: gf.progress}
		}
		switch gf.format {
		case "html", "h", "":
			req.HTML("")
		case "text", "t":
			groups.Print(req)
		}
	},
}

func init() {
	rootCmd.AddCommand(groupCmd)
	groupCmd.Flags().StringVarP(&gf.filter, "filter", "f", "", "filter groups (default all)\noptions: "+groups.Filters)
	groupCmd.Flags().BoolVarP(&gf.counts, "count", "c", false, "display the file totals for each group (SLOW)")
	groupCmd.Flags().BoolVarP(&gf.progress, "progress", "p", true, "show a progress indicator while fetching a large number of records")
	groupCmd.Flags().BoolVarP(&gf.cronjob, "cronjob", "j", false, "run in cronjob automated mode, ignores all other arguments")
	groupCmd.Flags().StringVarP(&gf.format, "format", "t", "", "output format (default html)\noptions: html,text")
	groupCmd.Flags().BoolVarP(&gf.init, "initialism", "i", false, "display the acronyms and initialisms for groups (SLOW)")
}
