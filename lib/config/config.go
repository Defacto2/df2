package config

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// ConfigName is the default configuration filename
const ConfigName string = ".df2.yaml"
const cmdPath = "df2 config"

// Settings configurations.
type settings struct {
	Errors   bool // flag a config file error
	filepath string
	ignore   bool   // ignore config file error
	nameFlag string // viper configuration path
}

var (
	home, _ = os.UserHomeDir()
	// Config settings.
	Config = settings{
		Errors:   false,
		filepath: path.Join(home, ConfigName),
		ignore:   false,
	}
)

// Create a configuration file.
func Create(ow bool) {
	Config.ignore = true
	if cfg := viper.ConfigFileUsed(); cfg != "" && !ow {
		configExists(cmdPath, "create")
	}
	writeConfig(false)
}

// Delete a configuration file.
func Delete() {
	cfg := viper.ConfigFileUsed()
	if cfg == "" {
		configMissing(cmdPath, "delete")
	}
	if _, err := os.Stat(cfg); os.IsNotExist(err) {
		configMissing(cmdPath, "delete")
	}
	switch logs.PromptYN("Remove the config file", false) {
	case true:
		if err := os.Remove(cfg); err != nil {
			logs.Check(fmt.Errorf("could not remove %v %v", cfg, err))
		}
		logs.Println("the config is gone")
	}
	os.Exit(0)
}

// Edit a configuration file.
func Edit() {
	cfg := viper.ConfigFileUsed()
	if cfg == "" {
		configMissing(cmdPath, "edit")
	}
	var edit string
	if err := viper.BindEnv("editor", "EDITOR"); err != nil {
		editors := [3]string{"micro", "nano", "vim"}
		for _, editor := range editors {
			if _, err := exec.LookPath(editor); err == nil {
				edit = editor
				break
			}
		}
		if edit != "" {
			log.Printf("there is no $EDITOR environment variable set so using %s\n", edit)
		} else {
			log.Println("no suitable editor could be found\nplease set one by creating a $EDITOR environment variable in your shell configuration")
			os.Exit(200)
		}
	} else {
		edit = viper.GetString("editor")
		if _, err := exec.LookPath(edit); err != nil {
			log.Printf("%q edit command not found\n%v", edit, exec.ErrNotFound)
			os.Exit(201)
		}
	}
	// credit: https://stackoverflow.com/questions/21513321/how-to-start-vim-from-go
	exe := exec.Command(edit, cfg)
	exe.Stdin = os.Stdin
	exe.Stdout = os.Stdout
	if err := exe.Run(); err != nil {
		logs.Printf("%s\n", err)
	}
}

// ErrCheck prints a missing configuration file notice.
func ErrCheck() {
	if !Config.ignore {
		errMsg()
	}
}

func errMsg() {
	if Config.Errors && !logs.Quiet {
		fmt.Printf("%s %s\n", color.Warn.Sprint("no config file in use, please run:"),
			color.Bold.Sprintf("%s config create", "rootCmd.CommandPath()"))
		os.Exit(102)
	} else if Config.Errors {
		os.Exit(101)
	}
}

// Info prints the content of a configuration file.
func Info() {
	println("These are the default configurations used by this tool when no flags are given.\n")
	sets, err := yaml.Marshal(viper.AllSettings())
	logs.Check(err)
	logs.Printf("%v%v %v\n", color.Cyan.Sprint("config file"), color.Red.Sprint(":"), Config.filepath)
	ErrCheck()
	dbTest := database.ConnectInfo()
	scanner := bufio.NewScanner(strings.NewReader(string(sets)))
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), ":")
		if s[0] == "directory" {
			s[0] = "directories"
		}
		color.Cyan.Print(s[0])
		color.Red.Print(":")
		if len(s) <= 1 {
			logs.Println()
			continue
		}
		val := strings.TrimSpace(strings.Join(s[1:], ""))
		switch strings.TrimSpace(s[0]) {
		case "server":
			if dbTest != "" {
				logs.Printf(" %s %s", logs.X(), dbTest)
			} else {
				logs.Printf(" %s %s", color.Success.Sprint("up"), logs.Y())
			}
		case `"000"`, `"150"`, `"400"`, "backup", "emu", "html", "files", "previews", "root", "uuid":
			if _, err := os.Stat(val); os.IsNotExist(err) {
				logs.Printf(" %s %s", val, logs.X())
			} else {
				logs.Printf(" %s %s", val, logs.Y())
			}
		case "password":
			logs.Print(color.Warn.Sprint(" **********"))
		default:
			logs.Printf(" %s", val)
		}
		logs.Println()
	}
}

// Set edits and saves a setting within a configuration file.
func Set(name string) {
	if viper.ConfigFileUsed() == "" {
		configMissing(cmdPath, "set")
	}
	Config.nameFlag = name
	keys := viper.AllKeys()
	sort.Strings(keys)
	// sort.SearchStrings() - The slice must be sorted in ascending order.
	if i := sort.SearchStrings(keys, name); i == len(keys) || keys[i] != name {
		logs.Printf("%s\n%s %s\n",
			color.Warn.Sprintf("invalid flag value %v", fmt.Sprintf("--name %s", name)),
			fmt.Sprint("to see a list of usable settings run:"),
			color.Bold.Sprint("df2 config info"))
		os.Exit(202)
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

func configExists(name, suffix string) {
	cmd := strings.TrimSuffix(name, suffix)
	color.Warn.Println("a config file already is in use")
	logs.Printf("to edit:\t%s\n", cmd+"edit")
	logs.Printf("to remove:\t%s\n", cmd+"delete")
	os.Exit(20)
}

func configMissing(name, suffix string) {
	cmd := strings.TrimSuffix(name, suffix) + "create"
	color.Warn.Println("no config file is in use")
	logs.Printf("to create:\t%s\n", cmd)
	os.Exit(21)
}

func configSave(value interface{}) {
	switch value.(type) {
	case int64, string:
	default:
		logs.Check(fmt.Errorf("unsupported value interface type"))
	}
	viper.Set(Config.nameFlag, value)
	logs.Printf("%s %s is now set to \"%v\"\n", logs.Y(), color.Primary.Sprint(Config.nameFlag), color.Info.Sprint(value))
	writeConfig(true)
}

// writeConfig saves all configs to a configuration file.
func writeConfig(update bool) {
	bs, err := yaml.Marshal(viper.AllSettings())
	logs.Check(err)
	err = ioutil.WriteFile(Config.filepath, bs, 0600) // owner+wr
	logs.Check(err)
	s := "created a new"
	if update {
		s = "updated the"
	}
	logs.Println(s+" config file", Config.filepath)
}

func recommend(value string) string {
	return color.Info.Sprintf("(recommend: %v)", value)
}
