package cmd

import (
	"fmt"
	"os"

	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/people"
	"github.com/Defacto2/df2/lib/sitemap"
	"github.com/spf13/cobra"
)

// htmlCmd represents the html command
var htmlCmd = &cobra.Command{
	Use:   "html",
	Short: "HTML and sitemap generator",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Usage()
			os.Exit(0)
		}
		_ = cmd.Usage()
		logs.Check(fmt.Errorf("invalid command %v please use one of the available fix commands", args[0]))
	},
}

func init() {
	rootCmd.AddCommand(htmlCmd)
	htmlCmd.AddCommand(groupCmd)
	groupCmd.Flags().StringVarP(&gf.filter, "filter", "f", "", "filter groups (default all)\noptions: "+groups.Filters)
	groupCmd.Flags().BoolVarP(&gf.counts, "count", "c", false, "display the file totals for each group (SLOW)")
	groupCmd.Flags().BoolVarP(&gf.progress, "progress", "p", true, "show a progress indicator while fetching a large number of records")
	groupCmd.Flags().BoolVarP(&gf.cronjob, "cronjob", "j", false, "run in cronjob automated mode, ignores all other arguments")
	groupCmd.Flags().StringVarP(&gf.format, "format", "t", "", "output format (default html)\noptions: datalist,html,text")
	groupCmd.Flags().BoolVarP(&gf.init, "initialism", "i", false, "display the acronyms and initialisms for groups (SLOW)")
	htmlCmd.AddCommand(peopleCmd)
	peopleCmd.Flags().StringVarP(&pf.filter, "filter", "f", "", "filter groups (default all)\noptions: "+people.Filters)
	peopleCmd.Flags().StringVarP(&pf.format, "format", "t", "", "output format (default html)\noptions: datalist,html,text")
	htmlCmd.AddCommand(sitemapCmd)
}

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
		filterFlag(groups.Wheres(), "filter", gf.filter)
		var req = groups.Request{Filter: gf.filter, Counts: gf.counts, Initialisms: gf.init, Progress: gf.progress}
		switch gf.format {
		case "datalist", "dl", "d":
			req.DataList("")
		case "html", "h", "":
			req.HTML("")
		case "text", "t":
			groups.Print(req)
		}
	},
}

type pplFlags struct {
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
		filterFlag(people.Wheres(), "filter", pf.filter)
		var req people.Request
		if filterFlag(fmtflags, "format", pf.format); pf.format != "" {
			req = people.Request{Filter: pf.filter, Progress: pf.progress}
		}
		switch pf.format {
		case "datalist", "dl", "d":
			people.DataList("", req)
		case "html", "h", "":
			people.HTML("", req)
		case "text", "t":
			people.Print(req)
		}
	},
}

// sitemapCmd represents the sitemap command
var sitemapCmd = &cobra.Command{
	Use:   "sitemap",
	Short: "A site map generator",
	Run: func(cmd *cobra.Command, args []string) {
		sitemap.Create()
	},
}
