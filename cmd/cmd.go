package cmd

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/carlmjohnson/versioninfo"
	"github.com/gookit/color"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ProgData struct {
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
	Platform   string
	GoVer      string
	GoOS       string
	Docker     string
	Title      string
	Cmd        string
}

const (
	About   = "A tool the optimise and manage " + Domain
	Author  = "Ben Garrett"                     // Author is the primary programmer of this program.
	Domain  = "defacto2.net"                    // Domain of the website.
	Program = "df2"                             // Program command.
	Title   = "The Defacto2 tool"               // Title of this program.
	URL     = "https://github.com/Defacto2/df2" // URL of the program repository.

	unknown = "unknown"
)

// Arch returns this program's system architecture.
func Arch() string {
	switch strings.ToLower(runtime.GOARCH) {
	case "amd64":
		return "Intel/AMD 64"
	case "arm":
		return "ARM 32"
	case "arm64":
		return "ARM 64"
	case "i386":
		return "x86"
	case "wasm":
		return "WebAssembly"
	}
	return runtime.GOARCH
}

// Brand prints the byte ASCII logo to the stdout.
func Brand(log *zap.SugaredLogger, b []byte) {
	if logo := string(b); len(logo) > 0 {
		w := bufio.NewWriter(os.Stdout)
		if _, err := fmt.Fprintf(w, "%s\n\n", logo); err != nil {
			log.Warnf("Could not print the brand logo: %s.", err)
		}
		w.Flush()
	}
}

// Commit returns a formatted, git commit description for this repository,
// including tag version and date.
func Commit(version string) string {
	s := ""
	c := versioninfo.Short()

	if version != "" {
		s = fmt.Sprintf("%s ", Vers(version))
	} else if c != "" {
		s += fmt.Sprintf("version %s, ", c)
	}
	if s == "" {
		return unknown
	}
	return strings.TrimSpace(s)
}

// Copyright returns the copyright years and author of this program.
func Copyright() string {
	const initYear = 2020
	t := versioninfo.LastCommit
	if t.Year() < initYear {
		t = time.Now()
	}
	s := fmt.Sprintf("© %d", initYear)
	if t.Year() > initYear {
		s += "-" + t.Local().Format("06")
	}
	s += fmt.Sprintf(" Defacto2 & %s", Author)
	return s
}

// ExeTmpl returns the template for the -version flag.
func ExeTmpl() string {
	const tmpl = `
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
	return tmpl
}

func LastCommit() string {
	d := versioninfo.LastCommit
	if d.IsZero() {
		return unknown
	}
	return d.Local().Format("2006 Jan 2 15:04")
}

// OS returns this program's operating system.
func OS() string {
	t := cases.Title(language.English)
	os := strings.Split(runtime.GOOS, "/")[0]
	switch os {
	case "darwin":
		return "macOS"
	case "freebsd":
		return "FreeBSD"
	case "js":
		return "JS"
	case "netbsd":
		return "NetBSD"
	case "openbsd":
		return "OpenBSD"
	}
	return t.String(os)
}

// Vers returns a formatted version.
// The version string is generated by GoReleaser.
func Vers(version string) string {
	const alpha, beta, dev = "\u03b1", "β", "0.0.0"
	if version == "" || version == dev {
		return fmt.Sprintf("v%s %slpha (developer build)", dev, alpha)
	}
	const next = "-next"
	if strings.HasSuffix(version, next) {
		return fmt.Sprintf("v%s %seta", strings.TrimSuffix(version, next), beta)
	}
	return version
}

// ProgInfo returns the response for the -version flag.
func ProgInfo(db *sql.DB, version string) (string, error) {
	bin, err := configger.BinPath()
	if err != nil {
		bin = fmt.Sprint(err)
	}
	l := check(db)
	data := ProgData{
		Database:   colorize(l["db"]),
		Ansilove:   colorize(l["ansilove"]),
		Webp:       colorize(l["cwebp"]),
		Magick:     colorize(l["convert"]),
		Netpbm:     colorize(l["pnmtopng"]),
		PngQuant:   colorize(l["pngquant"]),
		Arj:        colorize(l["arj"]),
		File:       colorize(l["file"]),
		Lha:        colorize(l["lha"]),
		UnRar:      colorize(l["unrar"]),
		UnZip:      colorize(l["unzip"]),
		ZipInfo:    colorize(l["zipinfo"]),
		Version:    Commit(version),
		Revision:   versioninfo.Revision,
		LastCommit: LastCommit(),
		Path:       bin,
		GoVer:      strings.Replace(runtime.Version(), "go", "", 1),
		GoOS:       OS(),
		Docker:     dockerBuild(),
		Title:      color.Bold.Sprint(color.Primary.Sprint(Title)),
		Cmd:        color.Primary.Sprint("df2"),
	}
	tmpl, err := template.New("checks").Parse(ExeTmpl())
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

type lookups = map[string]string

// check looks up the collection of dependencies and database connection.
func check(db *sql.DB) lookups {
	const (
		disconnect = "disconnect"
		ok         = "ok"
		miss       = "missing"
		d          = "db"
	)
	l := lookups{
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
	if err := database.ConnInfo(db, cfg); err == "" {
		l[d] = ok
	}
	for file := range l {
		if file == d {
			continue
		}
		if _, err := exec.LookPath(file); err == nil {
			l[file] = ok
		}
	}
	return l
}

// colorize applies color to s.
func colorize(s string) string {
	const (
		disconnect = "disconnect"
		ok         = "ok"
		miss       = "missing"
	)
	padding := 11
	switch s {
	case ok:
		padding = 9
		return color.Success.Sprint("okay") + strings.Repeat(" ", padding-len(s))
	case miss:
		return color.Error.Sprint(miss) + strings.Repeat(" ", padding-len(s))
	case disconnect:
		return color.Error.Sprint("disconnect") + strings.Repeat(" ", padding-len(s))
	}
	return ""
}

func dockerBuild() string {
	if _, ok := os.LookupEnv("DF2_HOST"); ok {
		return " in a Docker container"
	}
	return ""
}
