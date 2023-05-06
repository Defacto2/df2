//nolint:gochecknoglobals,gochecknoinits
package cmd

import (
	"encoding/json"
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
		b, err := json.MarshalIndent(confg, "", " ")
		if err != nil {
			logr.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "%s\n", b)
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
}
