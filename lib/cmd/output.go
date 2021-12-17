// nolint:gochecknoglobals
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Defacto2/df2/lib/cmd/internal/arg"
	"github.com/Defacto2/df2/lib/cmd/internal/run"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/people"
	"github.com/Defacto2/df2/lib/recent"
	"github.com/Defacto2/df2/lib/sitemap"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ErrNoOutput = errors.New("no output command used")
)

const notUsed = "\n\nThis document is not in use on the website."

var (
	dbf database.Flags
	gpf arg.Group
	ppf arg.People
	rcf arg.Recent
)

// outputCmd represents the output command.
var outputCmd = &cobra.Command{
	Use:     "output",
	Short:   "Generators for JSON, HTML, SQL and sitemap documents.",
	Aliases: []string{"o"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			logs.Println(ErrNoOutput)
			logs.Println()
			if err := cmd.Usage(); err != nil {
				logs.Fatal(err)
			}
			os.Exit(0)
		}
		if err := cmd.Usage(); err != nil {
			logs.Fatal(err)
		}
		logs.Danger(fmt.Errorf("output cmd %q: %w", args[0], ErrCmd))
	},
}

func init() { // nolint:gochecknoinits
	const fifteen = 15
	rootCmd.AddCommand(outputCmd)
	outputCmd.AddCommand(dataCmd)
	dataCmd.Flags().BoolVarP(&dbf.CronJob, "cronjob", "j", false,
		"data backup for the cron time-based job scheduler\nall other flags are ignored")
	dataCmd.Flags().BoolVarP(&dbf.Compress, "compress", "c", false,
		fmt.Sprintf("save and compress the SQL using bzip2\n%s/d2-sql-create.bz2", viper.Get("directory.sql")))
	dataCmd.Flags().UintVarP(&dbf.Limit, "limit", "l", 1,
		"limit the number of rows returned (no limit 0)")
	dataCmd.Flags().BoolVarP(&dbf.Parallel, "parallel", "p", true,
		"run --table=all queries in parallel")
	dataCmd.Flags().BoolVarP(&dbf.Save, "save", "s", false,
		fmt.Sprintf("save the SQL\n%s/d2-sql-update.sql", viper.Get("directory.sql")))
	dataCmd.Flags().StringVarP(&dbf.Tables, "table", "t", "files",
		fmt.Sprintf("database table to use\noptions: all, %s", database.Tbls()))
	dataCmd.Flags().StringVarP(&dbf.Type, "type", "y", "update",
		"database export type\noptions: create or update")
	if err := dataCmd.Flags().MarkHidden("parallel"); err != nil {
		logs.Fatal(err)
	}
	outputCmd.AddCommand(groupCmd)
	groupCmd.Flags().StringVarP(&gpf.Filter, "filter", "f", "",
		"filter groups (default all)\noptions: "+strings.Join(groups.Wheres(), ","))
	groupCmd.Flags().BoolVarP(&gpf.Counts, "count", "c", false,
		"display the file totals for each group (SLOW)")
	groupCmd.Flags().BoolVarP(&gpf.Progress, "progress", "p", true,
		"show a progress indicator while fetching a large number of records")
	groupCmd.Flags().BoolVarP(&gpf.Cronjob, "cronjob", "j", false,
		"run in cronjob automated mode, ignores all other arguments")
	groupCmd.Flags().BoolVar(&gpf.Forcejob, "forcejob", false,
		"force the running of the cronjob automated mode")
	groupCmd.Flags().StringVarP(&gpf.Format, "format", "t", "",
		"output format (default html)\noptions: datalist,html,text")
	groupCmd.Flags().BoolVarP(&gpf.Init, "initialism", "i", false,
		"display the acronyms and initialisms for groups (SLOW)")
	outputCmd.AddCommand(peopleCmd)
	peopleCmd.Flags().StringVarP(&ppf.Filter, "filter", "f", "",
		"filter groups (default all)\noptions: "+people.Roles())
	peopleCmd.Flags().StringVarP(&ppf.Format, "format", "t", "",
		"output format (default html)\noptions: datalist,html,text")
	outputCmd.AddCommand(recentCmd)
	recentCmd.Flags().BoolVarP(&rcf.Compress, "compress", "c", false,
		"remove insignificant whitespace characters")
	recentCmd.Flags().UintVarP(&rcf.Limit, "limit", "l", fifteen,
		"limit the number of rows returned")
	outputCmd.AddCommand(sitemapCmd)
}

var dataCmd = &cobra.Command{
	Use:     "data",
	Aliases: []string{"d", "sql"},
	Short:   "Generate SQL data dump export files.",
	Long: `Generate a logical backup of the MySQL database. It produces
SQL statements that can be used to recreate the database objects and data.
The dumps can be used with mysqldump or Adminer to manage content in the
MySQL databases. `,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run.Data(dbf); err != nil {
			log.Fatal(err)
		}
	},
}

// groupCmd represents the organisations command.
var groupCmd = &cobra.Command{
	Use:     "groups",
	Aliases: []string{"g", "group"},
	Short:   "HTML snippet generator to list groups.",
	Long: `A HTML snippet generate to list groups. Each group will be wrapped with
a heading 2 element containing a relative anchor link to the group's page
and the name of the group.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run.Groups(gpf); err != nil {
			log.Fatal(err)
		}
	},
}

// peopleCmd represents the authors command.
var peopleCmd = &cobra.Command{
	Use:     "people",
	Aliases: []string{"p", "ppl"},
	Short:   "HTML snippet generator to list people.",
	Long: `A HTML snippet generate to list people. Each person will be wrapped with
a heading 2 element containing a relative anchor link to the person's page
and their name.` + notUsed,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run.People(ppf); err != nil {
			log.Fatal(err)
		}
	},
}

var recentCmd = &cobra.Command{
	Use:     "recent",
	Aliases: []string{"r"},
	Short:   "JSON snippet generator to list recent additions.",
	Long:    `JSON snippet generator to list recent additions.` + notUsed,
	Run: func(cmd *cobra.Command, args []string) {
		if err := recent.List(rcf.Limit, rcf.Compress); err != nil {
			log.Fatal(err)
		}
	},
}

var sitemapCmd = &cobra.Command{
	Use:     "sitemap",
	Aliases: []string{"m", "s", "map"},
	Short:   "Sitemap generator.",
	Long: `A sitemap generator to help search engines index the website.

"A sitemap is a file where you provide information about the pages, videos,
and other files on your site,  and the relationships  between them.  Search
engines like  Google read  this file  to crawl  your site more efficiently.
A sitemap  tells Google  which pages and  files you think are  important in
your site,  and also  provides  valuable  information  about these  files."

"If your site's pages are properly linked, Google can usually discover most
of your site.  Proper linking means that all pages that you deem  important
can be reached through some form of navigation."

See: https://developers.google.com/search/docs/advanced/sitemaps/overview`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := sitemap.Create(); err != nil {
			log.Fatal(err)
		}
	},
}
