//nolint:gochecknoglobals,gochecknoinits
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/cmd/internal/run"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/Defacto2/df2/pkg/people"
	"github.com/Defacto2/df2/pkg/recent"
	"github.com/Defacto2/df2/pkg/sitemap"
	"github.com/spf13/cobra"
)

const unused = "\n\nThis document is not in use on the website."

var (
	dbase database.Flags
	group arg.Group
	peopl arg.People
	recnt arg.Recent
)

// outputCmd represents the output command.
var outputCmd = &cobra.Command{
	Use:     "output",
	Short:   "Generators for JSON, HTML, SQL and sitemap documents.",
	Aliases: []string{"o"},
	GroupID: "group1",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Fprintf(os.Stdout, "%s\n\n", ErrNoOutput)
			if err := cmd.Usage(); err != nil {
				logr.Fatal(err)
			}
			return
		}
		if err := cmd.Usage(); err != nil {
			logr.Fatal(err)
		}
		logr.Errorf("%q subcommand for output is an %w", args[0], ErrCommand)
	},
}

func init() {
	const fifteen = 15
	rootCmd.AddCommand(outputCmd)
	outputCmd.AddCommand(dataCmd)
	dataCmd.Flags().BoolVarP(&dbase.CronJob, "cronjob", "j", false,
		"data backup for the cron time-based job scheduler\nall other flags are ignored")
	dataCmd.Flags().BoolVarP(&dbase.Compress, "compress", "c", false,
		fmt.Sprintf("save and compress the SQL using bzip2\n%s/d2-sql-create.bz2", confg.SQLDumps))
	dataCmd.Flags().UintVarP(&dbase.Limit, "limit", "l", 1,
		"limit the number of rows returned (no limit 0)")
	dataCmd.Flags().BoolVarP(&dbase.Parallel, "parallel", "p", true,
		"run --table=all queries in parallel")
	dataCmd.Flags().BoolVarP(&dbase.Save, "save", "s", false,
		fmt.Sprintf("save the SQL\n%s/d2-sql-update.sql", confg.SQLDumps))
	dataCmd.Flags().StringVarP(&dbase.Tables, "table", "t", "files",
		fmt.Sprintf("database table to use\noptions: all, %s", database.Tbls()))
	dataCmd.Flags().StringVarP(&dbase.Type, "type", "y", "update",
		"database export type\noptions: create or update")
	if err := dataCmd.Flags().MarkHidden("parallel"); err != nil {
		logr.Fatal(err)
	}
	outputCmd.AddCommand(groupCmd)
	groupCmd.Flags().StringVarP(&group.Filter, "filter", "f", "",
		"filter groups (default all)\noptions: "+strings.Join(groups.Tags(), ","))
	groupCmd.Flags().BoolVarP(&group.Counts, "count", "c", false,
		"display the file totals for each group (SLOW)")
	groupCmd.Flags().BoolVarP(&group.Progress, "progress", "p", true,
		"show a progress indicator while fetching a large number of records")
	groupCmd.Flags().BoolVarP(&group.Cronjob, "cronjob", "j", false,
		"run in cronjob automated mode, ignores all other arguments")
	groupCmd.Flags().BoolVar(&group.Forcejob, "cronjob-force", false,
		"force the running of the cronjob automated mode")
	groupCmd.Flags().StringVarP(&group.Format, "format", "t", "",
		"output format (default html)\noptions: datalist,html,text")
	groupCmd.Flags().BoolVarP(&group.Init, "initialism", "i", false,
		"display the acronyms and initialisms for groups (SLOW)")
	outputCmd.AddCommand(peopleCmd)
	peopleCmd.Flags().StringVarP(&peopl.Filter, "filter", "f", "",
		"filter people (default all)\noptions: "+people.Roles())
	peopleCmd.Flags().StringVarP(&peopl.Format, "format", "t", "",
		"output format (default html)\noptions: datalist,html,text")

	peopleCmd.Flags().BoolVarP(&peopl.Cronjob, "cronjob", "j", false,
		"run in cronjob automated mode, ignores all other arguments")
	peopleCmd.Flags().BoolVar(&peopl.Forcejob, "cronjob-force", false,
		"force the running of the cronjob automated mode")

	outputCmd.AddCommand(recentCmd)
	recentCmd.Flags().BoolVarP(&recnt.Compress, "compress", "c", false,
		"remove insignificant whitespace characters")
	recentCmd.Flags().UintVarP(&recnt.Limit, "limit", "l", fifteen,
		"limit the number of rows returned")
	outputCmd.AddCommand(sitemapCmd)
}

