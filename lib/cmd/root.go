package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
Useful cobra funcs
rootCmd.CommandPath() || rootCmd.Use || rootCmd.Name() // df2
rootCmd.ResetCommands()
rootCmd.ResetFlags()
rootCmd.SilenceErrors()
rootCmd.SilenceUsage()
*/

const configName = ".df2.yaml"

var (
	panic      bool = false // debug log
	quiet      bool = false // quiet disables most printing or output to terminal
	configFile string
	home, _    = os.UserHomeDir()
	filepath   = path.Join(home, configName)
	fmtflags   = []string{"html", "text", "h", "t"}
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
		fmt.Printf("%s %s\n", logs.X(), err)
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
		logs.Sec(fmt.Sprintf("config file in use: %s", viper.ConfigFileUsed()))
	} else if e := fmt.Sprintf("%s", err); strings.Contains(e, "\""+configName+"\" Not Found in") {

		logs.Warn(fmt.Sprintf("no config file in use, please run: %s config create\n", rootCmd.CommandPath()))
	} else {
		println(fmt.Sprintf("%s", err))
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

// filterFlag compairs the value of the filter flag against the list of slice values.
func filterFlag(t interface{}, val string) {
	if val == "" {
		return
	}
	switch t := t.(type) {
	case []string:
		k := false
		for _, value := range t {
			if value == val || (val == value[:1]) {
				k = true
				break
			}
		}
		if !k {
			logs.Check(fmt.Errorf("unsupported --filter flag %q, valid flags: %s", val, strings.Join(t, ", ")))
			os.Exit(1)
		}
	}
}
