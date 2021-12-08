// nolint:gochecknoglobals
package cmd

import (
	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/demozoo"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type demozooFlags struct {
	all       bool // scan for all proofs, not just new submissions
	overwrite bool // overwrite all existing assets
	simulate  bool
	new       bool
	refresh   bool
	sync      bool
	id        string // auto-generated id or a uuid
	extract   []string
	ping      uint
	download  uint
}

var dzf demozooFlags

// demozooCmd represents the demozoo command.
var demozooCmd = &cobra.Command{
	Use:     "demozoo",
	Short:   "Interact with Demozoo.org upload submissions",
	Aliases: []string{"d", "dz"},
	Example: `  df2 demozoo [--new|--all|--id] (--dry-run,--overwrite)
  df2 demozoo [--refresh|--ping|--download]`,
	Run: func(cmd *cobra.Command, args []string) {
		var empty []string
		r := demozoo.Request{
			All:       dzf.all,
			Overwrite: dzf.overwrite,
			Refresh:   dzf.refresh,
			Simulate:  dzf.simulate,
		}
		switch {
		case dzf.new, dzf.all:
			if err := r.Queries(); err != nil {
				logs.Fatal(err)
			}
		case dzf.id != "":
			if err := r.Query(dzf.id); err != nil {
				logs.Fatal(err)
			}
		case dzf.refresh:
			if err := demozoo.RefreshMeta(); err != nil {
				logs.Fatal(err)
			}
		case dzf.sync:
			if err := demozoo.Sync(); err != nil {
				logs.Fatal(err)
			}
		case dzf.ping != 0:
			f, err := demozoo.Fetch(dzf.ping)
			if err != nil {
				logs.Fatal(err)
			}
			if !str.Piped() {
				logs.Printf("Demozoo ID %v, HTTP status %v\n", dzf.ping, f.Status)
			}
			if err := f.API.Print(); err != nil {
				logs.Fatal(err)
			}
		case dzf.download != 0:
			f, err := demozoo.Fetch(dzf.download)
			if err != nil {
				logs.Fatal(err)
			}
			logs.Printf("Demozoo ID %v, HTTP status %v\n", dzf.download, f.Status)
			f.API.Downloads()
			logs.Print("\n")
		case len(dzf.extract) == 1:
			id, err := uuid.NewRandom()
			if err != nil {
				logs.Fatal(err)
			}
			d, err := archive.Demozoo(dzf.extract[0], id.String(), &empty)
			if err != nil {
				logs.Fatal(err)
			}
			logs.Println(d.String())
		case len(dzf.extract) > 1: // limit to the first 2 flags
			d, err := archive.Demozoo(dzf.extract[0], dzf.extract[1], &empty)
			if err != nil {
				logs.Fatal(err)
			}
			logs.Println(d.String())
		default:
			if err := cmd.Usage(); err != nil {
				logs.Fatal(err)
			}
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(demozooCmd)
	demozooCmd.Flags().BoolVarP(&dzf.new, "new", "n", false,
		"scan for new demozoo submissions (recommended)")
	demozooCmd.Flags().BoolVar(&dzf.all, "all", false,
		"scan all files with demozoo links (SLOW)")
	demozooCmd.Flags().StringVarP(&dzf.id, "id", "i", "",
		"file id or uuid with a demozoo link to scan\n")
	demozooCmd.Flags().BoolVarP(&dzf.simulate, "dry-run", "d", false,
		"simulate the fixes and display the expected changes")
	demozooCmd.Flags().BoolVar(&dzf.overwrite, "overwrite", false,
		"rescan archives and overwrite all existing assets\n")
	demozooCmd.Flags().BoolVarP(&dzf.refresh, "refresh", "r", false,
		"replace missing files metadata with demozoo data (SLOW)")
	demozooCmd.Flags().BoolVarP(&dzf.sync, "sync", "s", false,
		"scan the demozoo api for missing bbstros and cracktros (SLOW)")
	demozooCmd.Flags().UintVarP(&dzf.ping, "ping", "p", 0,
		"fetch and display a production record from the Demozoo.org API")
	demozooCmd.Flags().UintVarP(&dzf.download, "download", "g", 0,
		"fetch and download a production's link file via the Demozoo.org API\n")
	demozooCmd.Flags().StringArrayVar(&dzf.extract, "extract", make([]string, 0),
		`extracts and parses an archived file
requires two flags: --extract [filename] --extract [uuid]`)
	if err := demozooCmd.MarkFlagFilename("extract"); err != nil {
		logs.Fatal(err)
	}
	if err := demozooCmd.Flags().MarkHidden("extract"); err != nil {
		logs.Fatal(err)
	}
	demozooCmd.Flags().SortFlags = false
}
