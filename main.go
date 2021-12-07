package main

/*
Copyright © 2021 Ben Garrett <code.by.ben@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/Defacto2/df2/lib/cmd"
	"github.com/Defacto2/df2/lib/database"
	"github.com/gookit/color"
	"github.com/spf13/pflag"
)

// goreleaser generated ldflags containers.
var (
	version = "0.0.0"
	commit  = "unset" // nolint: gochecknoglobals
	date    = "unset" // nolint: gochecknoglobals
)

func main() {
	// go flag lib
	fs := flag.NewFlagSet("df2", flag.ContinueOnError)
	ver := fs.Bool("version", false, "version and information for this program")
	v := fs.Bool("v", false, "alias for version")
	fs.Usage = func() {
		// disable go flag help
	}
	if err := fs.Parse(os.Args[1:]); err != nil && !errors.As(err, &pflag.ErrHelp) {
		log.Print(err)
	}
	if *ver || *v {
		fmt.Println(app())
		fmt.Println(info())
		return
	}
	// cobra flag lib
	cmd.Execute()
}

func app() string {
	type Data struct {
		Version string
	}
	const verTmp = `
   df2 tool version: {{.Version}}
`
	data := Data{
		Version: color.Primary.Sprint(version),
	}
	tmpl, err := template.New("ver").Parse(verTmp)
	if err != nil {
		log.Fatal(err)
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		log.Fatal(err)
	}
	return b.String()
}

func info() string {
	type Data struct {
		Database string
		Ansilove string
		Webp     string
		Magick   string
		Commit   string
		Date     string
		Path     string
		GoVer    string
		GoOS     string
	}
	const exeTmp = ` ┌── requirements ─────────────┐
 │  database:     {{.Database}}  │
 │  ansilove:     {{.Ansilove}}  │
 │  webp lib:     {{.Webp}}  │
 │  imagemagick:  {{.Magick}}  │
 └─────────────────────────────┘
     path: {{.Path}}
   commit: {{.Commit}}
     date: {{.Date}}
       go: v{{.GoVer}} {{.GoOS}}
`
	const (
		disconnect = "disconnect"
		ok         = "ok"
		miss       = "missing"
	)
	p := func(s string) string {
		switch s {
		case ok:
			const padding = 9
			return color.Success.Sprint("okay") + strings.Repeat(" ", padding-len(s))
		case miss:
			const padding = 11
			return color.Error.Sprint(miss) + strings.Repeat(" ", padding-len(s))
		case disconnect:
			const padding = 11
			return color.Error.Sprint("disconnect") + strings.Repeat(" ", padding-len(s))
		}
		return ""
	}
	a, w, m, d, bin := miss, miss, miss, disconnect, ""
	if err := database.ConnectInfo(); err == "" {
		d = ok
	}
	if _, err := exec.LookPath("ansilove"); err == nil {
		a = ok
	}
	if _, err := exec.LookPath("cwebp"); err == nil {
		w = ok
	}
	if _, err := exec.LookPath("convert"); err == nil {
		m = ok
	}
	bin, err := self()
	if err != nil {
		bin = fmt.Sprint(err)
	}

	data := Data{
		Database: p(d),
		Ansilove: p(a),
		Webp:     p(w),
		Magick:   p(m),
		Commit:   commit,
		Date:     localBuild(date),
		Path:     bin,
		GoVer:    strings.Replace(runtime.Version(), "go", "", 1),
		GoOS:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
	tmpl, err := template.New("checks").Parse(exeTmp)
	if err != nil {
		log.Fatal(err)
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		log.Fatal(err)
	}
	return b.String()
}

func localBuild(date string) string {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return date
	}
	return t.Local().Format("2006 Jan 2, 15:04 MST")
}

func self() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("self error: %w", err)
	}
	return exe, nil
}
