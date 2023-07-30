//nolint:gochecknoglobals,gochecknoinits
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/cmd/internal/run"
	"github.com/spf13/cobra"
)

var envs arg.Env

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
		if envs.Init {
			if err := run.Env(os.Stdout, logr, confg); err != nil {
				logr.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
	if !confg.IsProduction {
		envCmd.Flags().BoolVarP(&envs.Init, "init", "i", false,
			"a developer-mode flag to create the directories within the environment configuration")
	}
}
