// nolint:gochecknoglobals
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Defacto2/df2/lib/cmd/internal/arg"
	"github.com/Defacto2/df2/lib/config"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
)

var conf arg.Config

// configCmd represents the config command.
var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Configure the settings for this tool.",
	Long:    `Configure settings and defaults for df2.`,
	Aliases: []string{"cfg"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Usage(); err != nil {
			log.Print(fmt.Errorf("config cmd usage: %w", err))
		}
		if len(args) != 0 || cmd.Flags().NFlag() != 0 {
			if err := logs.Arg("config", true, args...); err != nil {
				log.Print(err)
			}
		}
	},
}

var configCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new config file.",
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Create(conf.Overwrite); err != nil {
			log.Print(fmt.Errorf("config create: %w", err))
		}
	},
}

var configDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Remove the config file.",
	Aliases: []string{"d"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Delete(); err != nil {
			log.Print(fmt.Errorf("config delete: %w", err))
		}
	},
}

var configEditCmd = &cobra.Command{
	Use:     "edit",
	Short:   "Edit the config file.",
	Aliases: []string{"e"},
	Run: func(cmd *cobra.Command, args []string) {
		config.Edit()
	},
}

var configInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "View settings configured by the config.",
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Info(conf.InfoSize); err != nil {
			log.Print(fmt.Errorf("config info: %w", err))
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Change a configuration.",
	Long: `Change a configuration setting using the name flag. After requesting
a setting change you will be prompted for a new value which will be validated.
See the examples for usage syntax and also see the --name flag description for
the command to list the available seettings.`,
	Aliases: []string{"s"},
	Example: `--name connection.server.host # to change the database host setting
--name directory.000          # to set the image preview directory`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Set(conf.Name); err != nil {
			if errors.Is(err, config.ErrSetName) {
				os.Exit(1)
			}
			log.Print(fmt.Errorf("config set: %w", err))
		}
	},
}

func init() { // nolint:gochecknoinits
	database.Init()
	directories.Init(false)
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configCreateCmd)
	configCreateCmd.Flags().BoolVarP(&conf.Overwrite, "overwrite", "y", false,
		"overwrite any existing config file")
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configInfoCmd)
	configInfoCmd.Flags().BoolVarP(&conf.InfoSize, "size", "s", false,
		"display directory sizes and file counts (SLOW)")
	configCmd.AddCommand(configSetCmd)
	configSetCmd.Flags().StringVarP(&conf.Name, "name", "n", "",
		`the configuration path to edit in dot syntax (see examples)
	to see a list of names run: df2 config info`)
	if err := configSetCmd.MarkFlagRequired("name"); err != nil {
		log.Print(err)
	}
}
