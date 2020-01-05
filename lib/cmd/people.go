package cmd

import (
	"github.com/Defacto2/df2/lib/people"
	"github.com/spf13/cobra"
)

type pplFlags struct {
	counts   bool
	cronjob  bool
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
		r := people.Request{Filter: pf.filter}
		people.Print(r)
	},
}

func init() {
	rootCmd.AddCommand(peopleCmd)
	peopleCmd.Flags().StringVarP(&pf.filter, "filter", "f", "", "filter groups (default all)\noptions: "+people.Filters)
}
