//nolint:gochecknoglobals
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/config"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/spf13/cobra"
)

var (
	conf arg.Config
	csf  config.SetFlags
)

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

var configSetConnCmd = &cobra.Command{
	Use:   "setdb",
	Short: "Set a database connection configuration using flags.",
	Long: `Setup a database connection configuration using flag values.
This command is intended for scripts and Docker containers.`,
	Aliases: []string{"db"},
	GroupID: "group1",
	Run: func(cmd *cobra.Command, args []string) {
		if err := csf.Set(); err != nil {
			log.Print(fmt.Errorf("config setdb: %w", err))
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
	GroupID: "group1",
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
	GroupID: "group1",
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

func init() { //nolint:gochecknoinits
	database.Init()
	directories.Init(false)
	rootCmd.AddCommand(configCmd)
	configCmd.AddGroup(&cobra.Group{ID: "group1", Title: "Settings:"})
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
	configCmd.AddCommand(configSetConnCmd)
	configSetConnCmd.Flags().StringVarP(&csf.Host, "host", "t", "",
		"Set a new MySQL hostname, it defines the location of your MySQL server (localhost)")
	configSetConnCmd.Flags().StringVarP(&csf.Protocol, "protocol", "l", "",
		"Set a new protocol to connect to the MySQL server (tcp|udp)")
	configSetConnCmd.Flags().IntVarP(&csf.Port, "port", "p", -1, "Set a new MySQL port")
}
