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
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/Defacto2/df2/cmd"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/carlmjohnson/versioninfo"
	"github.com/gookit/color"
)

//go:embed .version
var version string

func main() {
	// terminal stderr and stdout configurations
	rmLogTimestamp()
	if ascii() {
		color.Enable = false
	}
	// print the compile and version details
	if ver() {
		fmt.Println(info())
		return
	}
	// suppress stdout except when requesting help
	if quiet() && !help() {
		os.Stdout, _ = os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		defer os.Stdout.Close()
	}
	// cobra flag library
	if err := cmd.Execute(); err != nil {
		defer os.Exit(1)
	}
}

// global flags that should not be handled by the Cobra library
// to keep things simple, avoid using the flag standard library

func ascii() bool {
	for _, f := range os.Args {
		switch strings.ToLower(f) {
		case "-ascii", "--ascii":
			return true
		}
	}
	return false
}

func help() bool {
	for _, f := range os.Args {
		switch strings.ToLower(f) {
		case "-h", "--h", "-help", "--help":
			return true
		}
	}
	return false
}

func quiet() bool {
	for _, f := range os.Args {
		switch strings.ToLower(f) {
		case "-quiet", "--quiet":
			return true
		}
	}
	return false
}

func ver() bool {
	for _, f := range os.Args {
		switch strings.ToLower(f) {
		case "-v", "--v", "-version", "--version":
			return true
		}
	}
	return false
}

// rmLogTimestamp removes the timestamp prefixed to the print functions of the log lib.
func rmLogTimestamp() {
	// source: https://stackoverflow.com/questions/48629988/remove-timestamp-prefix-from-go-logger
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
}

func exeTmp() string {
	const tmp = `
 ┬───── {{.Title}} ─────┬─────────────────────────────┬
 │                             │                             │
 │  requirements               │   recommended               │
 │                             │                             │
 │      database  {{.Database}}  │           arj  {{.Arj}}  │
 │                             │    file magic  {{.File}}  │
 │      ansilove  {{.Ansilove}}  │         lhasa  {{.Lha}}  │
 │      webp lib  {{.Webp}}  │         unrar  {{.UnRar}}  │
 │   imagemagick  {{.Magick}}  │         unzip  {{.UnZip}}  │
 │        netpbm  {{.Netpbm}}  │       zipinfo  {{.ZipInfo}}  │
 │      pngquant  {{.PngQuant}}  │                             │
 │                             │                             │
 ┴─────────────────────────────┴─────────────────── {{.Cmd}} ─────┴
         version  {{.Version}}
            path  {{.Path}}
          commit  {{.Revision}}
            date  {{.LastCommit}}
              go  v{{.GoVer}} {{.GoOS}}{{.Docker}}
`
	return tmp
}

func check(s string) string {
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

type looks = map[string]string

func checks() looks {
	const (
		disconnect = "disconnect"
		ok         = "ok"
		miss       = "missing"
		db         = "db"
	)
	l := looks{
		"db":       disconnect,
		"ansilove": miss,
		"cwebp":    miss,
		"convert":  miss,
		"pnmtopng": miss,
		"pngquant": miss,
		"arj":      miss,
		"file":     miss,
		"lha":      miss,
		"unrar":    miss,
		"unzip":    miss,
		"zipinfo":  miss,
	}
	if err := database.ConnInfo(); err == "" {
		l[db] = ok
	}
	for file := range l {
		if file == db {
			continue
		}
		if _, err := exec.LookPath(file); err == nil {
			l[file] = ok
		}
	}
	return l
}

func info() string {
	type Data struct {
		Database   string
		Ansilove   string
		Webp       string
		Magick     string
		Netpbm     string
		PngQuant   string
		Arj        string
		File       string
		Lha        string
		UnRar      string
		UnZip      string
		ZipInfo    string
		Version    string
		Revision   string
		LastCommit string
		Path       string
		GoVer      string
		GoOS       string
		Docker     string
		Title      string
		Cmd        string
	}
	bin, err := selfBuild()
	if err != nil {
		bin = fmt.Sprint(err)
	}
	l := checks()
	data := Data{
		Database:   check(l["db"]),
		Ansilove:   check(l["ansilove"]),
		Webp:       check(l["cwebp"]),
		Magick:     check(l["convert"]),
		Netpbm:     check(l["pnmtopng"]),
		PngQuant:   check(l["pngquant"]),
		Arj:        check(l["arj"]),
		File:       check(l["file"]),
		Lha:        check(l["lha"]),
		UnRar:      check(l["unrar"]),
		UnZip:      check(l["unzip"]),
		ZipInfo:    check(l["zipinfo"]),
		Version:    tagVersion(),
		Revision:   versioninfo.Revision,
		LastCommit: localBuild(versioninfo.LastCommit),
		Path:       bin,
		GoVer:      strings.Replace(runtime.Version(), "go", "", 1),
		GoOS:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Docker:     dockerBuild(),
		Title:      color.Bold.Sprint(color.Primary.Sprint("The Defacto2 tool")),
		Cmd:        color.Primary.Sprint("df2"),
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

func dockerBuild() string {
	if _, ok := os.LookupEnv("DF2_HOST"); ok {
		return " in a Docker container"
	}
	return ""
}

func localBuild(t time.Time) string {
	const unknown = 1
	if t.Local().Year() <= unknown {
		return "unknown"
	}
	return t.Local().Format("2006 Jan 2, 15:04 MST")
}

func selfBuild() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("self error: %w", err)
	}
	return exe, nil
}

func tagVersion() string {
	s := color.Style{color.FgCyan, color.OpBold}.Sprint(version)
	if version == "0.0.0" {
		s += " (developer build)"
	}
	return s
}
