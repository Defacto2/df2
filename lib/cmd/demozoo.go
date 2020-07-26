package cmd

import (
	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/demozoo"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type dzFlags struct {
	all       bool // scan for all proofs, not just new submissions
	overwrite bool // overwrite all existing assets
	simulate  bool
	new       bool
	refresh   bool
	id        string // auto-generated id or a uuid
	extract   []string
	ping      uint
	download  uint
}

var dzoo dzFlags

// demozooCmd represents the demozoo command.
var demozooCmd = &cobra.Command{
	Use:     "demozoo",
	Short:   "Interact with Demozoo.org upload submissions",
	Aliases: []string{"d", "dz"},
	Example: `  df2 demozoo [--new|--all|--id] (--dry-run,--overwrite)
  df2 demozoo [--refresh|--ping|--download]`,
	Run: func(cmd *cobra.Command, args []string) {
		var empty []string
		var err error
		r := demozoo.Request{
			All:       dzoo.all,
			Overwrite: dzoo.overwrite,
			Refresh:   dzoo.refresh,
			Simulate:  dzoo.simulate,
		}
		switch {
		case dzoo.new, dzoo.all:
			err = r.Queries()
		case dzoo.id != "":
			err = r.Query(dzoo.id)
		case dzoo.refresh:
			err = demozoo.RefreshMeta()
		case dzoo.ping != 0:
			f, err := demozoo.Fetch(dzoo.ping)
			if err != nil {
				logs.Check(err)
			}
			logs.Printf("Demozoo ID %v, HTTP status %v\n", dzoo.ping, f.Status)
			f.API.Print()
		case dzoo.download != 0:
			f, err := demozoo.Fetch(dzoo.download)
			if err != nil {
				logs.Check(err)
			}
			logs.Printf("Demozoo ID %v, HTTP status %v\n", dzoo.download, f.Status)
			f.API.Downloads()
			logs.Print("\n")
		case len(dzoo.extract) == 1:
			id, err := uuid.NewRandom()
			logs.Check(err)
			d, err := archive.ExtractDemozoo(dzoo.extract[0], id.String(), &empty)
			logs.Check(err)
			if err == nil {
				logs.Println(d.String())
			}
		case len(dzoo.extract) > 1: // limit to the first 2 flags
			d, err := archive.ExtractDemozoo(dzoo.extract[0], dzoo.extract[1], &empty)
			logs.Check(err)
			if err == nil {
				logs.Println(d.String())
			}
		default:
			err = cmd.Usage()
		}
		logs.Check(err)
	},
}

func init() {
	var err error
	rootCmd.AddCommand(demozooCmd)
	demozooCmd.Flags().BoolVarP(&dzoo.new, "new", "n", false, "scan for new demozoo submissions (recommended)")
	demozooCmd.Flags().BoolVar(&dzoo.all, "all", false, "scan all files with demozoo links (SLOW)")
	demozooCmd.Flags().StringVarP(&dzoo.id, "id", "i", "", "file id or uuid with a demozoo link to scan\n")
	demozooCmd.Flags().BoolVarP(&dzoo.simulate, "dry-run", "d", false, "simulate the fixes and display the expected changes")
	demozooCmd.Flags().BoolVar(&dzoo.overwrite, "overwrite", false, "rescan archives and overwrite all existing assets\n")
	demozooCmd.Flags().BoolVarP(&dzoo.refresh, "refresh", "r", false, "replace missing files metadata with demozoo data (SLOW)")
	demozooCmd.Flags().UintVarP(&dzoo.ping, "ping", "p", 0, "fetch and display a production record from the Demozoo.org API")
	demozooCmd.Flags().UintVarP(&dzoo.download, "download", "g", 0, "fetch and download a production's link file via the Demozoo.org API\n")
	demozooCmd.Flags().StringArrayVar(&dzoo.extract, "extract", make([]string, 0), `extracts and parses an archived file
requires two flags: --extract [filename] --extract [uuid]`)
	err = demozooCmd.MarkFlagFilename("extract")
	logs.Check(err)
	err = demozooCmd.Flags().MarkHidden("extract")
	logs.Check(err)
	demozooCmd.Flags().SortFlags = false
}
