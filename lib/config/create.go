package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color" //nolint:misspell
	"github.com/spf13/viper"
)

// Create a configuration file.
func Create(ow bool) error {
	Config.ignore = true
	if cfg := viper.ConfigFileUsed(); cfg != "" && !ow {
		if _, err := os.Stat(cfg); !os.IsNotExist(err) {
			color.Warn.Println("a config file already is in use")
			logs.Printf("to edit:\t%s %s\nto remove:\t%s %s\n", cmdPath, "edit", cmdPath, "delete")
			os.Exit(1)
		}
		p := filepath.Dir(cfg)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			logs.Println(p)
			if err := os.MkdirAll(p, dir); err != nil {
				return fmt.Errorf("create mkdir %q: %w", dir, err)
			}
		}
	}
	if err := writeConfig(false); err != nil {
		return fmt.Errorf("create: %w", err)
	}
	return nil
}
