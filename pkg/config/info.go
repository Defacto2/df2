package config

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const colon = ":"

// Info prints the content of a configuration file.
func Info(w io.Writer, l *zap.SugaredLogger, sizes bool) error {
	fmt.Fprint(w, "\nDefault configurations in use when no flags are given.\n\n")
	sets, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		return fmt.Errorf("info config yaml marshal: %w", err)
	}
	fmt.Fprintf(w, "%v%v %v\n", color.Cyan.Sprint("config file"), color.Red.Sprint(colon), Filepath())
	Check()
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
			fmt.Fprintln(w)
			continue
		}
		val := strings.TrimSpace(strings.Join(s[1:], colon))
		switch strings.TrimSpace(s[0]) {
		case "server":
			if db != "" {
				fmt.Fprintf(w, " %s %s", str.X(), db)
				break
			}
			logFmt(w, color.Success.Sprint("up"), str.Y())
		case `"000"`, `"400"`, "backup", "emu", "html", "files", "previews", "sql", "root", "views", "uuid":
			if err := parse(w, sizes, val); err != nil {
				return err
			}
		case "password":
			fmt.Fprint(w, color.Warn.Sprint(" **********"))
		default:
			fmt.Fprintf(w, " %s", val)
		}
		fmt.Fprintln(w)
	}
	return nil
}

func parse(w io.Writer, sizes bool, val string) error {
	_, err := os.Stat(val)
	switch {
	case os.IsNotExist(err):
		logFmt(w, val, str.X())
	case err != nil:
		return fmt.Errorf("info stat %q: %w", val, err)
	default:
		if !sizes {
			logFmt(w, val, str.Y())
			break
		}
		count, size, err := directories.Size(val)
		if err != nil {
			log.Println(err)
		}
		if count == 0 {
			fmt.Fprintf(w, " %s (0 files) %s", val, str.Y())
			break
		}
		fmt.Fprintf(w, " %s (%d files, %s) %s", val, count, humanize.Bytes(size), str.Y())
	}
	return nil
}

func logFmt(w io.Writer, s, mark string) {
	fmt.Fprintf(w, " %s %s", s, mark)
}
