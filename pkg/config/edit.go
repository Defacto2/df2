package config

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/viper"
)

// ErrNoEditor no editor error.
var ErrNoEditor = errors.New(`no suitable editor could be found
please set one by creating a $EDITOR environment variable in your shell configuration`)

// Edit a configuration file.
func Edit(w io.Writer) {
	var editor string
	editors := []string{"micro", "nano", "vim"}
	cfg := viper.ConfigFileUsed()
	if cfg == "" {
		missing(w, "edit")
	}
	if err := viper.BindEnv("editor", "EDITOR"); err != nil {
		editor = fallback(editors...)
	} else {
		editor = saved()
	}
	// credit: https://stackoverflow.com/questions/21513321/how-to-start-vim-from-go
	exe := exec.Command(editor, cfg)
	exe.Stdin = os.Stdin
	exe.Stdout = os.Stdout
	if err := exe.Run(); err != nil {
		fmt.Fprintf(w, "%s\n", err)
	}
}

func fallback(editors ...string) string {
	path := ""
	for _, app := range editors {
		if path, err := exec.LookPath(app); err == nil && path != "" {
			return app
		}
	}
	if path != "" {
		log.Printf("there is no $EDITOR environment variable set so using %s\n", path)
		return ""
	}
	log.Print(ErrNoEditor)
	os.Exit(1)
	return ""
}

func saved() string {
	editor := viper.GetString("editor")
	if _, err := exec.LookPath(editor); err != nil {
		if editor != "" {
			log.Printf("%q edit command not found\n%v", editor, exec.ErrNotFound)
		} else {
			log.Print(ErrNoEditor)
		}
		os.Exit(1)
	}
	return editor
}
