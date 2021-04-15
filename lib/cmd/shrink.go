package cmd

import (
	"sync"

	"github.com/Defacto2/df2/lib/shrink"
	"github.com/spf13/cobra"
)

// shrinkCmd represents the compact command.
var shrinkCmd = &cobra.Command{
	Use:     "shrink",
	Short:   "Reduces the space used in directories",
	Aliases: []string{"s"},
	Run: func(cmd *cobra.Command, args []string) {
		const delta = 3
		var wg sync.WaitGroup
		wg.Add(delta)
		go func() {
			shrink.SQL()
			wg.Done()
		}()
		go func() {
			shrink.Files()
			wg.Done()
		}()
		go func() {
			shrink.Previews()
			wg.Done()
		}()
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(shrinkCmd)
}
