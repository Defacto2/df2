package cmd

// os.Exit() = 200+

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var cfgOWFlag bool

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure the settings for this tool",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Usage()
		logs.Check(err)
		if len(args) != 0 || cmd.Flags().NFlag() != 0 {
			logs.Arg("config", args)
		}
	},
}

var configCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new config file",
	Run: func(cmd *cobra.Command, args []string) {
		config.ignore = true
		if cfg := viper.ConfigFileUsed(); cfg != "" && !cfgOWFlag {
			configExists(cmd.CommandPath(), "create")
		}
		writeConfig(false)
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove the config file",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := viper.ConfigFileUsed()
		if cfg == "" {
			configMissing(cmd.CommandPath(), "delete")
		}
		if _, err := os.Stat(cfg); os.IsNotExist(err) {
			configMissing(cmd.CommandPath(), "delete")
		}
		switch logs.PromptYN("Remove the config file", false) {
		case true:
			if err := os.Remove(cfg); err != nil {
				logs.Check(fmt.Errorf("could not remove %v %v", cfg, err))
			}
			fmt.Println("the config is gone")
		}
		os.Exit(0)
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit the config file",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := viper.ConfigFileUsed()
		if cfg == "" {
			configMissing(cmd.CommandPath(), "edit")
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
			fmt.Printf("%s\n", err)
		}
	},
}

var configInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "View settings configured by the config",
	Run: func(cmd *cobra.Command, args []string) {
		println("These are the default configurations used by this tool when no flags are given.\n")
		sets, err := yaml.Marshal(viper.AllSettings())
		logs.Check(err)
		fmt.Printf("%v%v %v\n", color.Cyan.Sprint("config file"), color.Red.Sprint(":"), filepath)
		configErrCheck()
		scanner := bufio.NewScanner(strings.NewReader(string(sets)))
		for scanner.Scan() {
			s := strings.Split(scanner.Text(), ":")
			color.Cyan.Print(s[0])
			color.Red.Print(":")
			if len(s) > 1 {
				fmt.Printf("%s", strings.Join(s[1:], ""))
			}
			fmt.Println()
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Change a configuration",
	//todo add long with information on how to view settings
	Example: `--name connection.server.host # to change the database host setting
--name directory.000          # to set the image preview directory`,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.ConfigFileUsed() == "" {
			configMissing(cmd.CommandPath(), "set")
		}
		var name = config.nameFlag
		keys := viper.AllKeys()
		sort.Strings(keys)
		// sort.SearchStrings() - The slice must be sorted in ascending order.
		if i := sort.SearchStrings(keys, name); i == len(keys) || keys[i] != name {
			fmt.Printf("%s\n%s %s\n",
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
			fmt.Printf("\n%s is currently set to %q\n", name, s)
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
	},
}

func recommend(value string) string {
	return color.Info.Sprintf("(recommend: %v)", value)
}

// InitDefaults initialises flag and configuration defaults.
func InitDefaults() {
	viper.SetDefault("connection.name", "defacto2-inno")
	viper.SetDefault("connection.user", "root")
	viper.SetDefault("connection.password", "password")

	viper.SetDefault("connection.server.protocol", "tcp")
	viper.SetDefault("connection.server.host", "localhost")
	viper.SetDefault("connection.server.port", "3306")

	viper.SetDefault("directory.root", "/opt/assets")
	viper.SetDefault("directory.backup", "/opt/assets/backups")
	viper.SetDefault("directory.emu", "/opt/assets/emularity.zip")
	viper.SetDefault("directory.uuid", "/opt/assets/downloads")
	viper.SetDefault("directory.000", "/opt/assets/000")
	viper.SetDefault("directory.150", "/opt/assets/150")
	viper.SetDefault("directory.400", "/opt/assets/400")
	viper.SetDefault("directory.html", "/opt/assets/html")
	viper.SetDefault("directory.incoming.files", "/opt/incoming/files")
	viper.SetDefault("directory.incoming.previews", "/opt/incoming/previews")
}

func init() {
	InitDefaults()
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configCreateCmd)
	configCreateCmd.Flags().BoolVarP(&cfgOWFlag, "overwrite", "y", false, "overwrite any existing config file")
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configInfoCmd)
	configCmd.AddCommand(configSetCmd)
	configSetCmd.Flags().StringVarP(&config.nameFlag, "name", "n", "", `the configuration path to edit in dot syntax (see examples)
to see a list of names run: df2 config info`)
	_ = configSetCmd.MarkFlagRequired("name")
}

func configExists(name, suffix string) {
	cmd := strings.TrimSuffix(name, suffix)
	color.Warn.Println("a config file already is in use")
	fmt.Printf("to edit:\t%s\n", cmd+"edit")
	fmt.Printf("to remove:\t%s\n", cmd+"delete")
	os.Exit(20)
}

func configMissing(name, suffix string) {
	cmd := strings.TrimSuffix(name, suffix) + "create"
	color.Warn.Println("no config file is in use")
	fmt.Printf("to create:\t%s\n", cmd)
	os.Exit(21)
}

func configSave(value interface{}) {
	switch value.(type) {
	case int64, string:
	default:
		logs.Check(fmt.Errorf("unsupported value interface type"))
	}
	viper.Set(config.nameFlag, value)
	fmt.Printf("%s %s is now set to \"%v\"\n", logs.Y(), color.Primary.Sprint(config.nameFlag), color.Info.Sprint(value))
	writeConfig(true)
}

// writeConfig saves all configs to a configuration file.
func writeConfig(update bool) {
	bs, err := yaml.Marshal(viper.AllSettings())
	logs.Check(err)
	err = ioutil.WriteFile(filepath, bs, 0600) // owner+wr
	logs.Check(err)
	s := "created a new"
	if update {
		s = "updated the"
	}
	fmt.Println(s+" config file", filepath)
}
