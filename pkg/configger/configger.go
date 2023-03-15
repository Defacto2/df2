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

const EnvPrefix = "DF2_"

// Config environment overrides for the Defacto2 tool.
type Config struct {
	DBName       string `env:"DBNAME,notEmpty" envDefault:"defacto2-inno" help:"Name of the database to use"`
	DBUser       string `env:"DBUSER,notEmpty" envDefault:"root" help:"Database connection user name"`
	DBPass       string `env:"DBPASS,notEmpty" envDefault:"example" help:"Database connection password"`
	DBHost       string `env:"DBHOST,notEmpty" envDefault:"localhost" help:"Database connection host address"`
	DBPort       uint   `env:"DBPORT,notEmpty" envDefault:"3306" help:"Database connection TCP port"`
	WebRoot      string `env:"ROOT,notEmpty" help:"Path to the root directory of the website"`
	Downloads    string `env:"DOWNLOAD" help:"Path containing UUID named files served as downloads"`
	Images       string `env:"IMG000" help:"Path containing screenshots and previews"`
	Thumbs       string `env:"IMG400" help:"Path containing 400x400 thumbnails of the screenshots"`
	Backups      string `env:"BACKUP" help:"Path containing backup archives or previously removed files"`
	Emulator     string `env:"EMULATOR" help:"Path containing the DOSee emulation files"`
	Timeout      uint   `env:"TIMEOUT" envDefault:"5" help:"The timeout value for remote TCP connections"`
	IsProduction bool   `env:"PRODUCTION" envDefault:"false" help:"Use the production mode to log all errors and warnings to a file"` //nolint:lll
	MaxProcs     uint   `env:"MAXPROCS" envDefault:"0" help:"Limit the number of operating system threads the program can use"`       //nolint:lll

	// TODO: add other missing directories??
}

func Defaults() Config {
	const root = string(os.PathSeparator)
	assets := filepath.Join(root, "opt", "assets-defacto2")
	webRoot := filepath.Join(root, "opt", "Defacto2-2020", "ROOT")
	return Config{
		WebRoot:   webRoot,
		Downloads: filepath.Join(assets, "downloads"),
		Images:    filepath.Join(assets, "images000"),
		Thumbs:    filepath.Join(assets, "images400"),
		Backups:   filepath.Join(webRoot, "files", "backups"),
		Emulator:  filepath.Join(webRoot, "files", "emularity.zip"),
	}
}

func Options() env.Options {
	return env.Options{
		Prefix: EnvPrefix,
		// Environment: map[string]string{
		// 	//"DF2_ROOT": filepath.Join("opt", "Defacto2-2020", "ROOT"),
		// },
	}
}

// TODO: move cmd/root/readIn to here and use the home path as envDefault?
// TODO: use the dir path colouriser to display paths
// TODO: database/internal/connect/connect.go (also remove func defaults())
// TODO: search for `viper` and replace all viper.GetString funcs etc.
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
