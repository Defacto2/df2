package cmd

import (
	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/demozoo"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type dzFlags struct {
	all       bool   // scan for all proofs, not just new submissions
	id        string // auto-generated id or a uuid
	overwrite bool   // overwrite all existing assets
	ping      uint
	download  uint
	//extract   string // struct?
	simulate bool
	new      bool
	extract  []string //map[string]string
}

var dzoo dzFlags

// demozooCmd represents the demozoo command
var demozooCmd = &cobra.Command{
	Use:   "demozoo",
	Short: "Interact with Demozoo.org upload submissions",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		r := demozoo.Request{
			All:       dzoo.all,
			Overwrite: dzoo.overwrite,
			Simulate:  dzoo.simulate}
		demozoo.Verbose = true
		switch {
		case dzoo.id != "":
			err = r.Query(dzoo.id)
		case dzoo.download != 0:
			_, s, data := demozoo.Fetch(dzoo.download)
			logs.Printf("Demozoo ID %v, HTTP status %v\n", dzoo.download, s)
			data.Downloads()
		case dzoo.ping != 0:
			_, s, data := demozoo.Fetch(dzoo.ping)
			logs.Printf("Demozoo ID %v, HTTP status %v\n", dzoo.ping, s)
			data.Print()
		case dzoo.new:
			err = r.Queries()
		case len(dzoo.extract) == 1:
			id, err := uuid.NewRandom()
			logs.Check(err)
			d, err := archive.ExtractDemozoo(dzoo.extract[0], id.String(), []string{})
			logs.Check(err)
			if err == nil {
				println(d.String())
			}
		case len(dzoo.extract) > 1: // only use the first 2 flags
			d, err := archive.ExtractDemozoo(dzoo.extract[0], dzoo.extract[1], []string{})
			logs.Check(err)
			if err == nil {
				println(d.String())
			}
		default:
			err = cmd.Usage()
		}
		logs.Check(err)
	},
}

func init() {
	rootCmd.AddCommand(demozooCmd)
	demozooCmd.Flags().BoolVarP(&dzoo.new, "new", "n", false, "scan for new demozoo submissions (recommended)")
	demozooCmd.Flags().StringVarP(&dzoo.id, "id", "i", "", "id or uuid to handle only one demozoo entry")
	demozooCmd.Flags().BoolVar(&dzoo.all, "all", false, "scan for all demozoo entries, not just new submissions")
	demozooCmd.Flags().BoolVarP(&dzoo.simulate, "simulate", "s", false, "simulate the fixes and display the expected changes")
	demozooCmd.Flags().BoolVar(&dzoo.overwrite, "overwrite", false, "rescan archives and overwrite all existing assets\n")
	demozooCmd.Flags().UintVarP(&dzoo.ping, "ping", "p", 0, "fetch and display a production record from the Demozoo.org API")
	demozooCmd.Flags().UintVarP(&dzoo.download, "download", "d", 0, "fetch and download a production's link file via the Demozoo.org API\n")
	demozooCmd.Flags().StringArrayVar(&dzoo.extract, "extract", make([]string, 0), `extracts and parses an archived file
requires two flags: --extract [filename] --extract [uuid]`)
	demozooCmd.MarkFlagFilename("extract")
	err := demozooCmd.Flags().MarkHidden("extract")
	logs.Check(err)
	demozooCmd.Flags().SortFlags = false
}
