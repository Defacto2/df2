package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/Defacto2/uuid/v2/lib/logs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cfgFilename = ".df2.yaml"
)

var (
	quiet    bool = false // quiet disables most printing or output to terminal
	cfgFile  string
	home, _  = os.UserHomeDir()
	filepath = path.Join(home, cfgFilename)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "uuid",
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/%s)", cfgFilename))
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suspend feedback to the terminal")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	initQuiet(quiet)
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(home)
		viper.SetConfigName(cfgFilename)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && !quiet {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// initQuiet quiets the terminal output
func initQuiet(q bool) {
	logs.Quiet = q
}
