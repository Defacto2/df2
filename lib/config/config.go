package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
	gap "github.com/muesli/go-app-paths"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// directories are initialized and configured by InitDefaults() in lib/cmd.go

const (
	ConfigName             = "config.yaml"
	cmdPath                = "df2 config"
	dir        os.FileMode = 0700
	file       os.FileMode = 0600
)

// settings configurations.
type settings struct {
	Errors   bool   // flag a config file error, used by root.go initConfig()
	ignore   bool   // ignore config file error
	nameFlag string // viper configuration path
}

var (
	scope = gap.NewScope(gap.User, "df2")
	// Config settings.
	Config = settings{
		Errors: false,
		ignore: false,
	}
)

// Filepath is the absolute path and filename of the configuration file.
func Filepath() (dir string) {
	dir, err := scope.ConfigPath(ConfigName)
	if err != nil {
		h, _ := os.UserHomeDir()
		return filepath.Join(h, ConfigName)
	}
	return dir
}

func configMissing(suffix string) {
	cmd := strings.TrimSuffix(cmdPath, suffix) + "create"
	color.Warn.Println("no config file is in use")
	logs.Printf("to create:\t%s\n", cmd)
}

// writeConfig saves all configs to a configuration file.
func writeConfig(update bool) {
	bs, err := yaml.Marshal(viper.AllSettings())
	logs.Check(err)
	err = ioutil.WriteFile(Filepath(), bs, file) // owner+wr
	logs.Check(err)
	s := "created a new"
	if update {
		s = "updated the"
	}
	logs.Println(s+" config file", Filepath())
}
