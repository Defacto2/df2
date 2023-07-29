// Package conf sets the configurations of this program using the host system
// environment variables.
package conf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/caarlos0/env/v7"
	"github.com/charmbracelet/lipgloss"
	"github.com/gookit/color"
)

const (
	EnvPrefix  = "DF2_"     // EnvPrefix is the prefix applied to all environment variable names.
	LiveServer = "DF2_HOST" // LiveServer is environment variable name to identify the live web server.
	GapUser    = "df2"      // GapUser is the Go Application Paths username.
)

// Config environment overrides for the Defacto2 tool.
// There are no envDefault attributes in this struct,
// instead they're found in the func Defaults().
type Config struct {
	IsProduction  bool   `env:"PRODUCTION" help:"Use the production mode to log all errors and warnings to a file"` //nolint:lll
	MaxProcs      uint   `env:"MAXPROCS" help:"Limit the number of operating system threads the program can use"`   //nolint:lll
	DBName        string `env:"DBNAME" help:"Name of the database to use"`
	DBUser        string `env:"DBUSER" help:"Database connection user name"`
	DBPass        string `env:"DBPASS" help:"Database connection password"`
	DBHost        string `env:"DBHOST" help:"Database connection host address"`
	DBPort        uint   `env:"DBPORT" help:"Database connection TCP port"`
	WebRoot       string `env:"ROOT" help:"Path to the root directory of the website"`
	Downloads     string `env:"DOWNLOAD" help:"Path containing UUID named files served as downloads"`
	Images        string `env:"IMG000" help:"Path containing screenshots and previews"`
	Thumbs        string `env:"IMG400" help:"Path containing 400x400 thumbnails of the screenshots"`
	Backups       string `env:"BACKUP" help:"Path containing backup archives or previously removed files"`
	Emulator      string `env:"EMULATOR" help:"Path containing the DOSee emulation files"`
	HTMLExports   string `env:"HTML" help:"Path to save the HTML files generated by this tool"`
	IncomingFiles string `env:"INCOMING" help:"Path containing user uploaded files"`
	IncomingImgs  string `env:"INCOMINGIMG" help:"Path containing screenshots of user uploaded files"`
	HTMLViews     string `env:"VIEWS" help:"Path to save the HTML files generated by this tool"`
	SQLDumps      string `env:"SQLDUMP" help:"Path containing database data exports as SQL dumps"`
	Timeout       uint   `env:"TIMEOUT" help:"The timeout in seconds value for database connections"`
}

// Defaults for the Config environment struct.
// Directory paths are different based on weather the DF2_HOST environment variable is set.
// When set, a /opt/ parent directory is used as the root, otherwise the user home directory is used.
func Defaults() Config {
	// remote server defaults
	const root = string(os.PathSeparator)
	opt := filepath.Join(root, "opt")
	assets := filepath.Join(opt, "assets-defacto2")
	webRoot := filepath.Join(opt, "Defacto2-2020", "ROOT")
	incoming := filepath.Join(webRoot, "incoming", "user_submissions")
	// local developer overrides
	value, ok := os.LookupEnv(LiveServer)
	if !ok || value == "" {
		home, _ := os.UserHomeDir()
		assets = filepath.Join(home, "assets-defacto2")
		incoming = filepath.Join(home, "user_submissions")
		opt = filepath.Join(home, "opt")
		webRoot = filepath.Join(home, "github", "Defacto2-2020", "ROOT")
	}

	const (
		mysqlPort  = 3306
		timeoutSec = 30
	)

	init := Config{
		// program settings
		IsProduction: false,
		MaxProcs:     0,
		// database connection
		DBName:  "defacto2-inno",
		DBUser:  "root",
		DBPass:  "password",
		DBHost:  "localhost",
		DBPort:  mysqlPort,
		Timeout: timeoutSec, // Timeout value matches the 30s timeout for unit tests.
		// directory paths
		WebRoot:       webRoot,
		Downloads:     filepath.Join(assets, "downloads"),
		Images:        filepath.Join(assets, "images000"),
		Thumbs:        filepath.Join(assets, "images400"),
		Backups:       filepath.Join(webRoot, "files", "backups"),
		Emulator:      filepath.Join(webRoot, "files", "emularity.zip"),
		HTMLExports:   filepath.Join(webRoot, "files", "html"),
		HTMLViews:     filepath.Join(webRoot, "views"),
		IncomingFiles: filepath.Join(incoming, "files"),
		IncomingImgs:  filepath.Join(incoming, "previews"),
		SQLDumps:      filepath.Join(opt, "backup"),
	}
	if ok && value != "" {
		init.DBHost = value
	}
	return init
}

