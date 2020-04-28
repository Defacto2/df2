package cmd

// os.Exit() = 200+

import (
	"github.com/Defacto2/df2/lib/config"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
)

var cfgOWFlag bool
var cfgNameFlag string

//var cfg config.

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Configure the settings for this tool",
	Aliases: []string{"cfg"},
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Usage()
		logs.Check(err)
		if len(args) != 0 || cmd.Flags().NFlag() != 0 {
			logs.Arg("config", args)
		}
	},
}

var configCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new config file",
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {
		config.Create(cfgOWFlag)
	},
}

var configDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Remove the config file",
	Aliases: []string{"d"},
	Run: func(cmd *cobra.Command, args []string) {
		config.Delete()
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
		config.Info()
	},
}

var configSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Change a configuration",
	Aliases: []string{"s"},
	//TODO: add long with information on how to view settings
	Example: `--name connection.server.host # to change the database host setting
--name directory.000          # to set the image preview directory`,
	Run: func(cmd *cobra.Command, args []string) {
		config.Set(cfgNameFlag)
	},
}

func init() {
	database.Init()
	directories.Init(false)

	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configCreateCmd)
	configCreateCmd.Flags().BoolVarP(&cfgOWFlag, "overwrite", "y", false, "overwrite any existing config file")
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configInfoCmd)
	configCmd.AddCommand(configSetCmd)
	configSetCmd.Flags().StringVarP(&cfgNameFlag, "name", "n", "", `the configuration path to edit in dot syntax (see examples)
	to see a list of names run: df2 config info`)
	_ = configSetCmd.MarkFlagRequired("name")
}
