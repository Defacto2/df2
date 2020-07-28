package cmd

// os.Exit() = 10x

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Defacto2/df2/lib/config"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var simulate bool

const version string = "1.1.0" // df2 version

const verTmp = `
   df2 tool version {{.Version}}  
`

var (
	ErrCmd  = errors.New("invalid command, please use one of the available commands")
	ErrNoID = errors.New("requires an id or uuid argument")
	ErrID   = errors.New("invalid id or uuid specified")
)

var (
	configName = ""
	panic      = false // debug log
	quiet      = false // quiet disables most printing or output to terminal
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     "df2",
	Example: "  df2 new",
	Short:   "A tool to optimise and manage defacto2.net",
	Long: fmt.Sprintf("%s\nCopyright © %v Ben Garrett\n%v",
		color.Info.Sprint("A tool to optimise and manage defacto2.net"),
		copyright(),
		color.Primary.Sprint("https://github.com/Defacto2/df2")),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	execVersion()
	rootCmd.SetVersionTemplate(verTmp)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(color.Warn.Sprintf("%s", err))
		e := err.Error()
		if strings.Contains(e, "required flag(s) \"name\"") {
			logs.Println("see Examples for usage or run to list setting choices:", color.Bold.Sprintf("%s config info", rootCmd.CommandPath()))
		}
		os.Exit(1)
	}
	config.Check()
}

func init() {
	cobra.OnInitialize()
	initConfig()
	rootCmd.PersistentFlags().StringVar(&configName, "config", "", fmt.Sprintf("config file (default is %s)", config.Filepath()))
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suspend feedback to the terminal")
	rootCmd.PersistentFlags().BoolVar(&panic, "panic", false, "panic in the disco")
	if err := rootCmd.PersistentFlags().MarkHidden("panic"); err != nil {
		logs.Fatal(err)
	}
}

// copyright returns a © Copyright year, or a range of years.
func copyright() string {
	var y, c int = time.Now().Year(), 2020
	if y == c {
		return strconv.Itoa(c) // © 2020
	}
	return fmt.Sprintf("%s-%s", strconv.Itoa(c), time.Now().Format("06")) // © 2020-21
}

func execVersion() {
	type VerData struct {
		Version  string
		Database string
		Ansilove string
		Webp     string
	}
	const exeTmp = `
┌──────────────────────────┐
│  database: {{.Database}}  │
│  ansilove: {{.Ansilove}}  │
│  webp lib: {{.Webp}}  │
└──────────────────────────┘
`
	p := func(s string) string {
		switch s {
		case "ok":
			return color.Success.Sprint("okay") + strings.Repeat(" ", 10-len(s))
		case "missing":
			return color.Error.Sprint("missing") + strings.Repeat(" ", 12-len(s))
		case "disconnect":
			return color.Error.Sprint("disconnected") + strings.Repeat(" ", 10-len(s))
		}
		return ""
	}
	a, w, d := "missing", "missing", "disconnect"
	if err := database.ConnectInfo(); err == "" {
		d = "ok"
	}
	if _, err := exec.LookPath("ansilove"); err == nil {
		a = "ok"
	}
	if _, err := exec.LookPath("webpng"); err == nil {
		w = "ok"
	}
	var data = VerData{
		Version:  color.Primary.Sprint(version),
		Database: p(d),
		Ansilove: p(a),
		Webp:     p(w),
	}
	tmpl, err := template.New("version").Parse(exeTmp)
	if err != nil {
		log.Fatal(err)
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		log.Fatal(err)
	}
	rootCmd.Version = color.Primary.Sprint(version) + b.String()
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
	logs.Panic = panic
	logs.Quiet = quiet
	cf := config.Filepath()
	if cf != "" {
		viper.SetConfigFile(cf)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			logs.Fatal(err)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(config.Config.Name)
	}
	viper.AutomaticEnv() // read in environment variables that match
	// if a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		config.Config.Errors = true
	} else if !quiet {
		logs.Println(str.Sec(fmt.Sprintf("config file in use: %s", viper.ConfigFileUsed())))
	}
}
