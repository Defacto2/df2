//nolint:gochecknoglobals
package cmd

import (
	"errors"
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

var ErrNoOutput = errors.New("no output command used")

const notUsed = "\n\nThis document is not in use on the website."

var (
	dbf database.Flags
	gro arg.Group
	peo arg.People
	rec arg.Recent
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
		logr.Errorf("%q subcommand for output is an %w", args[0], ErrCmd)
	},
}

func init() { //nolint:gochecknoinits
	const fifteen = 15
	rootCmd.AddCommand(outputCmd)
	outputCmd.AddCommand(dataCmd)
	dataCmd.Flags().BoolVarP(&dbf.CronJob, "cronjob", "j", false,
		"data backup for the cron time-based job scheduler\nall other flags are ignored")
	dataCmd.Flags().BoolVarP(&dbf.Compress, "compress", "c", false,
		fmt.Sprintf("save and compress the SQL using bzip2\n%s/d2-sql-create.bz2", cfg.SQLDumps))
	dataCmd.Flags().UintVarP(&dbf.Limit, "limit", "l", 1,
		"limit the number of rows returned (no limit 0)")
	dataCmd.Flags().BoolVarP(&dbf.Parallel, "parallel", "p", true,
		"run --table=all queries in parallel")
	dataCmd.Flags().BoolVarP(&dbf.Save, "save", "s", false,
		fmt.Sprintf("save the SQL\n%s/d2-sql-update.sql", cfg.SQLDumps))
	dataCmd.Flags().StringVarP(&dbf.Tables, "table", "t", "files",
		fmt.Sprintf("database table to use\noptions: all, %s", database.Tbls()))
	dataCmd.Flags().StringVarP(&dbf.Type, "type", "y", "update",
		"database export type\noptions: create or update")
	if err := dataCmd.Flags().MarkHidden("parallel"); err != nil {
		logr.Fatal(err)
	}
	outputCmd.AddCommand(groupCmd)
	groupCmd.Flags().StringVarP(&gro.Filter, "filter", "f", "",
		"filter groups (default all)\noptions: "+strings.Join(groups.Wheres(), ","))
	groupCmd.Flags().BoolVarP(&gro.Counts, "count", "c", false,
		"display the file totals for each group (SLOW)")
	groupCmd.Flags().BoolVarP(&gro.Progress, "progress", "p", true,
		"show a progress indicator while fetching a large number of records")
	groupCmd.Flags().BoolVarP(&gro.Cronjob, "cronjob", "j", false,
		"run in cronjob automated mode, ignores all other arguments")
	groupCmd.Flags().BoolVar(&gro.Forcejob, "cronjob-force", false,
		"force the running of the cronjob automated mode")
	groupCmd.Flags().StringVarP(&gro.Format, "format", "t", "",
		"output format (default html)\noptions: datalist,html,text")
	groupCmd.Flags().BoolVarP(&gro.Init, "initialism", "i", false,
		"display the acronyms and initialisms for groups (SLOW)")
	outputCmd.AddCommand(peopleCmd)
	peopleCmd.Flags().StringVarP(&peo.Filter, "filter", "f", "",
		"filter people (default all)\noptions: "+people.Roles())
	peopleCmd.Flags().StringVarP(&peo.Format, "format", "t", "",
		"output format (default html)\noptions: datalist,html,text")

	peopleCmd.Flags().BoolVarP(&peo.Cronjob, "cronjob", "j", false,
		"run in cronjob automated mode, ignores all other arguments")
	peopleCmd.Flags().BoolVar(&peo.Forcejob, "cronjob-force", false,
		"force the running of the cronjob automated mode")

	outputCmd.AddCommand(recentCmd)
	recentCmd.Flags().BoolVarP(&rec.Compress, "compress", "c", false,
		"remove insignificant whitespace characters")
	recentCmd.Flags().UintVarP(&rec.Limit, "limit", "l", fifteen,
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
		db, err := database.Connect(cfg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		dbf.SQLDumps = cfg.SQLDumps
		if err := run.Data(db, os.Stdout, dbf); err != nil {
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
		db, err := database.Connect(cfg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		if err := run.Groups(db, os.Stdout, cfg.HTMLExports, gro); err != nil {
			logr.Error(err)
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
the website stylization.` + notUsed,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(cfg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		if err := run.People(db, os.Stdout, cfg.HTMLExports, peo); err != nil {
			logr.Error(err)
		}
	},
}

var recentCmd = &cobra.Command{
	Use:     "recent",
	Aliases: []string{"r"},
	Short:   "JSON snippet generator to list recent additions.",
	Long:    `JSON snippet generator to list recent additions.` + notUsed,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(cfg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		if err := recent.List(db, rec.Limit, rec.Compress); err != nil {
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
		db, err := database.Connect(cfg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()
		if err := sitemap.Create(db, cfg.HTMLViews); err != nil {
			logr.Error(err)
		}
	},
}
