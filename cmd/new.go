//nolint:gochecknoglobals
package cmd

import (
	"github.com/Defacto2/df2/cmd/internal/run"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new",
	Short:   "Manage files marked as waiting to go live (default).",
	Aliases: []string{"n"},
	GroupID: "group1",
	Long: `Runs a sequence of commands to handle the files waiting to go live.
This is the default df2 command when used without any flags or arguments.

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

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(newCmd)
}
