package config

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Defacto2/df2/pkg/prompt"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/gookit/color"
	"github.com/spf13/viper"
)

var (
	ErrSetName = errors.New("invalid flag name")
	ErrPort    = errors.New("value is not a valid port")
)

type SetFlags struct {
	Host     string
	Protocol string
	Port     int
}

func (flag SetFlags) Set(w io.Writer) error {
	const (
		host = "connection.server.host"
		prot = "connection.server.protocol"
		port = "connection.server.port"
	)
	if viper.ConfigFileUsed() == "" {
		missing(w, "set")
	}
	if flag.Host != "" {
		viper.Set(host, flag.Host)
		fmt.Fprintf(w, "%s %s is now set to \"%v\"\n", str.Y(),
			color.Primary.Sprint(host), color.Info.Sprint(flag.Host))
	}
	if i := flag.Port; i >= prompt.PortMin {
		if ok := prompt.IsPort(i); !ok {
			return fmt.Errorf("%w (%d-%d): %d",
				ErrPort, prompt.PortMin, prompt.PortMax, i)
		}
		viper.Set(port, i)
		fmt.Fprintf(w, "%s %s is now set to \"%v\"\n", str.Y(),
			color.Primary.Sprint(port), color.Info.Sprint(i))
	}
	if flag.Protocol != "" {
		viper.Set(prot, flag.Protocol)
		fmt.Fprintf(w, "%s %s is now set to \"%v\"\n", str.Y(),
			color.Primary.Sprint(prot), color.Info.Sprint(flag.Protocol))
	}
	if err := write(w, true); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	return nil
}

// Set edits and saves a setting within a configuration file.
func Set(w io.Writer, name string) error {
	if viper.ConfigFileUsed() == "" {
		missing(w, "set")
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
		fmt.Fprintf(w, "%s\n%s %s\n",
			color.Warn.Sprintf("invalid flag value %v", fmt.Sprintf("--name %s", name)),
			"to see a list of usable settings run:",
			color.Bold.Sprint("df2 config info"))
		return ErrSetName
	}
	Config.nameFlag = name
	if err := sets(w, name); err != nil {
		return fmt.Errorf("set %s: %w", name, err)
	}
	return nil
}

func sets(w io.Writer, name string) error {
	const suffix = 10
	rec := func(value string) string {
		return color.Info.Sprintf("(recommend: %v)", value)
	}
	s := viper.GetString(name)
	switch s {
	case "":
		fmt.Fprintf(w, "\n%s is currently disabled\n", name)
		return nil
	default:
		fmt.Fprintf(w, "\n%s is currently set to \"%v\"\n", name, color.Primary.Sprint(s))
	}
	switch {
	case name == "connection.server.host":
		fmt.Fprintf(w, "\nSet a new host, leave blank to keep as-is %v: \n", rec("localhost"))
		return save(w, prompt.String(s))
	case name == "connection.server.protocol":
		fmt.Fprintf(w, "\nSet a new protocol, leave blank to keep as-is %v: \n", rec("tcp"))
		return save(w, prompt.String(s))
	case name == "connection.server.port":
		fmt.Fprintf(w, "Set a new MySQL port, choices: %v-%v %v\n", prompt.PortMin, prompt.PortMax, rec("3306"))
		return save(w, prompt.Port())
	case name[:suffix] == "directory.":
		fmt.Fprintf(w, "\nSet a new directory or leave blank to keep as-is: \n")
		return save(w, prompt.Dir())
	case name == "connection.password":
		fmt.Fprintf(w, "\nSet a new MySQL user encrypted or plaintext password or leave blank to keep as-is: \n")
		return save(w, prompt.String(s))
	default:
		fmt.Fprintf(w, "\nSet a new value, leave blank to keep as-is or use a dash [-] to disable: \n")
		return save(w, prompt.String(s))
	}
}

func save(w io.Writer, value any) error {
	switch value.(type) {
	case int64, string:
	default:
		return fmt.Errorf("save: %w", ErrSaveType)
	}
	viper.Set(Config.nameFlag, value)
	fmt.Fprintf(w, "%s %s is now set to \"%v\"\n", str.Y(), color.Primary.Sprint(Config.nameFlag), color.Info.Sprint(value))
	if err := write(w, true); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	return nil
}
