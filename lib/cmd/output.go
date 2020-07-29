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

type groupFlags struct {
	counts   bool
	cronjob  bool
	init     bool
	progress bool
	filter   string
	format   string
}

type recentFlags struct {
	compress bool
	limit    uint
}

var (
	dbf database.Flags
	gpf groupFlags
	rcf recentFlags
)

var fmtflags = [7]string{"datalist", "html", "text", "dl", "d", "h", "t"}

// outputCmd represents the output command.
var outputCmd = &cobra.Command{
	Use:     "output",
	Short:   "JSON, HTML, SQL and sitemap generator",
	Aliases: []string{"o"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Usage(); err != nil {
				logs.Fatal(err)
			}
			os.Exit(0)
		}
		if err := cmd.Usage(); err != nil {
			logs.Fatal(err)
		}
		logs.Danger(fmt.Errorf("ouptut cmd %q: %w", args[0], ErrCmd))
	},
}

func init() {
	rootCmd.AddCommand(outputCmd)
	outputCmd.AddCommand(dataCmd)
	dataCmd.Flags().BoolVarP(&dbf.CronJob, "cronjob", "j", false, "data backup for the cron time-based job scheduler\nall other flags are ignored")
	dataCmd.Flags().BoolVarP(&dbf.Compress, "compress", "c", false, fmt.Sprintf("save and compress the SQL using bzip2\n%s/d2-sql-create.bz2", viper.Get("directory.sql")))
	dataCmd.Flags().UintVarP(&dbf.Limit, "limit", "l", 1, "limit the number of rows returned (no limit 0)")
	dataCmd.Flags().BoolVarP(&dbf.Parallel, "parallel", "p", true, "run --table=all queries in parallel")
	dataCmd.Flags().BoolVarP(&dbf.Save, "save", "s", false, fmt.Sprintf("save the SQL\n%s/d2-sql-update.sql", viper.Get("directory.sql")))
	dataCmd.Flags().StringVarP(&dbf.Tables, "table", "t", "files", fmt.Sprintf("database table to use\noptions: all,%s", database.Tbls))
	dataCmd.Flags().StringVarP(&dbf.Type, "type", "y", "update", "database export type\noptions: create or update")
	if err := dataCmd.Flags().MarkHidden("parallel"); err != nil {
		logs.Fatal(err)
	}
	outputCmd.AddCommand(groupCmd)
	groupCmd.Flags().StringVarP(&gpf.filter, "filter", "f", "", "filter groups (default all)\noptions: "+groups.Filters)
	groupCmd.Flags().BoolVarP(&gpf.counts, "count", "c", false, "display the file totals for each group (SLOW)")
	groupCmd.Flags().BoolVarP(&gpf.progress, "progress", "p", true, "show a progress indicator while fetching a large number of records")
	groupCmd.Flags().BoolVarP(&gpf.cronjob, "cronjob", "j", false, "run in cronjob automated mode, ignores all other arguments")
	groupCmd.Flags().StringVarP(&gpf.format, "format", "t", "", "output format (default html)\noptions: datalist,html,text")
	groupCmd.Flags().BoolVarP(&gpf.init, "initialism", "i", false, "display the acronyms and initialisms for groups (SLOW)")
	outputCmd.AddCommand(peopleCmd)
	peopleCmd.Flags().StringVarP(&pf.filter, "filter", "f", "", "filter groups (default all)\noptions: "+people.Roles)
	peopleCmd.Flags().StringVarP(&pf.format, "format", "t", "", "output format (default html)\noptions: datalist,html,text")
	outputCmd.AddCommand(recentCmd)
	recentCmd.Flags().BoolVarP(&rcf.compress, "compress", "c", false, "remove insignificant whitespace characters")
	recentCmd.Flags().UintVarP(&rcf.limit, "limit", "l", 15, "limit the number of rows returned")
	outputCmd.AddCommand(sitemapCmd)
}

var dataCmd = &cobra.Command{
	Use:     "data",
	Aliases: []string{"d", "sql"},
	Short:   "An SQL dump generator to export files",
	Run: func(cmd *cobra.Command, args []string) {
		dbf.Version = version
		switch {
		case dbf.CronJob:
			if err := dbf.ExportCronJob(); err != nil {
				log.Fatal(err)
			}
		case dbf.Tables == "all":
			if err := dbf.ExportDB(); err != nil {
				log.Fatal(err)
			}
		default:
			if err := dbf.ExportTable(); err != nil {
				log.Fatal(err)
			}
		}
	},
}

// groupCmd represents the organisations command.
var groupCmd = &cobra.Command{
	Use:     "groups",
	Aliases: []string{"g", "group"},
	Short:   "A HTML snippet generator to list groups",
	Run: func(cmd *cobra.Command, args []string) {
		if gpf.cronjob {
			if err := groups.Cronjob(); err != nil {
				log.Fatal(err)
			}
			return
		}
		filterFlag(groups.Wheres(), "filter", gpf.filter)
		req := groups.Request{Filter: gpf.filter, Counts: gpf.counts, Initialisms: gpf.init, Progress: gpf.progress}
		switch gpf.format {
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

var recentCmd = &cobra.Command{
	Use:     "recent",
	Aliases: []string{"r"},
	Short:   "A JSON snippet generator to list recent file additions",
	Run: func(cmd *cobra.Command, args []string) {
		if err := recent.List(rcf.limit, rcf.compress); err != nil {
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
