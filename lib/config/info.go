package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Info prints the content of a configuration file.
func Info() {
	logs.Print("\nDefault configurations in use when no flags are given.\n\n")
	sets, err := yaml.Marshal(viper.AllSettings())
	logs.Check(err)
	logs.Printf("%v%v %v\n", color.Cyan.Sprint("config file"), color.Red.Sprint(":"), Filepath())
	ErrCheck()
	logs.Printf("%v%v %v\n", color.Cyan.Sprint("log file"), color.Red.Sprint(":"), logs.Filepath())
	dbTest := database.ConnectInfo()
	scanner := bufio.NewScanner(strings.NewReader(string(sets)))
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), ":")
		if s[0] == "directory" {
			s[0] = "directories"
		}
		color.Cyan.Print(s[0])
		color.Red.Print(":")
		if len(s) <= 1 {
			logs.Println()
			continue
		}
		val := strings.TrimSpace(strings.Join(s[1:], ""))
		switch strings.TrimSpace(s[0]) {
		case "server":
			if dbTest != "" {
				logs.Printf(" %s %s", logs.X(), dbTest)
			} else {
				logs.Printf(" %s %s", color.Success.Sprint("up"), logs.Y())
			}
		case `"000"`, `"150"`, `"400"`, "backup", "emu", "html", "files", "previews", "sql", "root", "views", "uuid":
			if _, err := os.Stat(val); os.IsNotExist(err) {
				logs.Printf(" %s %s", val, logs.X())
			} else {
				logs.Printf(" %s %s", val, logs.Y())
			}
		case "password":
			logs.Print(color.Warn.Sprint(" **********"))
		default:
			logs.Printf(" %s", val)
		}
		logs.Println()
	}
}

// ErrCheck prints a missing configuration file notice.
func ErrCheck() {
	if !Config.ignore {
		errMsg()
	}
}

func errMsg() {
	if Config.Errors && !logs.Quiet {
		fmt.Printf("%s %s\n", color.Warn.Sprint("config: no config file in use, please run"),
			color.Bold.Sprintf("df2 config create"))
	} else if Config.Errors {
		os.Exit(1)
	}
}
