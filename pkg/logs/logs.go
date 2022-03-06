// Package logs handles errors and user feedback.
package logs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/pkg/logs/internal/terminal"
	"github.com/gookit/color"
	gap "github.com/muesli/go-app-paths"
)

var ErrNoArg = errors.New("no arguments were provided")

const (
	GapUser  string = "df2"        // GapUser is the configuration and logs subdirectory name.
	Filename string = "errors.log" // Filename is the default error log filename.

	dmode   os.FileMode = 0o700
	fmode   os.FileMode = 0o600
	flags   int         = log.Ldate | log.Ltime | log.LUTC
	newmode int         = os.O_APPEND | os.O_CREATE | os.O_WRONLY
)

var (
	panicErr = false //nolint:gochecknoglobals
	quiet    = false //nolint:gochecknoglobals
)

// Panic enables or disables panicking when Danger is used.
func Panic(b bool) {
	panicErr = b
}

// Quiet stops most writing to the standard output.
func Quiet(b bool) {
	quiet = b
}

func IsQuiet() bool {
	return quiet
}

// Arg returns instructions for invalid command arguments and exits with an error code.
func Arg(arg string, exit bool, args ...string) error {
	if args == nil {
		return fmt.Errorf("arg: %w", ErrNoArg)
	}
	fmt.Printf("%s %s %s\n",
		color.Warn.Sprint("invalid command"),
		color.Bold.Sprintf("\"%s %s\"", arg, args[0]),
		color.Warn.Sprint("\nplease use one of the Available Commands shown above"))
	if !exit {
		return nil
	}
	os.Exit(1)
	return nil
}

// Danger logs the error to stdout, but continues the program.
func Danger(err error) {
	switch panicErr {
	case true:
		log.Println(fmt.Sprintf("error type %T\t: %v", err, err))
		log.Panic(err)
	default:
		log.Printf("%s %s", color.Danger.Sprint("!"), err)
	}
}

// Fatal logs error to stdout and exits with an error code.
func Fatal(err error) {
	save(Filename, err)
	switch panicErr {
	case true:
		log.Println(fmt.Sprintf("error type: %T\tmsg: %v", err, err))
		log.Panic(err)
	default:
		log.Fatal(color.Danger.Sprint("ERROR: "), err)
	}
}

// Filepath is the absolute path and filename of the error log file.
func Filepath(filename string) string {
	fp, err := gap.NewScope(gap.User, GapUser).LogPath(filename)
	if err != nil {
		h, err := os.UserHomeDir()
		if err != nil {
			Log(err)
		}
		return path.Join(h, filename)
	}
	return fp
}

// Log the error to stdout, but continue the program.
func Log(err error) {
	save(Filename, err)
	Danger(err)
}

// Path returns the named file or directory path with all missing elements marked in red.
func Path(name string) string {
	a := strings.Split(name, "/")
	var p, s string
	for i, e := range a {
		if e == "" {
			s = "/"
			continue
		}
		p = strings.Join(a[0:i+1], "/")
		if _, err := os.Stat(p); os.IsNotExist(err) {
			s = filepath.Join(s, color.Danger.Sprint(e))
		} else {
			s = filepath.Join(s, e)
		}
	}
	return fmt.Sprint(s)
}

// Print obeys the --quiet flag
// or formats using the default formats for its operands and writes to standard output.
func Print(a ...interface{}) {
	if quiet {
		return
	}
	if _, err := fmt.Print(a...); err != nil {
		fatalLog(err)
	}
}

// Printcr obeys the --quiet flag
// or otherwise erases the current line and writes to standard output.
func Printcr(a ...interface{}) {
	if quiet {
		return
	}
	if _, err := fmt.Printf("\r%s\r", strings.Repeat(" ", int(terminal.Size()))); err != nil {
		fatalLog(err)
	}
	if _, err := fmt.Print(a...); err != nil {
		fatalLog(err)
	}
}

// Printf obeys the --quiet flag
// or formats according to a format specifier and writes to standard output.
func Printf(format string, a ...interface{}) {
	if quiet {
		return
	}
	if _, err := fmt.Printf(format, a...); err != nil {
		fatalLog(err)
	}
}

// Println obeys the --quiet flag or formats using the default formats for its operands and writes to standard output.
func Println(a ...interface{}) {
	if quiet {
		return
	}
	if _, err := fmt.Println(a...); err != nil {
		fatalLog(err)
	}
}

// Printcrf obeys the --quiet flag
// or otherwise erases the current line and formats according to a format specifier.
func Printcrf(format string, a ...interface{}) {
	if quiet {
		return
	}
	if _, err := fmt.Printf("\r%s\r%s",
		strings.Repeat(" ", int(terminal.Size())),
		fmt.Sprintf(format, a...)); err != nil {
		fatalLog(err)
	}
}

// Simulate prints the --simulate=false flag info.
func Simulate() {
	Println(color.Notice.Sprint("use the --simulate=false flag to apply these fixes"))
}

func fatal(err error) {
	log.Printf("%s %s", color.Danger.Sprint("!"), err)
	os.Exit(1)
}

func fatalLog(err error) {
	save(Filename, err)
	fatal(err)
}

// save an error to the logs.
// path is available for unit tests.
func save(filename string, err error) (ok bool) {
	if err == nil {
		return false
	}
	// use UTC date and times in the log file
	log.SetFlags(flags)
	f := Filepath(filename)
	p := filepath.Dir(f)
	if _, err1 := os.Stat(p); os.IsNotExist(err1) {
		if err2 := os.MkdirAll(p, dmode); err != nil {
			fatal(err2)
		}
	} else if err1 != nil {
		fatal(err1)
	}
	file, err1 := os.OpenFile(f, newmode, fmode)
	if err1 != nil {
		fatal(err1)
	}
	defer file.Close()
	log.SetOutput(file)
	log.Print(err)
	log.SetOutput(os.Stderr)
	return true
}
