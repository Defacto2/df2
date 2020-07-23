package cmd

import (
	"log"

	"github.com/Defacto2/df2/lib/config"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/demozoo"
	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/images"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/proof"
	"github.com/Defacto2/df2/lib/text"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new",
	Short:   "Handler for files flagged as waiting to go live",
	Aliases: []string{"n"},
	Long: `Runs a sequence of commands to handle files waiting to go live.

  df2 demozoo --new
      proof
      fix images
      fix text
      fix demozoo
      fix database`,
	Run: func(cmd *cobra.Command, args []string) {
		config.ErrCheck()
		var err error
		// demozoo handler
		dz := demozoo.Request{
			All:       false,
			Overwrite: false,
			Refresh:   false,
			Simulate:  false,
		}
		color.Info.Println("scan for new demozoo submissions")
		err = dz.Queries()
		logs.Check(err)
		// proofs handler
		p := proof.Request{
			Overwrite:   false,
			AllProofs:   false,
			HideMissing: false,
		}
		color.Info.Println("scan for new proof submissions")
		err = p.Queries()
		logs.Check(err)
		// missing image previews
		color.Info.Println("generate missing images")
		err = images.Fix(false)
		logs.Check(err)
		// missing text file previews
		color.Info.Println("generate missing text previews")
		err = text.Fix(false)
		logs.Check(err)
		// fix database entries
		color.Info.Println("fix demozoo data conflicts")
		demozoo.Fix()
		color.Info.Println("fix malformed database entries")
		if err := database.Fix(); err != nil {
			log.Fatal(err)
		}
		if err := groups.Fix(false); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
