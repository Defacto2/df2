package main

/*
Copyright © 2021-22 Ben Garrett <code.by.ben@gmail.com>

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

	"github.com/Defacto2/df2/pkg/cmd"
	"github.com/Defacto2/df2/pkg/database"
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
	if err := fs.Parse(os.Args[1:]); err != nil &&
		!errors.As(err, &pflag.ErrHelp) {
		log.Print(err)
	}
	if *ver || *v {
		fmt.Println(app())
		fmt.Println(info())
		return
	}
	// cobra flag lib
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func app() string {
	type Data struct {
		Version string
	}
	const verTmp = `
   Defacto2 / df2 tool version: {{.Version}}
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

func exeTmp() string {
	const tmp = ` ┌──── requirements ───────────┐
 │  database:     {{.Database}}  │
 │  ansilove:     {{.Ansilove}}  │
 │  webp lib:     {{.Webp}}  │
 │  imagemagick:  {{.Magick}}  │
 │  netpbm:       {{.Netpbm}}  │
 │  -------------              │
 │  arj:          {{.Arj}}  │
 │  unzip:        {{.UnZip}}  │
 │  zipinfo:      {{.ZipInfo}}  │
 └─────────────────────────────┘
     path: {{.Path}}
   commit: {{.Commit}}
     date: {{.Date}}
       go: v{{.GoVer}} {{.GoOS}}
`
	return tmp
}

func checks(s string) string {
	const (
		disconnect = "disconnect"
		ok         = "ok"
		miss       = "missing"
	)
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

func info() string {
	type Data struct {
		Database string
		Ansilove string
		Webp     string
		Magick   string
		Netpbm   string
		Arj      string
		UnZip    string
		ZipInfo  string
		Commit   string
		Date     string
		Path     string
		GoVer    string
		GoOS     string
	}
	const (
		disconnect = "disconnect"
		ok         = "ok"
		miss       = "missing"
	)
	var bin string
	d, a, w, m, n, ar, uz, zi :=
		disconnect, miss, miss, miss, miss, miss, miss, miss
	if err := database.ConnInfo(); err == "" {
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
	if _, err := exec.LookPath("pnmtopng"); err == nil {
		n = ok
	}
	if _, err := exec.LookPath("arj"); err == nil {
		ar = ok
	}
	if _, err := exec.LookPath("unzip"); err == nil {
		uz = ok
	}
	if _, err := exec.LookPath("zipinfo"); err == nil {
		zi = ok
	}
	bin, err := self()
	if err != nil {
		bin = fmt.Sprint(err)
	}

	data := Data{
		Database: checks(d),
		Ansilove: checks(a),
		Webp:     checks(w),
		Magick:   checks(m),
		Netpbm:   checks(n),
		Arj:      checks(ar),
		UnZip:    checks(uz),
		ZipInfo:  checks(zi),
		Commit:   commit,
		Date:     localBuild(date),
		Path:     bin,
		GoVer:    strings.Replace(runtime.Version(), "go", "", 1),
		GoOS:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
	tmpl, err := template.New("checks").Parse(exeTmp())
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
