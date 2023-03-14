//nolint:gochecknoglobals
package cmd

import (
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:     "log",
	Short:   "Display the df2 error log",
	Aliases: []string{},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: make this print zapper's log
		// if err := run.Log(os.Stdout, log); err != nil {
		// 	log.Fatal(err)
		// }
	},
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(logCmd)
}
