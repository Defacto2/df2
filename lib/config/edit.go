package config

import (
	"log"
	"os"
	"os/exec"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/viper"
)

// Edit a configuration file.
func Edit() {
	cfg := viper.ConfigFileUsed()
	if cfg == "" {
		configMissing("edit")
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
			editNotFound()
		}
	} else {
		edit = viper.GetString("editor")
		if _, err := exec.LookPath(edit); err != nil {
			if edit != "" {
				log.Printf("%q edit command not found\n%v", edit, exec.ErrNotFound)
				os.Exit(1)
			} else {
				editNotFound()
			}
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

func editNotFound() {
	log.Println("no suitable editor could be found\nplease set one by creating a $EDITOR environment variable in your shell configuration")
	os.Exit(1)
}
