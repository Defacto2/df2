//nolint:gochecknoglobals,gochecknoinits
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Show environment variables used for configuration.",
	Long: "Display the operating system environmental variables" +
		" used as configuations for this tool.",
	Aliases: []string{"e"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stdout, confg)
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
}
