// nolint:gochecknoglobals
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
	"github.com/gookit/color" //nolint:misspell
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
		i := 1
		config.Check()
		color.Info.Printf("%d. scan for new demozoo submissions\n", i)
		dz := demozoo.Request{
			All:       false,
			Overwrite: false,
			Refresh:   false,
			Simulate:  false,
		}
		if err := dz.Queries(); err != nil {
			logs.Fatal(err)
		}
		i++
		color.Info.Printf("%d. scan for new proof submissions\n", i)
		p := proof.Request{
			Overwrite:   false,
			AllProofs:   false,
			HideMissing: false,
		}
		if err := p.Queries(); err != nil {
			logs.Fatal(err)
		}
		i++
		color.Info.Printf("%d. generate missing images\n", i)
		if err := images.Fix(false); err != nil {
			logs.Fatal(err)
		}
		i++
		color.Info.Printf("%d. generate missing text previews\n", i)
		if err := text.Fix(false); err != nil {
			logs.Fatal(err)
		}
		i++
		color.Info.Printf("%d. fix demozoo data conflicts\n", i)
		if err := demozoo.Fix(); err != nil {
			log.Fatal(err)
		}
		i++
		color.Info.Printf("%d. fix malformed database entries\n", i)
		if err := database.Fix(); err != nil {
			log.Fatal(err)
		}
		if err := groups.Fix(false); err != nil {
			log.Fatal(err)
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(newCmd)
}
