// Package cmd handles the commandline user interface and interactions.
// nolint:gochecknoglobals
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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

	configName = ""
	panics     = false // debug log
	quiet      = false // quiet disables most printing or output to terminal
	simulate   bool
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "df2",
	Short: "The tool to optimise and manage defacto2.net",
	Long: fmt.Sprintf("%s\nCopyright © %v Ben Garrett\n%v",
		color.Info.Sprint("The tool to optimise and manage defacto2.net"),
		copyright(),
		color.Primary.Sprint("https://github.com/Defacto2/df2")),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runNew(); err != nil {
			logs.Fatal(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(color.Warn.Sprintf("%s", err))
		if e := err.Error(); strings.Contains(e, "required flag(s) \"name\"") {
			logs.Println("see Examples for usage or run to list setting choices:",
				color.Bold.Sprintf("%s config info", rootCmd.CommandPath()))
		}
		os.Exit(1)
	}
	config.Check()
}

func init() { // nolint:gochecknoinits
	cobra.OnInitialize()
	initConfig()
	rootCmd.PersistentFlags().StringVar(&configName, "config", "",
		fmt.Sprintf("config file (default is %s)", config.Filepath()))
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false,
		"suspend feedback to the terminal")
	rootCmd.PersistentFlags().BoolVar(&panics, "panic", false,
		"panic in the disco")
	if err := rootCmd.PersistentFlags().MarkHidden("panic"); err != nil {
		logs.Fatal(err)
	}
}

// copyright returns a © Copyright year, or a range of years.
func copyright() string {
	const initYear = 2020
	y, c := time.Now().Year(), initYear
	if y == c {
		return strconv.Itoa(c) // © 2020
	}
	return fmt.Sprintf("%s-%s", strconv.Itoa(c), time.Now().Format("06")) // © 2020-21
}

// filterFlag compairs the value of the filter flag against the list of slice values.
func filterFlag(t interface{}, flag, val string) {
	if val == "" {
		return
	}
	if t, ok := t.([]string); ok {
		sup := false
		for _, value := range t {
			if value == val || (val == value[:1]) {
				sup = true
				break
			}
		}
		if !sup {
			fmt.Printf("%s %s\n%s %s\n", color.Warn.Sprintf("unsupported --%s flag value", flag),
				color.Bold.Sprintf("%q", val),
				color.Warn.Sprint("available flag values"),
				color.Primary.Sprint(strings.Join(t, ",")))
			os.Exit(1)
		}
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	logs.Panic(panics)
	logs.Quiet(quiet)
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
	if !quiet && !str.Piped() {
		logs.Println(str.Sec(fmt.Sprintf("config file in use: %s",
			viper.ConfigFileUsed())))
	}
}
