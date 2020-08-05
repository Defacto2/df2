package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/gookit/color"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Info prints the content of a configuration file.
func Info() error {
	logs.Print("\nDefault configurations in use when no flags are given.\n\n")
	sets, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		return fmt.Errorf("info config yaml marshal: %w", err)
	}
	logs.Printf("%v%v %v\n", color.Cyan.Sprint("config file"), color.Red.Sprint(":"), Filepath())
	Check()
	logs.Printf("%v%v %v\n", color.Cyan.Sprint("log file"), color.Red.Sprint(":"), logs.Filepath())
	db := database.ConnectInfo()
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
			if db != "" {
				logs.Printf(" %s %s", str.X(), db)
			} else {
				logs.Printf(" %s %s", color.Success.Sprint("up"), str.Y())
			}
		case `"000"`, `"150"`, `"400"`, "backup", "emu", "html", "files", "previews", "sql", "root", "views", "uuid":
			_, err := os.Stat(val)
			switch {
			case os.IsNotExist(err):
				logs.Printf(" %s %s", val, str.X())
			case err != nil:
				return fmt.Errorf("info stat %q: %w", val, err)
			default:
				logs.Printf(" %s %s", val, str.Y())
			}
		case "password":
			logs.Print(color.Warn.Sprint(" **********"))
		default:
			logs.Printf(" %s", val)
		}
		logs.Println()
	}
	return nil
}
