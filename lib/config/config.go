// Package config saves and fetches settings used by the df2 tool.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
	gap "github.com/muesli/go-app-paths"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var (
	ErrSaveType = errors.New("unsupported value interface type")
)

// directories are initialised and configured by InitDefaults() in lib/cmd.go.

const (
	cmdRun               = "df2 config"
	filename             = "config.yaml"
	dir      fs.FileMode = 0o700
	file     fs.FileMode = 0o600
)

// Settings for the configuration.
type Settings struct {
	Name     string // config filename
	Errors   bool   // flag a config file error, used by root.go initConfig()
	ignore   bool   // ignore config file error
	nameFlag string // viper configuration path
}

// Config settings.
var Config = Settings{ //nolint:gochecknoglobals
	Name:   filename,
	Errors: false,
	ignore: false,
}

// Check prints a notice for the missing configuration file.
func Check() {
	if Config.ignore {
		return
	}
	if Config.Errors {
		if logs.IsQuiet() {
			os.Exit(1)
		}
		fmt.Printf("%s %s\n",
			color.Warn.Sprint("config: no config file in use, please run"),
			color.Bold.Sprintf(cmdRun+" create"))
	}
}

// Filepath is the absolute path and filename of the configuration file.
func Filepath() string {
	dir, err := gap.NewScope(gap.User, logs.GapUser).ConfigPath(filename)
	if err != nil {
		h, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(fmt.Errorf("filepath userhomedir: %w", err))
		}
		return filepath.Join(h, filename)
	}
	return dir
}

func missing(suffix string) {
	color.Warn.Println("no config file is in use")
	logs.Printf("to create:\t%s %s\n", cmdRun, suffix)
}

// write saves all configs to a configuration file.
func write(update bool) error {
	bs, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		return fmt.Errorf("write config yaml marshal: %w", err)
	}
	err = ioutil.WriteFile(Filepath(), bs, file) // owner+wr
	if err != nil {
		return fmt.Errorf("write config file %s: %w", file, err)
	}
	s := "created a new"
	if update {
		s = "updated the"
	}
	logs.Println(s+" config file", Filepath())
	return nil
}
