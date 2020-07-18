package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
	"github.com/spf13/viper"
)

// Set edits and saves a setting within a configuration file.
func Set(name string) {
	if viper.ConfigFileUsed() == "" {
		configMissing(cmdPath, "set")
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
		os.Exit(202)
	}
	Config.nameFlag = name
	s := viper.GetString(name)
	switch s {
	case "":
		fmt.Printf("\n%s is currently disabled\n", name)
	default:
		fmt.Printf("\n%s is currently set to \"%v\"\n", name, color.Primary.Sprint(s))
	}
	switch {
	case name == "connection.server.host":
		fmt.Printf("\nSet a new host, leave blank to keep as-is %v: \n", recommend("localhost"))
		configSave(logs.PromptString(s))
	case name == "connection.server.protocol":
		fmt.Printf("\nSet a new protocol, leave blank to keep as-is %v: \n", recommend("tcp"))
		configSave(logs.PromptString(s))
	case name == "connection.server.port":
		fmt.Printf("Set a new MySQL port, choices: %v-%v %v\n", logs.PortMin, logs.PortMax, recommend("3306"))
		configSave(logs.PromptPort())
	case name[:10] == "directory.":
		fmt.Printf("\nSet a new directory or leave blank to keep as-is: \n")
		configSave(logs.PromptDir())
	case name == "connection.password":
		fmt.Printf("\nSet a new MySQL user encrypted or plaintext password or leave blank to keep as-is: \n")
		configSave(logs.PromptString(s))
	default:
		fmt.Printf("\nSet a new value, leave blank to keep as-is or use a dash [-] to disable: \n")
		configSave(logs.PromptString(s))
	}
}

func configSave(value interface{}) {
	switch value.(type) {
	case int64, string:
	default:
		logs.Check(fmt.Errorf("config save: unsupported value interface type"))
	}
	viper.Set(Config.nameFlag, value)
	logs.Printf("%s %s is now set to \"%v\"\n", logs.Y(), color.Primary.Sprint(Config.nameFlag), color.Info.Sprint(value))
	writeConfig(true)
}

func recommend(value string) string {
	return color.Info.Sprintf("(recommend: %v)", value)
}
