package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/people"
	"github.com/Defacto2/df2/lib/recent"
	"github.com/Defacto2/df2/lib/sitemap"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// outputCmd represents the output command.
var outputCmd = &cobra.Command{
	Use:     "output",
	Short:   "JSON, HTML, SQL and sitemap generator",
	Aliases: []string{"o"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Usage()
			os.Exit(0)
		}
		_ = cmd.Usage()
		logs.Check(fmt.Errorf("output: invalid command %v please use one of the available fix commands", args[0]))
	},
}

func init() {
	rootCmd.AddCommand(outputCmd)
	outputCmd.AddCommand(dataCmd)
	dataCmd.Flags().BoolVarP(&df.CronJob, "cronjob", "j", false, "data backup for the cron time-based job scheduler\nall other flags are ignored")
	dataCmd.Flags().BoolVarP(&df.Compress, "compress", "c", false, fmt.Sprintf("save and compress the SQL using bzip2\n%s/d2-sql-create.bz2", viper.Get("directory.sql")))
	dataCmd.Flags().UintVarP(&df.Limit, "limit", "l", 1, "limit the number of rows returned (no limit 0)")
	dataCmd.Flags().BoolVarP(&df.Parallel, "parallel", "p", true, "run --table=all queries in parallel")
	dataCmd.Flags().BoolVarP(&df.Save, "save", "s", false, fmt.Sprintf("save the SQL\n%s/d2-sql-update.sql", viper.Get("directory.sql")))
	dataCmd.Flags().StringVarP(&df.Table, "table", "t", "files", fmt.Sprintf("database table to use\noptions: all,%s", database.Tbls))
	dataCmd.Flags().StringVarP(&df.Type, "type", "y", "update", "database export type\noptions: create or update")
	err := dataCmd.Flags().MarkHidden("parallel")
	logs.Check(err)
	outputCmd.AddCommand(groupCmd)
	groupCmd.Flags().StringVarP(&gf.filter, "filter", "f", "", "filter groups (default all)\noptions: "+groups.Filters)
	groupCmd.Flags().BoolVarP(&gf.counts, "count", "c", false, "display the file totals for each group (SLOW)")
	groupCmd.Flags().BoolVarP(&gf.progress, "progress", "p", true, "show a progress indicator while fetching a large number of records")
	groupCmd.Flags().BoolVarP(&gf.cronjob, "cronjob", "j", false, "run in cronjob automated mode, ignores all other arguments")
	groupCmd.Flags().StringVarP(&gf.format, "format", "t", "", "output format (default html)\noptions: datalist,html,text")
	groupCmd.Flags().BoolVarP(&gf.init, "initialism", "i", false, "display the acronyms and initialisms for groups (SLOW)")
	outputCmd.AddCommand(peopleCmd)
	peopleCmd.Flags().StringVarP(&pf.filter, "filter", "f", "", "filter groups (default all)\noptions: "+people.Roles)
	peopleCmd.Flags().StringVarP(&pf.format, "format", "t", "", "output format (default html)\noptions: datalist,html,text")
	outputCmd.AddCommand(recentCmd)
	recentCmd.Flags().BoolVarP(&rf.compress, "compress", "c", false, "remove insignificant whitespace characters")
	recentCmd.Flags().UintVarP(&rf.limit, "limit", "l", 15, "limit the number of rows returned")
	outputCmd.AddCommand(sitemapCmd)
}

var df database.Flags

var dataCmd = &cobra.Command{
	Use:     "data",
	Aliases: []string{"d", "sql"},
	Short:   "An SQL dump generator to export files",
	Run: func(cmd *cobra.Command, args []string) {
		df.Version = version
		switch {
		case df.CronJob:
			if err := df.ExportCronJob(); err != nil {
				log.Fatal(err)
			}
		case df.Table == "all":
			if err := df.ExportDB(); err != nil {
				log.Fatal(err)
			}
		default:
			if err := df.ExportTable(); err != nil {
				log.Fatal(err)
			}
		}
	},
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

// groupCmd represents the organisations command.
var groupCmd = &cobra.Command{
	Use:     "groups",
	Aliases: []string{"g", "group"},
	Short:   "A HTML snippet generator to list groups",
	Run: func(cmd *cobra.Command, args []string) {
		if gf.cronjob {
			if err := groups.Cronjob(); err != nil {
				log.Fatal(err)
			}
			return
		}
		filterFlag(groups.Wheres(), "filter", gf.filter)
		var req = groups.Request{Filter: gf.filter, Counts: gf.counts, Initialisms: gf.init, Progress: gf.progress}
		switch gf.format {
		case "datalist", "dl", "d":
			if err := req.DataList(""); err != nil {
				log.Fatal(err)
			}
		case "html", "h", "":
			if err := req.HTML(""); err != nil {
				log.Fatal(err)
			}
		case "text", "t":
			if _, err := groups.Print(req); err != nil {
				log.Fatal(err)
			}
		}
	},
}

type pplFlags struct {
	filter   string
	format   string
	progress bool
}

var pf pplFlags

// peopleCmd represents the authors command.
var peopleCmd = &cobra.Command{
	Use:     "people",
	Aliases: []string{"p", "ppl"},
	Short:   "A HTML snippet generator to list people",
	Run: func(cmd *cobra.Command, args []string) {
		filterFlag(people.Filters(), "filter", pf.filter)
		var req people.Request
		if filterFlag(fmtflags, "format", pf.format); pf.format != "" {
			req = people.Request{Filter: pf.filter, Progress: pf.progress}
		}
		switch pf.format {
		case "datalist", "dl", "d":
			if err := people.DataList("", req); err != nil {
				log.Fatal(err)
			}
		case "html", "h", "":
			if err := people.HTML("", req); err != nil {
				log.Fatal(err)
			}
		case "text", "t":
			if err := people.Print(req); err != nil {
				log.Fatal(err)
			}
		}
	},
}

type recentFlags struct {
	compress bool
	limit    uint
}

var rf recentFlags

var recentCmd = &cobra.Command{
	Use:     "recent",
	Aliases: []string{"r"},
	Short:   "A JSON snippet generator to list recent file additions",
	Run: func(cmd *cobra.Command, args []string) {
		if err := recent.List(rf.limit, rf.compress); err != nil {
			log.Fatal(err)
		}
	},
}

var sitemapCmd = &cobra.Command{
	Use:     "sitemap",
	Aliases: []string{"m", "s", "map"},
	Short:   "A site map generator",
	Run: func(cmd *cobra.Command, args []string) {
		if err := sitemap.Create(); err != nil {
			log.Fatal(err)
		}
	},
}
