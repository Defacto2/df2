package cmd

// os.Exit() = 10x

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/config"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/demozoo"
	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/images"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/proof"
	"github.com/Defacto2/df2/lib/text"
	"github.com/gookit/color"
	"github.com/hako/durafmt"
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

const version string = "0.9.15" // df2 version

var (
	copyright       = copyYears()
	configName      = ""
	fmtflags        = []string{"datalist", "html", "text", "dl", "d", "h", "t"}
	panic      bool = false // debug log
	quiet      bool = false // quiet disables most printing or output to terminal
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

var lookupCmd = &cobra.Command{
	Use:   "lookup (id|uuid)",
	Short: "Lookup the file URL of a database ID or UUID",
	Example: `  id is a a unique numeric identifier
  uuid is a unique 35-character hexadecimal string representation of a 128-bit integer
  uuid character groups are 8-4-4-16 (xxxxxxxx-xxxx-xxxx-xxxxxxxxxxxxxxxx)`,
	Hidden: false,
	Args: func(cmd *cobra.Command, args []string) error {
		const help = ""
		if len(args) != 1 {
			return errors.New("lookup: requires an id or uuid argument")
		}
		if err := database.CheckID(args[0]); err == nil {
			return nil
		}
		return fmt.Errorf("lookup: invalid id or uuid specified %q", args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		id, err := database.LookupID(args[0])
		if err != nil {
			fmt.Printf("%s\n", err)
		} else {
			fmt.Printf("https://defacto2.net/f/%v\n", database.ObfuscateParam(fmt.Sprint(id)))
		}
	},
}

var logCmd = &cobra.Command{
	Use:    "log",
	Short:  "Display the df2 error log",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		logs.Printf("%v%v %v\n", color.Cyan.Sprint("log file"), color.Red.Sprint(":"), logs.Filepath())
		f, err := os.Open(logs.Filepath())
		logs.Check(err)
		scanner := bufio.NewScanner(f)
		c := 0
		scanner.Text()
		for scanner.Scan() {
			c++
			s := strings.SplitN(scanner.Text(), " ", 5)
			t, err := time.Parse("2006/01/02 15:04:05", strings.Join(s[0:2], " "))
			if err != nil {
				fmt.Printf("%d. %v\n", c, scanner.Text())
				continue
			}
			// todo get system local timezone and set it here
			// OR log file to use UTC
			logs.Printf("%v\n", t)
			duration := durafmt.Parse(time.Since(t)).LimitFirstN(1)
			fmt.Printf("%v %v ago  %v %s\n", color.Secondary.Sprintf("%d.", c), duration, color.Info.Sprint(s[2]), strings.Join(s[3:], " "))
		}
		if err := scanner.Err(); err != nil {
			logs.Check(err)
		}
	},
}

var waitCmd = &cobra.Command{
	Use:   "waiting",
	Short: "Handler for files flagged as waiting to go live",
	Long: `Runs a sequence of commands to handle files waiting to go live.

  df2 demozoo --new
      proof
      fix images
      fix text
      fix demozoo
      fix database`,
	Run: func(cmd *cobra.Command, args []string) {
		config.ErrCheck()
		var err error
		// demozoo handler
		dz := demozoo.Request{
			All:       false,
			Overwrite: false,
			Refresh:   false,
			Simulate:  false}
		err = dz.Queries()
		logs.Check(err)
		// proofs handler
		p := proof.Request{
			Overwrite:   false,
			AllProofs:   false,
			HideMissing: false}
		err = p.Queries()
		logs.Check(err)
		// missing image previews
		err = images.Fix(false)
		logs.Check(err)
		// missing text file previews
		err = text.Fix(false)
		logs.Check(err)
		// fix database entries
		demozoo.Fix()
		database.Fix()
		groups.Fix(false)
	},
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

func copyYears() string {
	var y, c int = time.Now().Year(), 2020
	if y == c {
		return strconv.Itoa(c)
	}
	return fmt.Sprintf("%s-%s", strconv.Itoa(c), time.Now().Format("06")) // 2020-21
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&configName, "config", "", fmt.Sprintf("config file (default is %s)", config.Filepath()))
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suspend feedback to the terminal")
	rootCmd.PersistentFlags().BoolVar(&panic, "panic", false, "panic in the disco")
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(lookupCmd)
	rootCmd.AddCommand(waitCmd)
	err := rootCmd.PersistentFlags().MarkHidden("panic")
	logs.Check(err)
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
