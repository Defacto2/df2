//nolint:gochecknoglobals
package cmd

import (
	"os"
	"sync"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/shrink"
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
	GroupID: "group2",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.Connect(cfg)
		if err != nil {
			logr.Fatal(err)
		}
		defer db.Close()

		const delta = 3
		w := os.Stdout
		var wg sync.WaitGroup
		wg.Add(delta)
		go func() {
			if err := shrink.SQL(w, cfg.SQLDumps); err != nil {
				logr.Error(err)
			}
			wg.Done()
		}()
		go func() {
			if err := shrink.Files(db, w, cfg.IncomingFiles); err != nil {
				logr.Error(err)
			}
			wg.Done()
		}()
		go func() {
			if err := shrink.Previews(db, w, cfg.IncomingImgs); err != nil {
				logr.Error(err)
			}
			wg.Done()
		}()
		wg.Wait()
	},
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(shrinkCmd)
}
