package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
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

var (
	configSetName string
	fileOverwrite bool
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure the settings for this tool",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 && cmd.Flags().NFlag() == 0 {
			_ = cmd.Usage()
			os.Exit(0)
		}
		_ = cmd.Usage()
		logs.Arg(args)
	},
}

var configCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new config file",
	Run: func(cmd *cobra.Command, args []string) {
		if cfg := viper.ConfigFileUsed(); cfg != "" && !fileOverwrite {
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
		switch logs.PromptYN("Confirm the file deletion", false) {
		case true:
			if err := os.Remove(cfg); err != nil {
				logs.Check(fmt.Errorf("Could not remove %v %v", cfg, err))
			}
			fmt.Println("The config is gone")
		}
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
			editors := []string{"micro", "nano", "vim"}
			for _, editor := range editors {
				if _, err := exec.LookPath(editor); err == nil {
					edit = editor
					break
				}
			}
			if edit != "" {
				fmt.Printf("There is no %s environment variable set so using: %s\n", "EDITOR", edit)
			}
		} else {
			edit = viper.GetString("editor")
			if _, err := exec.LookPath(edit); err != nil {
				logs.Check(fmt.Errorf("%v command not found %v", edit, exec.ErrNotFound))
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
		var name = configSetName
		keys := viper.AllKeys()
		sort.Strings(keys)
		// sort.SearchStrings() - The slice must be sorted in ascending order.
		if i := sort.SearchStrings(keys, name); i == len(keys) || keys[i] != name {
			err := fmt.Errorf("to see a list of usable settings, run: df2 config info")
			logs.Check(fmt.Errorf("invalid flag %v %v", fmt.Sprintf("--name %s", name), err))
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
	configCreateCmd.Flags().BoolVarP(&fileOverwrite, "overwrite", "y", false, "overwrite any existing config file")
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configInfoCmd)
	configCmd.AddCommand(configSetCmd)
	configSetCmd.Flags().StringVarP(&configSetName, "name", "n", "", `the configuration path to edit in dot syntax (see examples)
to see a list of names run: df2 config info`)
	_ = configSetCmd.MarkFlagRequired("name")
}

func configExists(name string, suffix string) {
	cmd := strings.TrimSuffix(name, suffix)
	fmt.Printf("A config file already is in use at: %s\n", viper.ConfigFileUsed())
	fmt.Printf("To edit it: %s\n", cmd+"edit")
	fmt.Printf("To delete:  %s\n", cmd+"delete")
	os.Exit(1)
}

func configMissing(name string, suffix string) {
	cmd := strings.TrimSuffix(name, suffix) + "create"
	fmt.Printf("No config file is in use.\nTo create one run: %s\n", cmd)
	os.Exit(1)
}

func configSave(value interface{}) {
	switch value.(type) {
	case int64, string:
	default:
		logs.Check(fmt.Errorf("unsupported value interface type"))
	}
	viper.Set(configSetName, value)
	fmt.Printf("%s %s is now set to \"%v\"\n", "✓", configSetName, value)
	writeConfig(true)
}

// writeConfig saves all configs to a configuration file.
func writeConfig(update bool) {
	bs, err := yaml.Marshal(viper.AllSettings())
	logs.Check(err)
	err = ioutil.WriteFile(filepath, bs, 0660)
	logs.Check(err)
	s := "Created a new"
	if update {
		s = "Updated the"
	}
	fmt.Println(s+" config file at:", filepath)
}
