package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
	gap "github.com/muesli/go-app-paths"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// directories are initialized and configured by InitDefaults() in lib/cmd.go.

const (
	cmdPath              = "df2 config"
	filename             = "config.yaml"
	dir      os.FileMode = 0700
	file     os.FileMode = 0600
)

var ErrSaveType = errors.New("unsupported value interface type")

// settings configurations.
type settings struct {
	Errors   bool   // flag a config file error, used by root.go initConfig()
	ignore   bool   // ignore config file error
	Name     string // config filename
	nameFlag string // viper configuration path
}

var (
	scope = gap.NewScope(gap.User, "df2")
	// Config settings.
	Config = settings{
		Name:   filename,
		Errors: false,
		ignore: false,
	}
)

// Check prints a missing configuration file notice.
func Check() {
	if Config.ignore {
		return
	}
	if Config.Errors && !logs.Quiet {
		fmt.Printf("%s %s\n",
			color.Warn.Sprint("config: no config file in use, please run"),
			color.Bold.Sprintf("df2 config create"))
	} else if Config.Errors {
		os.Exit(1)
	}
}

// Filepath is the absolute path and filename of the configuration file.
func Filepath() (dir string) {
	dir, err := scope.ConfigPath(filename)
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
	cmd := strings.TrimSuffix(cmdPath, suffix) + "create"
	color.Warn.Println("no config file is in use")
	logs.Printf("to create:\t%s\n", cmd)
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
