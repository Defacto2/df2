package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const colon = ":"

// Info prints the content of a configuration file.
func Info(sizes bool) error {
	logs.Print("\nDefault configurations in use when no flags are given.\n\n")
	sets, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		return fmt.Errorf("info config yaml marshal: %w", err)
	}
	logs.Printf("%v%v %v\n", color.Cyan.Sprint("config file"), color.Red.Sprint(colon), Filepath())
	Check()
	logs.Printf("%v%v %v\n", color.Cyan.Sprint("log file"), color.Red.Sprint(colon), logs.Filepath(logs.Filename))
	db := database.ConnInfo()
	scanner := bufio.NewScanner(strings.NewReader(string(sets)))
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), colon)
		if s[0] == "directory" {
			s[0] = "directories"
		}
		color.Cyan.Print(s[0])
		color.Red.Print(colon)
		if len(s) <= 1 {
			logs.Println()
			continue
		}
		val := strings.TrimSpace(strings.Join(s[1:], colon))
		switch strings.TrimSpace(s[0]) {
		case "server":
			if db != "" {
				logs.Printf(" %s %s", str.X(), db)
				break
			}
			logFmt(color.Success.Sprint("up"), str.Y())
		case `"000"`, `"400"`, "backup", "emu", "html", "files", "previews", "sql", "root", "views", "uuid":
			if err := parse(sizes, val); err != nil {
				return err
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

func parse(sizes bool, val string) error {
	_, err := os.Stat(val)
	switch {
	case os.IsNotExist(err):
		logFmt(val, str.X())
	case err != nil:
		return fmt.Errorf("info stat %q: %w", val, err)
	default:
		if !sizes {
			logFmt(val, str.Y())
			break
		}
		count, size, err := directories.Size(val)
		if err != nil {
			log.Println(err)
		}
		if count == 0 {
			logs.Printf(" %s (0 files) %s", val, str.Y())
			break
		}
		logs.Printf(" %s (%d files, %s) %s", val, count, humanize.Bytes(size), str.Y())
	}
	return nil
}

func logFmt(s, mark string) {
	logs.Printf(" %s %s", s, mark)
}
