package cmd

// os.Exit() = 200+

import (
	"github.com/Defacto2/df2/lib/config"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgOWFlag bool
var cfgNameFlag string

//var cfg config.

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure the settings for this tool",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Usage()
		logs.Check(err)
		if len(args) != 0 || cmd.Flags().NFlag() != 0 {
			logs.Arg("config", args)
		}
	},
}

var configCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new config file",
	Run: func(cmd *cobra.Command, args []string) {
		config.Create(cfgOWFlag)
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove the config file",
	Run: func(cmd *cobra.Command, args []string) {
		config.Delete()
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit the config file",
	Run: func(cmd *cobra.Command, args []string) {
		config.Edit()
	},
}

var configInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "View settings configured by the config",
	Run: func(cmd *cobra.Command, args []string) {
		config.Info()
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Change a configuration",
	//todo add long with information on how to view settings
	Example: `--name connection.server.host # to change the database host setting
--name directory.000          # to set the image preview directory`,
	Run: func(cmd *cobra.Command, args []string) {
		config.Set(cfgNameFlag)
	},
}

// InitDefaults initialises flag and configuration defaults.
func InitDefaults() {
	viper.SetDefault("connection.name", "defacto2-inno")
	viper.SetDefault("connection.user", "root")
	viper.SetDefault("connection.password", "password")

	viper.SetDefault("connection.server.protocol", "tcp")
	viper.SetDefault("connection.server.host", "localhost")
	viper.SetDefault("connection.server.port", "3306")

	viper.SetDefault("directory.root", "/opt/assets")
	viper.SetDefault("directory.backup", "/opt/assets/backups")
	viper.SetDefault("directory.emu", "/opt/assets/emularity.zip")
	viper.SetDefault("directory.uuid", "/opt/assets/downloads")
	viper.SetDefault("directory.000", "/opt/assets/000")
	viper.SetDefault("directory.150", "/opt/assets/150")
	viper.SetDefault("directory.400", "/opt/assets/400")
	viper.SetDefault("directory.html", "/opt/assets/html")
	viper.SetDefault("directory.views", "/opt/assets/views")
	viper.SetDefault("directory.incoming.files", "/opt/incoming/files")
	viper.SetDefault("directory.incoming.previews", "/opt/incoming/previews")
}

func init() {
	InitDefaults()
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
