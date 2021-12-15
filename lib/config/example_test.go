package config_test

import (
	"fmt"

	"github.com/Defacto2/df2/lib/config"
	"github.com/spf13/viper"
)

func ExampleSet() {
	viper.SetConfigFile("test.cfg")
	if err := config.Set("directory.000"); err != nil {
		fmt.Print(err)
	}
	// Output:invalid flag value --name directory.000
	// to see a list of usable settings run: df2 config info
	// invalid flag name
}
