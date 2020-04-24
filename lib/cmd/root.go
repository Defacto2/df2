package cmd

// os.Exit() = 10x

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/config"
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

var simulate bool

const version string = "0.10.0" // df2 version

var (
	copyright  string = cYear()
	configName        = ""
	fmtflags          = [7]string{"datalist", "html", "text", "dl", "d", "h", "t"}
	panic      bool   = false // debug log
	quiet      bool   = false // quiet disables most printing or output to terminal
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "df2",
	Short: "A tool to optimise and manage defacto2.net",
	Long: fmt.Sprintf("%s\nCopyright © %v Ben Garrett\n%v",
		color.Info.Sprint("A tool to optimise and manage defacto2.net"),
		copyright,
		color.Primary.Sprint("https://github.com/Defacto2/df2")),
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
			logs.Println("see Examples for usage or run to list setting choices:", color.Bold.Sprintf("%s config info", rootCmd.CommandPath()))
		}
		os.Exit(100)
	}
	config.ErrCheck()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&configName, "config", "", fmt.Sprintf("config file (default is %s)", config.Filepath()))
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suspend feedback to the terminal")
	rootCmd.PersistentFlags().BoolVar(&panic, "panic", false, "panic in the disco")
	err := rootCmd.PersistentFlags().MarkHidden("panic")
	logs.Check(err)
}

// cYear returns a © Copyright year, or a range of years.
func cYear() string {
	var y, c int = time.Now().Year(), 2020
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

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	initPanic(panic)
	initQuiet(quiet)
	cf := config.Filepath()
	if cf != "" {
		viper.SetConfigFile(cf)
	} else {
		home, err := os.UserHomeDir()
		logs.Check(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(config.ConfigName)
	}
	viper.AutomaticEnv() // read in environment variables that match
	// if a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		config.Config.Errors = true
	} else if !quiet {
		logs.Println(logs.Sec(fmt.Sprintf("config file in use: %s", viper.ConfigFileUsed())))
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
