// nolint:gochecknoglobals
package cmd

import (
	"fmt"
	"log"

	"github.com/Defacto2/df2/lib/config"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
)

type configFlags struct {
	name      string
	overwrite bool
}

var (
	cfgf     configFlags
	infoSize bool
)

// configCmd represents the config command.
var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Configure the settings for this tool",
	Aliases: []string{"cfg"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Usage(); err != nil {
			log.Fatal(fmt.Errorf("config cmd usage: %w", err))
		}
		if len(args) != 0 || cmd.Flags().NFlag() != 0 {
			if err := logs.Arg("config", args...); err != nil {
				log.Fatal(err)
			}
		}
	},
}

var configCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new config file",
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Create(cfgf.overwrite); err != nil {
			log.Fatal(fmt.Errorf("config create: %w", err))
		}
	},
}

var configDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Remove the config file",
	Aliases: []string{"d"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Delete(); err != nil {
			log.Fatal(fmt.Errorf("config delete: %w", err))
		}
	},
}

var configEditCmd = &cobra.Command{
	Use:     "edit",
	Short:   "Edit the config file",
	Aliases: []string{"e"},
	Run: func(cmd *cobra.Command, args []string) {
		config.Edit()
	},
}

var configInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "View settings configured by the config",
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Info(infoSize); err != nil {
			log.Fatal(fmt.Errorf("config info: %w", err))
		}
	},
}

const configSetLong = `Change a configuration setting using the name flag. After requesting
a setting change you will be prompted for a new value which will be validated.
See the examples for usage syntax and also see the --name flag description for
the command to list the available seettings.`

var configSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Change a configuration",
	Long:    configSetLong,
	Aliases: []string{"s"},
	Example: `--name connection.server.host # to change the database host setting
--name directory.000          # to set the image preview directory`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Set(cfgf.name); err != nil {
			log.Fatal(fmt.Errorf("config set: %w", err))
		}
	},
}

func init() { // nolint:gochecknoinits
	database.Init()
	directories.Init(false)
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configCreateCmd)
	configCreateCmd.Flags().BoolVarP(&cfgf.overwrite, "overwrite", "y", false, "overwrite any existing config file")
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configInfoCmd)
	configInfoCmd.Flags().BoolVarP(&infoSize, "size", "s", false, "display directory sizes and file counts (SLOW)")
	configCmd.AddCommand(configSetCmd)
	configSetCmd.Flags().StringVarP(&cfgf.name, "name", "n", "", `the configuration path to edit in dot syntax (see examples)
	to see a list of names run: df2 config info`)
	if err := configSetCmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}
}
