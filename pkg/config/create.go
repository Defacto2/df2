package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gookit/color"
	"github.com/spf13/viper"
)

// Create a configuration file.
func Create(w io.Writer, ow bool) error {
	Config.ignore = true
	if cfg := viper.ConfigFileUsed(); cfg != "" && !ow {
		if _, err := os.Stat(cfg); !os.IsNotExist(err) {
			color.Warn.Println("a config file already is in use")
			fmt.Fprintf(w, "to edit:\t%s %s\nto remove:\t%s %s\n", cmdRun, "edit", cmdRun, "delete")
			os.Exit(1)
		}
		p := filepath.Dir(cfg)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			fmt.Fprintln(w, p)
			if err := os.MkdirAll(p, dir); err != nil {
				return fmt.Errorf("create mkdir %q: %w", dir, err)
			}
		}
	}
	if err := write(w, false); err != nil {
		return fmt.Errorf("create: %w", err)
	}
	return nil
}