// TestData returns the directory paths but with the temporary directory as root.
// This is intended for directories unit tests.
func TestData() Config {
	tmp := filepath.Join(os.TempDir(), "df2-mocker")
	assets := filepath.Join(tmp, "assets-defacto2")
	webRoot := filepath.Join(tmp, "github", "Defacto2-2020", "ROOT")
	incoming := filepath.Join(webRoot, "incoming", "user_submissions")
	return Config{
		// directory paths
		WebRoot:       webRoot,
		Downloads:     filepath.Join(assets, "downloads"),
		Images:        filepath.Join(assets, "images000"),
		Thumbs:        filepath.Join(assets, "images400"),
		Backups:       filepath.Join(webRoot, "files", "backups"),
		Emulator:      filepath.Join(webRoot, "files", "emularity.zip"),
		HTMLExports:   filepath.Join(webRoot, "files", "html"),
		HTMLViews:     filepath.Join(webRoot, "views"),
		IncomingFiles: filepath.Join(incoming, "files"),
		IncomingImgs:  filepath.Join(incoming, "previews"),
		SQLDumps:      filepath.Join(tmp, "backup"),
	}
}

func Options() env.Options {
	return env.Options{
		Prefix: EnvPrefix,
	}
}

// TODO: use the dir path colouriser to display paths
// TODO: replace dir../dir...go Init() func.

func (c Config) String(w io.Writer) { //nolint:funlen
	if w == nil {
		w = io.Discard
	}
	const (
		minwidth = 2
		tabwidth = 4
		padding  = 2
		padchar  = ' '
		flags    = 0
		h1       = "Variable"
		h2       = "Value"
		h3       = "Variable"
		h4       = "Value type"
		h5       = "Help"
		line     = "─"
		donotuse = 5
	)

	// TODO:
	// list configurations in use (ie non-blanks)
	// DBNAME defacto2-inno (italic, wrap HELP)
	// then list all environment variables, types and help

	tw := tabwriter.NewWriter(w, minwidth, tabwidth, padding, padchar, flags)
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("228")).
		PaddingLeft(padding).
		PaddingRight(padding).
		Margin(1)
	fmt.Fprintln(tw, style.Render("Environment variables and configurations"))

	fmt.Fprintf(tw, "\t%s\t%s\t\t\n", h1, h2)
	fmt.Fprintf(tw, "\t%s\t%s\t\t\n",
		strings.Repeat(line, len(h1)), strings.Repeat(line, len(h2)))

	fields := reflect.VisibleFields(reflect.TypeOf(c))
	values := reflect.ValueOf(c)
	for _, field := range fields {
		if !field.IsExported() {
			continue
		}
		val, def := values.FieldByName(field.Name), field.Tag.Get("envDefault")
		fmt.Fprintf(tw, "\t%s\t%v\t%v\t\n",
			EnvPrefix+field.Tag.Get("env"),
			val,
			match(fmt.Sprint(val), def),
		)
	}
	fmt.Fprintln(tw)
	tw.Flush()

	tw = tabwriter.NewWriter(w, minwidth, tabwidth, padding, padchar, flags)
	fmt.Fprintf(tw, "\t%s\t%s\t%s\n", h3, h4, h5)
	fmt.Fprintf(tw, "\t%s\t%s\t%s\n",
		strings.Repeat(line, len(h3)), strings.Repeat(line, len(h4)), strings.Repeat(line, len(h5)))
	for j, field := range fields {
		if !field.IsExported() {
			continue
		}
		if j == donotuse {
			fmt.Fprintf(tw, "\t\t\t\t\n")
			fmt.Fprintf(tw, "\t\t\t  These variables below are not recommended.\t\n")
		}
		fmt.Fprintf(tw, "\t%s\t%s\t",
			field.Tag.Get("env"),
			types(field.Type),
		)
		sp := ""
		if field.Tag.Get("avoid") != "" {
			sp = " "
		}
		fmt.Fprintf(tw, "%s%s%s.\n",
			avoid(field.Tag.Get("avoid")),
			sp,
			field.Tag.Get("help"),
		)
	}
	tw.Flush()
}

func avoid(x string) string {
	if x == "true" {
		c := color.New(color.FgRed, color.Bold)
		return c.Sprint("✗")
	}
	return ""
}

func match(x, y string) string {
	if x != y {
		c := color.New(color.FgGreen, color.Bold)
		return c.Sprint("✓")
	}
	return ""
}

func types(t reflect.Type) string {
	switch t.Kind() { //nolint:exhaustive
	case reflect.Bool:
		return "true|false"
	case reflect.Uint:
		return "number"
	default:
		return t.String()
	}
}
