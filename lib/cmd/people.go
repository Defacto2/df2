package cmd

import (
	"github.com/Defacto2/df2/lib/people"
	"github.com/spf13/cobra"
)

type pplFlags struct {
	//cronjob  bool
	filter   string
	format   string
	progress bool
}

var pf pplFlags

// peopleCmd represents the authors command
var peopleCmd = &cobra.Command{
	Use:   "people",
	Short: "A HTML snippet generator to list people",
	Run: func(cmd *cobra.Command, args []string) {
		filterFlag(people.Wheres(), pf.filter)
		var req people.Request
		if filterFlag(fmtflags, pf.format); pf.format != "" {
			req = people.Request{Filter: pf.filter, Progress: pf.progress}
		}
		switch pf.format {
		case "html", "h", "":
			people.HTML("", req)
		case "text", "t":
			people.Print(req)
		}
	},
}

func init() {
	rootCmd.AddCommand(peopleCmd)
	peopleCmd.Flags().StringVarP(&pf.filter, "filter", "f", "", "filter groups (default all)\noptions: "+people.Filters)
	peopleCmd.Flags().StringVarP(&pf.format, "format", "t", "", "output format (default html)\noptions: html,text")
}
