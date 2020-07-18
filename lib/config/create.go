package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
	"github.com/spf13/viper"
)

// Create a configuration file.
func Create(ow bool) {
	Config.ignore = true
	if cfg := viper.ConfigFileUsed(); cfg != "" && !ow {
		if _, err := os.Stat(cfg); !os.IsNotExist(err) {
			configExists(cmdPath, "create")
		}
		p := filepath.Dir(cfg)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			logs.Println(p)
			if err := os.MkdirAll(p, dir); err != nil {
				logs.Check(err)
				os.Exit(1)
			}
		}
	}
	writeConfig(false)
}

func configExists(name, suffix string) {
	cmd := strings.TrimSuffix(name, suffix)
	color.Warn.Println("a config file already is in use")
	logs.Printf("to edit:\t%s %s\n", cmd, "edit")
	logs.Printf("to remove:\t%s %s\n", cmd, "delete")
	os.Exit(1)
}
