// nolint:gochecknoglobals
package cmd

import (
	"sync"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/shrink"
	"github.com/spf13/cobra"
)

// shrinkCmd represents the compact command.
var shrinkCmd = &cobra.Command{
	Use:   "shrink",
	Short: "Reduces the space used in directories.",
	Long: `Shrink reduces the hard drive space used for directories on
the website. This command will only work when no records in the database
are 'waiting for approval.'`,
	Aliases: []string{"s"},
	Run: func(cmd *cobra.Command, args []string) {
		const delta = 3
		var wg sync.WaitGroup
		wg.Add(delta)
		go func() {
			if err := shrink.SQL(); err != nil {
				logs.Danger(err)
			}
			wg.Done()
		}()
		go func() {
			if err := shrink.Files(); err != nil {
				logs.Danger(err)
			}
			wg.Done()
		}()
		go func() {
			if err := shrink.Previews(); err != nil {
				logs.Danger(err)
			}
			wg.Done()
		}()
		wg.Wait()
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(shrinkCmd)
}
