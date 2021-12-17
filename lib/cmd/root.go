// Package cmd handles the commandline user interface and interactions.
// nolint:gochecknoglobals
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Defacto2/df2/lib/cmd/internal/arg"
	"github.com/Defacto2/df2/lib/cmd/internal/run"
	"github.com/Defacto2/df2/lib/config"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ErrCmd  = errors.New("invalid command, please use one of the available commands")
	ErrNoID = errors.New("requires an id or uuid argument")
	ErrID   = errors.New("invalid id or uuid specified")
)

var gf arg.Execute

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "df2",
	Short: "The tool to optimise and manage defacto2.net",
	Long: fmt.Sprintf("%s\nCopyright © %v Ben Garrett\n%v",
		color.Info.Sprint("The tool to optimise and manage defacto2.net"),
		run.Copyright(),
		color.Primary.Sprint("https://github.com/Defacto2/df2")),
	Run: func(cmd *cobra.Command, args []string) {
		if err := run.New(); err != nil {
			logs.Fatal(err)
		}
	},
}

// Execute is a Cobra command that adds all child commands to the root and
// sets the appropriate flags. It is called by main.main() and only needs
// to be called once in the rootCmd.
func Execute() error {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(color.Warn.Sprintf("%s", err))
		if e := err.Error(); strings.Contains(e, "required flag(s) \"name\"") {
			logs.Println("see Examples for usage or run to list setting choices:",
				color.Bold.Sprintf("%s config info", rootCmd.CommandPath()))
		}
		return err
	}
	config.Check()
	return nil
}

func init() { // nolint:gochecknoinits
	cobra.OnInitialize()
	readIn()
	rootCmd.PersistentFlags().StringVar(&gf.Filename, "config", "",
		fmt.Sprintf("config file (default is %s)", config.Filepath()))
	rootCmd.PersistentFlags().BoolVarP(&gf.Quiet, "quiet", "q", false,
		"suspend feedback to the terminal")
	rootCmd.PersistentFlags().BoolVar(&gf.Panic, "panic", false,
		"panic in the disco")
	if err := rootCmd.PersistentFlags().MarkHidden("panic"); err != nil {
		logs.Fatal(err)
	}
}

// readIn the config file and ENV variables if set.
func readIn() {
	logs.Panic(gf.Panic)
	logs.Quiet(gf.Quiet)
	if cf := config.Filepath(); cf == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			logs.Fatal(err)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(config.Config.Name)
	} else {
		viper.SetConfigFile(cf)
	}
	viper.AutomaticEnv() // read in environment variables that match
	// if a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		config.Config.Errors = true
		return
	}
	if !gf.Quiet && !str.Piped() {
		logs.Println(str.Sec(fmt.Sprintf("config file in use: %s",
			viper.ConfigFileUsed())))
	}
}
