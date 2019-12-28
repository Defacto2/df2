package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	// portMin is the lowest permitted network port
	portMin int = 0
	// portMax is the largest permitted network port
	portMax int = 65535
)

var (
	configSetName string
	fileOverwrite bool
	infoStyles    string
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
		logs.Check(fmt.Errorf("invalid command %v please use one of the available config commands", args[0]))
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
		switch promptYN("Confirm the file deletion", false) {
		case true:
			if err := os.Remove(cfg); err != nil {
				logs.Check(fmt.Errorf("Could not remove %v %v", cfg, err))
			}
			fmt.Println("The file is deleted")
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
		fmt.Println("config file:", filepath)
		fmt.Println(string(sets))
		println()
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Change a configuration",
	//todo add long with information on how to view settings
	Example: `--name create.meta.description # to change the meta description setting
--name version.format          # to set the version command output format`,
	Run: func(cmd *cobra.Command, args []string) {
		var name = configSetName
		keys := viper.AllKeys()
		sort.Strings(keys)
		// sort.SearchStrings() - The slice must be sorted in ascending order.
		if i := sort.SearchStrings(keys, name); i == len(keys) || keys[i] != name {
			err := fmt.Errorf("to see a list of usable settings, run: retrotxt config info")
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
			fmt.Printf("\nSet a new host, leave blank to keep as-is (recommended: localhost): \n")
			promptString(s)
		case name == "connection.server.protocol":
			fmt.Printf("\nSet a new protocol, leave blank to keep as-is (recommended: tcp): \n")
			promptString(s)
		case name == "connection.server.port":
			fmt.Printf("Set a new MySQL port, choices: %v-%v (recommended: 3306)\n", portMin, portMax)
			promptPort()
		case name[:10] == "directory.":
			fmt.Printf("\nSet a new directory or leave blank to keep as-is: \n")
			promptDir()
		case name == "connection.password":
			fmt.Printf("\nSet a new MySQL user encrypted or plaintext password or leave blank to keep as-is: \n")
			promptString(s)
		default:
			fmt.Printf("\nSet a new value, leave blank to keep as-is or use a dash [-] to disable: \n")
			promptString(s)
		}
	},
}

// InitDefaults initialises flag and configuration defaults.
func InitDefaults() {
	viper.SetDefault("connection.name", "defacto2-inno")
	viper.SetDefault("connection.user", "root")
	viper.SetDefault("connection.password", "password")
	viper.SetDefault("connection.server.protocol", "tcp")
	viper.SetDefault("connection.server.host", "localhost")
	viper.SetDefault("connection.server.port", "3306")

	viper.SetDefault("directory.root", "/var/www")
	viper.SetDefault("directory.backup", "/var/www/files/backups")
	viper.SetDefault("directory.emu", "/var/www/uuid/emularity.zip")
	viper.SetDefault("directory.uuid", "/var/www/uuid")
	viper.SetDefault("directory.000", "/var/www/images/uuid/original")
	viper.SetDefault("directory.150", "/var/www/images/uuid/150x150")
	viper.SetDefault("directory.400", "/var/www/images/uuid/400x400")
	viper.SetDefault("directory.incoming.files", "/var/www/incoming/user_submissions/files")
	viper.SetDefault("directory.incoming.previews", "/var/www/incoming/user_submissions/previews")
}

func init() {
	InitDefaults()
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configCreateCmd)
	configCreateCmd.Flags().BoolVarP(&fileOverwrite, "overwrite", "y", false, "overwrite any existing config file")
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configInfoCmd)
	configInfoCmd.Flags().StringVarP(&infoStyles, "syntax-style", "c", "monokai", "config syntax highligher, \"use none\" to disable")
	configCmd.AddCommand(configSetCmd)
	configSetCmd.Flags().StringVarP(&configSetName, "name", "n", "", `the configuration path to edit in dot syntax (see examples)
to see a list of names run: retrotxt config info`)
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

func promptCheck(cnt int) {
	switch {
	case cnt == 2:
		fmt.Println("Ctrl+C to keep the existing port")
	case cnt >= 4:
		os.Exit(1)
	}
}

func promptDir() {
	// allow multiple word user input
	scanner := bufio.NewScanner(os.Stdin)
	var save string
	for scanner.Scan() {
		txt := scanner.Text()
		switch txt {
		case "":
			os.Exit(0)
		case "-":
			save = ""
		default:
			save = txt
		}
		if _, err := os.Stat(save); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "will not save the change:", err)
			os.Exit(1)
		}
		viper.Set(configSetName, save)
		fmt.Printf("%s %s is now set to \"%v\"\n", "✓", configSetName, save)
		writeConfig(true)
		os.Exit(0)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}
}

func promptPort() {
	var input string
	cnt := 0
	for {
		input = ""
		cnt++
		fmt.Scanln(&input)
		if input == "" {
			promptCheck(cnt)
			continue
		}
		i, err := strconv.ParseInt(input, 10, 0)
		if err != nil && input != "" {
			fmt.Printf("%s %v\n", "✗", input)
			promptCheck(cnt)
			continue
		}
		// check that the input a valid port
		if v := validPort(int(i)); !v {
			fmt.Printf("%s %v, is out of range\n", "✗", input)
			promptCheck(cnt)
			continue
		}
		viper.Set(configSetName, i)
		fmt.Printf("%s %s is now set to \"%v\"\n", "✓", configSetName, i)
		writeConfig(true)
		os.Exit(0)
	}
}

func promptString(keep string) {
	// allow multiple word user input
	scanner := bufio.NewScanner(os.Stdin)
	var save string
	for scanner.Scan() {
		txt := scanner.Text()
		switch txt {
		case "":
			os.Exit(0)
		case "-":
			save = ""
		default:
			save = txt
		}
		viper.Set(configSetName, save)
		fmt.Printf("%s %s is now set to \"%v\"\n", "✓", configSetName, save)
		writeConfig(true)
		os.Exit(0)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}
}

func promptYN(query string, yesDefault bool) bool {
	var input string
	y := "Y"
	n := "n"
	if !yesDefault {
		y = "y"
		n = "N"
	}
	fmt.Printf("%s? [%s/%s] ", query, y, n)
	fmt.Scanln(&input)
	switch input {
	case "":
		if yesDefault {
			return true
		}
	case "yes", "y":
		return true
	}
	return false
}

func writeConfig(update bool) {
	bs, err := yaml.Marshal(viper.AllSettings())
	logs.Check(err)
	logs.Check(err)
	err = ioutil.WriteFile(filepath, bs, 0660)
	logs.Check(err)
	s := "Created a new"
	if update {
		s = "Updated the"
	}
	fmt.Println(s+" config file at:", filepath)
}

func validPort(p int) bool {
	if p < portMin || p > portMax {
		return false
	}
	return true
}
