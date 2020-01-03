package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const configName = ".df2.yaml"

var (
	panic      bool = false // debug log
	quiet      bool = false // quiet disables most printing or output to terminal
	configFile string
	home, _    = os.UserHomeDir()
	filepath   = path.Join(home, configName)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "df2",
	Short: "A tool to configure and manage defacto2.net",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", fmt.Sprintf("config file (default is $HOME/%s)", configName))
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suspend feedback to the terminal")
	rootCmd.PersistentFlags().BoolVar(&panic, "panic", false, "panic in the disco")
	err := rootCmd.PersistentFlags().MarkHidden("panic")
	logs.Check(err)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	initPanic(panic)
	initQuiet(quiet)
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath(home)
		viper.SetConfigName(configName)
	}
	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && !quiet {
		println("Using config file:", viper.ConfigFileUsed())
	}
}

// initPanic toggles panics for all logged errors.
func initPanic(q bool) {
	logs.Panic = q
}

// initQuiet quiets the terminal output.
func initQuiet(q bool) {
	logs.Quiet = q
}
