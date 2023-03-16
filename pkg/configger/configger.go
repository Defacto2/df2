package configger

import (
	"fmt"
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
	EnvPrefix = "DF2_"
	GapUser   = "df2"
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
	assets := filepath.Join(root, "opt", "assets-defacto2")
	webRoot := filepath.Join(root, "opt", "Defacto2-2020", "ROOT")
	// local developer defaults
	value, ok := os.LookupEnv("DF2_HOST")
	if !ok || value == "" {
		home, _ := os.UserHomeDir()
		assets = filepath.Join(home, "assets-defacto2")
		webRoot = filepath.Join(home, "github", "Defacto2-2020", "ROOT")
	}
	// shared defaults
	incoming := filepath.Join(webRoot, "incoming", "user_submissions")

	init := Config{
		// program settings
		IsProduction: false,
		MaxProcs:     0,
		// database connection
		DBName:  "defacto2-inno",
		DBUser:  "root",
		DBPass:  "example",
		DBHost:  "localhost",
		DBPort:  3306,
		Timeout: 5,
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
		SQLDumps:      filepath.Join(root, "backup"),
	}
	if ok && value != "" {
		init.DBHost = value
	}
	return init
}

func Options() env.Options {
	return env.Options{
		Prefix: EnvPrefix,
	}
}

// TODO: use the dir path colouriser to display paths
// TODO: replace dir../dir...go Init() func.

func (c Config) String() string { //nolint:funlen
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

	b := new(strings.Builder)
	w := tabwriter.NewWriter(b, minwidth, tabwidth, padding, padchar, flags)

	var style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("228")).
		PaddingLeft(2).
		PaddingRight(2).
		Margin(1)
	fmt.Fprintln(w, style.Render("Environment variables and configurations"))

	fmt.Fprintf(w, "\t%s\t%s\t\t\n", h1, h2)
	fmt.Fprintf(w, "\t%s\t%s\t\t\n",
		strings.Repeat(line, len(h1)), strings.Repeat(line, len(h2)))

	fields := reflect.VisibleFields(reflect.TypeOf(c))
	values := reflect.ValueOf(c)
	for _, field := range fields {
		if !field.IsExported() {
			continue
		}
		val, def := values.FieldByName(field.Name), field.Tag.Get("envDefault")
		fmt.Fprintf(w, "\t%s\t%v\t%v\t\n",
			EnvPrefix+field.Tag.Get("env"),
			val,
			match(fmt.Sprint(val), def),
		)
	}
	fmt.Fprintln(w)
	w.Flush()

	w = tabwriter.NewWriter(b, minwidth, tabwidth, padding, padchar, flags)
	fmt.Fprintf(w, "\t%s\t%s\t%s\n", h3, h4, h5)
	fmt.Fprintf(w, "\t%s\t%s\t%s\n",
		strings.Repeat(line, len(h3)), strings.Repeat(line, len(h4)), strings.Repeat(line, len(h5)))
	for j, field := range fields {
		if !field.IsExported() {
			continue
		}
		if j == donotuse {
			fmt.Fprintf(w, "\t\t\t\t\n")
			fmt.Fprintf(w, "\t\t\t  These variables below are not recommended.\t\n")
		}
		fmt.Fprintf(w, "\t%s\t%s\t",
			field.Tag.Get("env"),
			types(field.Type),
		)
		sp := ""
		if field.Tag.Get("avoid") != "" {
			sp = " "
		}
		fmt.Fprintf(w, "%s%s%s.\n",
			avoid(field.Tag.Get("avoid")),
			sp,
			field.Tag.Get("help"),
		)
	}
	w.Flush()

	return b.String()
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
