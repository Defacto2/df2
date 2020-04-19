package cmd

import (
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

var waitingCmd = &cobra.Command{
	Use:   "waiting",
	Short: "Handler for files flagged as waiting to go live",
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
			Simulate:  false}
		color.Info.Println("scan for new demozoo submissions")
		err = dz.Queries()
		logs.Check(err)
		// proofs handler
		p := proof.Request{
			Overwrite:   false,
			AllProofs:   false,
			HideMissing: false}
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
		database.Fix()
		groups.Fix(false)
	},
}

func init() {
	rootCmd.AddCommand(waitingCmd)
}
