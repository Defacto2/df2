package cmd

import (
	"fmt"
	"strings"

	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
)

const enforced = "bbs,ftp,group,magazine"

var (
	groupCntFiles bool
	groupCronJob  bool
	groupFilter   string
	groupProgress bool
)

// groupCmd represents the html command
var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "A HTML snippet generator to list groups",
	Run: func(cmd *cobra.Command, args []string) {
		if groupCronJob {
			groups.Cronjob()
			return
		}
		validateFilter()
		groups.HTML(groupFilter, groupCntFiles, groupProgress, "")
	},
}

func init() {
	rootCmd.AddCommand(groupCmd)
	groupCmd.Flags().StringVarP(&groupFilter, "filter", "f", "", "filter groups (default all)\noptions: "+enforced)
	groupCmd.Flags().BoolVarP(&groupCntFiles, "count", "c", false, "display the file totals for each group (SLOW)")
	groupCmd.Flags().BoolVarP(&groupProgress, "progress", "p", true, "show a progress indicator while fetching a large number of records")
	groupCmd.Flags().BoolVarP(&groupCronJob, "cronjob", "j", false, "run in cronjob automated mode, ignores all other arguments")
}

// validateFilter compairs the value of the filter flag against the list of values in the enforced const.
func validateFilter() {
	okay := false
	list := strings.Split(enforced, ",")
	println("validating")
	for _, n := range list {
		if groupFilter == n {
			okay = true
		}
	}
	if !okay {
		logs.Check(fmt.Errorf("unsupported filter flag %q, valid flags: %s", groupFilter, enforced))
	}
}
