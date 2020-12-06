package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/prompt"
	"github.com/Defacto2/df2/lib/str"
	"github.com/gookit/color"
	"github.com/spf13/viper"
)

// Set edits and saves a setting within a configuration file.
func Set(name string) error {
	if viper.ConfigFileUsed() == "" {
		configMissing("set")
	}
	keys := viper.AllKeys()
	sort.Strings(keys)
	// prefix name alias
	if strings.HasPrefix(name, "directories.") {
		name = strings.Replace(name, "directories.", "directory.", 1)
	}
	// suffix aliases, name equaling "uuid" will match "directory.uuid"
	for i, key := range keys {
		if strings.Contains(key, "."+name) {
			name = keys[i]
		}
	}
	// sort.SearchStrings() - The slice must be sorted in ascending order.
	if i := sort.SearchStrings(keys, name); i == len(keys) || keys[i] != name {
		logs.Printf("%s\n%s %s\n",
			color.Warn.Sprintf("invalid flag value %v", fmt.Sprintf("--name %s", name)),
			"to see a list of usable settings run:",
			color.Bold.Sprint("df2 config info"))
		os.Exit(1)
	}
	Config.nameFlag = name
	if err := sets(name); err != nil {
		return fmt.Errorf("set %s: %w", name, err)
	}
	return nil
}

func sets(name string) error {
	rec := func(value string) string {
		return color.Info.Sprintf("(recommend: %v)", value)
	}
	s := viper.GetString(name)
	switch s {
	case "":
		fmt.Printf("\n%s is currently disabled\n", name)
	default:
		fmt.Printf("\n%s is currently set to \"%v\"\n", name, color.Primary.Sprint(s))
	}
	switch {
	case name == "connection.server.host":
		fmt.Printf("\nSet a new host, leave blank to keep as-is %v: \n", rec("localhost"))
		return configSave(prompt.String(s))
	case name == "connection.server.protocol":
		fmt.Printf("\nSet a new protocol, leave blank to keep as-is %v: \n", rec("tcp"))
		return configSave(prompt.String(s))
	case name == "connection.server.port":
		fmt.Printf("Set a new MySQL port, choices: %v-%v %v\n", prompt.PortMin, prompt.PortMax, rec("3306"))
		return configSave(prompt.Port())
	case name[:10] == "directory.":
		fmt.Printf("\nSet a new directory or leave blank to keep as-is: \n")
		return configSave(prompt.Dir())
	case name == "connection.password":
		fmt.Printf("\nSet a new MySQL user encrypted or plaintext password or leave blank to keep as-is: \n")
		return configSave(prompt.String(s))
	default:
		fmt.Printf("\nSet a new value, leave blank to keep as-is or use a dash [-] to disable: \n")
		return configSave(prompt.String(s))
	}
}

func configSave(value interface{}) error {
	switch value.(type) {
	case int64, string:
	default:
		return fmt.Errorf("config save: %w", ErrSaveType)
	}
	viper.Set(Config.nameFlag, value)
	logs.Printf("%s %s is now set to \"%v\"\n", str.Y(), color.Primary.Sprint(Config.nameFlag), color.Info.Sprint(value))
	if err := writeConfig(true); err != nil {
		return fmt.Errorf("config save: %w", err)
	}
	return nil
}
