// nolint:gochecknoglobals
package cmd

import (
	"github.com/Defacto2/df2/lib/cmd/internal/run"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new",
	Short:   "Handler for files flagged as waiting to go live (default)",
	Aliases: []string{"n"},
	Long: `Runs a sequence of commands to handle files waiting to go live.

  df2 demozoo --new
      proof
      fix images
      fix text
      fix demozoo
      fix database`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run.New(); err != nil {
			logs.Fatal(err)
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(newCmd)
}
