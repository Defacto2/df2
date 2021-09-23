// nolint:gochecknoglobals
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
	"github.com/Defacto2/df2/lib/zipcontent"
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
		if err := runNew(); err != nil {
			logs.Fatal(err)
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(newCmd)
}

func runNew() error {
	i := 0
	color.Primary.Println("Scans for new submissions and record cleanup")
	config.Check()
	i++
	color.Info.Printf("%d. scan for new demozoo submissions\n", i)
	newDZ := demozoo.Request{
		All:       false,
		Overwrite: false,
		Refresh:   false,
		Simulate:  false,
	}
	if err := newDZ.Queries(); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. scan for new proof submissions\n", i)
	newProof := proof.Request{
		Overwrite:   false,
		AllProofs:   false,
		HideMissing: false,
	}
	if err := newProof.Queries(); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. scan for empty archives\n", i)
	if err := zipcontent.Fix(); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. generate missing images\n", i)
	if err := images.Fix(false); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. generate missing text previews\n", i)
	if err := text.Fix(false); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. fix demozoo data conflicts\n", i)
	if err := demozoo.Fix(); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. fix malformed database entries\n", i)
	if err := database.Fix(); err != nil {
		return err
	}
	if err := groups.Fix(false); err != nil {
		return err
	}
	return nil
}
