//nolint:gochecknoglobals
package cmd

import (
	"os"

	"github.com/Defacto2/df2/cmd/internal/run"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:     "log",
	Short:   "Display the df2 error log",
	Aliases: []string{},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run.Log(os.Stdout, log); err != nil {
			log.Fatal(err)
		}
	},
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(logCmd)
}
