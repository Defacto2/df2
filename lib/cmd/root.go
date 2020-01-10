package cmd

// os.Exit() = 10x

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
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

type configuration struct {
	errors   bool   // flag a config file error
	filename string // config file persistant flag
	ignore   bool   // ignore config file error
	nameFlag string // viper configuration path
}

var config = configuration{
	errors: false,
	ignore: false,
}
var simulate bool

const (
	configName string = ".df2.yaml" // default configuration filename
	version    string = "0.0.1"     // df2 version
)

var (
	panic     bool = false // debug log
	quiet     bool = false // quiet disables most printing or output to terminal
	home, _        = os.UserHomeDir()
	filepath       = path.Join(home, configName)
	fmtflags       = []string{"html", "text", "h", "t"}
	copyright      = copyYears()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "df2",
	Short: "A tool to optimise and manage defacto2.net",
	Long: fmt.Sprintf(`A tool to optimise and manage defacto2.net
Copyright © %v Ben Garrett
https://github.com/Defacto2/df2
`, copyright),
	Version: color.Primary.Sprint(version) + "\n",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.SetVersionTemplate(`df2 tool version {{.Version}}`)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(color.Warn.Sprintf("%s", err))
		e := err.Error()
		switch {
		case strings.Contains(e, "required flag(s) \"name\""):
			println("see Examples for usage or run to list setting choices:", color.Bold.Sprintf("%s config info", rootCmd.CommandPath()))
		}
		os.Exit(100)
	}
	configErrCheck()
}

func configErrCheck() {
	if !config.ignore {
		configErrMsg()
	}
}

func configErrMsg() {
	if !quiet && config.errors {
		fmt.Printf("%s %s\n", color.Warn.Sprint("no config file in use, please run:"),
			color.Bold.Sprintf("%s config create", rootCmd.CommandPath()))
		os.Exit(102)
	} else if config.errors {
		os.Exit(101)
	}
}

func copyYears() string {
	var y, c int = time.Now().Year(), 2020
	if y == c {
		return strconv.Itoa(c)
	}
	return fmt.Sprintf("%s-%s", strconv.Itoa(c), time.Now().Format("06")) // 2020-21
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&config.filename, "config", "", fmt.Sprintf("config file (default is $HOME/%s)", configName))
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suspend feedback to the terminal")
	rootCmd.PersistentFlags().BoolVar(&panic, "panic", false, "panic in the disco")
	err := rootCmd.PersistentFlags().MarkHidden("panic")
	logs.Check(err)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	initPanic(panic)
	initQuiet(quiet)
	if config.filename != "" {
		viper.SetConfigFile(config.filename)
	} else {
		viper.AddConfigPath(home)
		viper.SetConfigName(configName)
	}
	viper.AutomaticEnv() // read in environment variables that match
	// if a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			config.errors = true
		}
	} else if !quiet {
		println(logs.Sec(fmt.Sprintf("config file in use: %s", viper.ConfigFileUsed())))
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
func filterFlag(t interface{}, flag, val string) {
	if val == "" {
		return
	}
	switch t := t.(type) {
	case []string:
		ok := false
		for _, value := range t {
			if value == val || (val == value[:1]) {
				ok = true
				break
			}
		}
		if !ok {
			fmt.Printf("%s %s\n%s %s\n", color.Warn.Sprintf("unsupported --%s flag value", flag),
				color.Bold.Sprintf("%q", val),
				color.Warn.Sprint("available flag values"),
				color.Primary.Sprint(strings.Join(t, ",")))
			os.Exit(103)
		}
	}
}
