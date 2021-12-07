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

// directories are initialised and configured by InitDefaults() in lib/cmd.go.

const (
	cmdPath              = "df2 config"
	filename             = "config.yaml"
	dir      fs.FileMode = 0o700
	file     fs.FileMode = 0o600
)

// ErrSaveType bad value type.
var ErrSaveType = errors.New("unsupported value interface type")

// settings configurations.
type settings struct {
	Errors   bool   // flag a config file error, used by root.go initConfig()
	ignore   bool   // ignore config file error
	Name     string // config filename
	nameFlag string // viper configuration path
}

// Config settings.
var Config = settings{ //nolint:gochecknoglobals
	Name:   filename,
	Errors: false,
	ignore: false,
}

// Check prints a missing configuration file notice.
func Check() {
	if Config.ignore {
		return
	}
	if Config.Errors {
		if logs.Quiet {
			os.Exit(1)
		}
		fmt.Printf("%s %s\n",
			color.Warn.Sprint("config: no config file in use, please run"),
			color.Bold.Sprintf("df2 config create"))
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

func configMissing(suffix string) {
	color.Warn.Println("no config file is in use")
	logs.Printf("to create:\t%s %s\n", cmdPath, suffix)
}

// writeConfig saves all configs to a configuration file.
func writeConfig(update bool) error {
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
