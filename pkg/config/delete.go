package config

import (
	"fmt"
	"io"
	"os"

	"github.com/Defacto2/df2/pkg/prompt"
	"github.com/spf13/viper"
)

// Delete a configuration file.
func Delete(w io.Writer) error {
	cfg := viper.ConfigFileUsed()
	if cfg == "" {
		missing(w, "delete")
	}
	if _, err := os.Stat(cfg); os.IsNotExist(err) {
		missing(w, "delete")
	}
	if ok := prompt.YN("Remove the config file", false); ok {
		if err := os.Remove(cfg); err != nil {
			return fmt.Errorf("delete remove %q: %w", cfg, err)
		}
		fmt.Fprintln(w, "the config is gone")
	}
	os.Exit(0)
	return nil
}