var dataCmd = &cobra.Command{
	Use:     "data",
	Aliases: []string{"d", "sql"},
	Short:   "Generate SQL data dump export files.",
	Long: `Generate a logical backup of the MySQL database. It produces
	SQL statements that can recreate the database objects and data. These can be
	used with mysqldump or Adminer to manage content in the MySQL databases.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(confg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		dbase.SQLDumps = confg.SQLDumps
		if err := run.Data(db, os.Stdout, dbase); err != nil {
			logr.Error(err)
		}
	},
}

// groupCmd represents the organisations command.
var groupCmd = &cobra.Command{
	Use:     "groups",
	Aliases: []string{"g", "group"},
	Short:   "HTML snippet generator to list groups.",
	Long: `An HTML snippet generator to list groups. Each group is wrapped with a
heading-2 element containing a relative anchor link to the group's page and name.

The HTML output returned by the cronjob flag includes additional elements for
the website stylization.`,
	Run: func(cmd *cobra.Command, args []string) {
		w := os.Stdout
		if err := arg.FilterFlag(w, groups.Tags(), "filter", group.Filter); err != nil {
			os.Exit(1)
		}
		db, err := database.Connect(confg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		switch {
		case group.Cronjob, group.Forcejob:
			if err := run.GroupCron(db, w, confg, group); err != nil {
				logr.Error(err)
			}
		default:
			if err := run.Groups(db, w, w, group); err != nil {
				logr.Error(err)
			}
		}
	},
}

// peopleCmd represents the authors command.
var peopleCmd = &cobra.Command{
	Use:     "people",
	Aliases: []string{"p", "ppl"},
	Short:   "HTML snippet generator to list people.",
	Long: `An HTML snippet generator to list people. Each person is wrapped with a
heading-2 element containing a relative anchor link to the person's page and name.

The HTML output returned by the cronjob flag includes additional elements for
the website stylization.` + unused,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(confg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		if err := run.People(db, os.Stdout, confg.HTMLExports, peopl); err != nil {
			logr.Error(err)
		}
	},
}

var recentCmd = &cobra.Command{
	Use:     "recent",
	Aliases: []string{"r"},
	Short:   "JSON snippet generator to list recent additions.",
	Long:    `JSON snippet generator to list recent additions.` + unused,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(confg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		if err := recent.List(db, os.Stdout, recnt.Limit, recnt.Compress); err != nil {
			logr.Error(err)
		}
	},
}

var sitemapCmd = &cobra.Command{
	Use:     "sitemap",
	Aliases: []string{"m", "s", "map"},
	Short:   "Sitemap generator.",
	Long: `A sitemap generator to help search engines index the website.

"A sitemap is a file where you provide information about the pages,
videos, and other files on your site and the relationships between them.
Search engines like Google use this file to help crawl the site more
efficiently. A sitemap tells Google which pages and files you think are
essential to the site and provides valuable information about these
files. If the site's pages are correctly linked, Google can usually
discover most of the site."

See: https://developers.google.com/search/docs/advanced/sitemaps/overview`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(confg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		if err := sitemap.Create(db, os.Stdout, confg.HTMLViews); err != nil {
			logr.Error(err)
		}
	},
}
