package config

import (
	"fmt"
	"os"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/viper"
)

// Delete a configuration file.
func Delete() {
	cfg := viper.ConfigFileUsed()
	if cfg == "" {
		configMissing("delete")
	}
	if _, err := os.Stat(cfg); os.IsNotExist(err) {
		configMissing("delete")
	}
	if ok := logs.PromptYN("Remove the config file", false); ok {
		if err := os.Remove(cfg); err != nil {
			logs.Check(fmt.Errorf("config delete: could not remove %v %v", cfg, err))
		}
		logs.Println("the config is gone")
	}
	os.Exit(0)
}
