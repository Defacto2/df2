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

var ErrNoCmd = errors.New("no command argument was provided")

const (
	GapUser  string = "df2"        // GapUser is the configuration and logs subdirectory name.
	Filename string = "errors.log" // Filename is the default error log filename.

	dmode   os.FileMode = 0o700
	fmode   os.FileMode = 0o600
	flags   int         = log.Ldate | log.Ltime | log.LUTC
	newmode int         = os.O_APPEND | os.O_CREATE | os.O_WRONLY
)

var panicErr = false //nolint:gochecknoglobals

// Panic enables or disables panicking when Danger is used.
func Panic(b bool) {
	panicErr = b
}

// Arg returns instructions for invalid command arguments and exits with an error code.
func Arg(arg string, exit bool, args ...string) error {
	if arg == "" {
		return ErrNoCmd
	}
	s := ""
	if len(args) == 0 {
		s = fmt.Sprintf("%s %s", color.Warn.Sprint("invalid command"),
			color.Bold.Sprintf("\"%s\"", arg))
	}
	if len(args) > 0 {
		s = fmt.Sprintf("%s %s",
			color.Warn.Sprint("invalid command"),
			color.Bold.Sprintf("\"%s %s\"", arg, args[0]))
	}
	s += fmt.Sprint("\n" + color.Warn.Sprint("please use one of the Available Commands shown above"))
	log.Println(s)
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
		log.Printf("error type %T\t: %v\n", err, err)
		log.Panic(err)
	default:
		log.Printf("%s %s", color.Danger.Sprint("!"), err)
	}
}

// Fatal logs error to stdout and exits with an error code.
func Fatal(err error) {
	Save(Filename, err)
	switch panicErr {
	case true:
		log.Printf("error type: %T\tmsg: %v\n", err, err)
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
	Save(Filename, err)
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

// Print formats using the default formats for its operands and writes to standard output.
func Print(a ...any) {
	if _, err := fmt.Fprint(os.Stdout, a...); err != nil {
		fatalLog(err)
	}
}

// Printcr otherwise erases the current line and writes to standard output.
func Printcr(a ...any) {
	if _, err := fmt.Fprintf(os.Stdout, "\r%s\r", strings.Repeat(" ", int(terminal.Size()))); err != nil {
		fatalLog(err)
	}
	if _, err := fmt.Fprint(os.Stdout, a...); err != nil {
		fatalLog(err)
	}
}

// Printf formats according to a format specifier and writes to standard output.
func Printf(format string, a ...any) {
	if _, err := fmt.Fprintf(os.Stdout, format, a...); err != nil {
		fatalLog(err)
	}
}

// Println formats using the default formats for its operands and writes to standard output.
func Println(a ...any) {
	if _, err := fmt.Fprintln(os.Stdout, a...); err != nil {
		fatalLog(err)
	}
}

// Printcrf erases the current line and formats according to a format specifier.
func Printcrf(format string, a ...any) {
	if _, err := fmt.Fprintf(os.Stdout, "\r%s\r%s",
		strings.Repeat(" ", int(terminal.Size())),
		fmt.Sprintf(format, a...)); err != nil {
		fatalLog(err)
	}
}

func fatal(err error) {
	log.Printf("%s %s", color.Danger.Sprint("!"), err)
	os.Exit(1)
}

func fatalLog(err error) {
	Save(Filename, err)
	fatal(err)
}

// Save an error to the logs.
// path is available for unit tests.
func Save(filename string, err error) (ok bool) {
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
