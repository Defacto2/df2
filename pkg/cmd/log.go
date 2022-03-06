// nolint:gochecknoglobals
package cmd

import (
	"github.com/Defacto2/df2/pkg/cmd/internal/run"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:     "log",
	Short:   "Display the df2 error log",
	Aliases: []string{},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run.Log(); err != nil {
			logs.Fatal(err)
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(logCmd)
}
