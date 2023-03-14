// Package logs handles errors and user feedback.
package logs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/pkg/logs/internal/terminal"
	"github.com/gookit/color"
	gap "github.com/muesli/go-app-paths"
	"go.uber.org/zap"
)

var (
	ErrNoCmd = errors.New("no command argument was provided")
	ErrNoErr = errors.New("cannot save a nil error")
	ErrNoZap = errors.New("no zap logger was provided")
)

const (
	GapUser  string = "df2"        // GapUser is the configuration and logs subdirectory name.
	Filename string = "errors.log" // Filename is the default error log filename.

	dmode   os.FileMode = 0o700
	fmode   os.FileMode = 0o600
	flags   int         = log.Ldate | log.Ltime | log.LUTC
	newmode int         = os.O_APPEND | os.O_CREATE | os.O_WRONLY
)

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

// Filepath is the absolute path and filename of the error log file.
func Filepath(log *zap.SugaredLogger, filename string) string {
	fp, err := gap.NewScope(gap.User, GapUser).LogPath(filename)
	if err != nil {
		h, err := os.UserHomeDir()
		if err != nil {
			return filename
		}
		return path.Join(h, filename)
	}
	return fp
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

// Printcr otherwise erases the current line and writes to standard output.
func Printcr(w io.Writer, a ...any) {
	fmt.Fprintf(w, "\r%s\r", strings.Repeat(" ", int(terminal.Size())))
	fmt.Fprint(w, a...)
}

// Printcrf erases the current line and formats according to a format specifier.
func Printcrf(w io.Writer, format string, a ...any) {
	fmt.Fprintf(w, "\r%s\r%s",
		strings.Repeat(" ", int(terminal.Size())),
		fmt.Sprintf(format, a...))
}

// Save logs and stores the error to a log file.
func Save(l *zap.SugaredLogger, err error) error {
	return Saver(l, Filename, err)
}

// Saver stores the error to the named log file.
// path is available for unit tests.
func Saver(l *zap.SugaredLogger, name string, err error) error {
	if l == nil {
		return ErrNoZap
	}
	if err == nil {
		return ErrNoErr
	}
	// use UTC date and times in the log file
	log.SetFlags(flags)
	f := Filepath(l, name)
	p := filepath.Dir(f)
	if _, err1 := os.Stat(p); os.IsNotExist(err1) {
		if err2 := os.MkdirAll(p, dmode); err != nil {
			fatal(l, err2)
		}
	} else if err1 != nil {
		fatal(l, err1)
	}
	file, err1 := os.OpenFile(f, newmode, fmode)
	if err1 != nil {
		fatal(l, err1)
	}
	defer file.Close()
	log.SetOutput(file)
	log.Print(err)
	log.SetOutput(os.Stderr)
	return nil
}

func fatal(l *zap.SugaredLogger, err error) {
	if l == nil {
		log.Fatal(err)
	}
	l.Errorln(err)
	//log.Printf("%s %s", color.Danger.Sprint("!"), err)
	os.Exit(1)
}
