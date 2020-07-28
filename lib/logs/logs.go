package logs

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	gap "github.com/muesli/go-app-paths"
	"gopkg.in/gookit/color.v1"
)

const (
	// AED is the ANSI Erase in Display
	AED = "\r\003[2J"
	// AEL is the ANSI Erase in Line sequence
	AEL = "\r\033[0K"
	// Filename is the default error log filename
	Filename = "errors.log"
)

var (
	scope = gap.NewScope(gap.User, "df2")
	// Panic uses the panic function to handle all error logs.
	Panic = false
	// Quiet stops most writing to the standard output.
	Quiet = false
)

// Arg returns instructions for invalid command arguments.
func Arg(arg string, args []string) {
	fmt.Printf("%s %s %s\n",
		color.Warn.Sprint("invalid command"),
		color.Bold.Sprintf("\"%s %s\"", arg, args[0]),
		color.Warn.Sprint("\nplease use one of the Available Commands shown above"))
	os.Exit(1)
}

// Check logs any errors and exits to the operating system with error code 1.
func Check(err error) {
	if err != nil {
		save(err, "")
		switch Panic {
		case true:
			log.Println(fmt.Sprintf("error type: %T\tmsg: %v", err, err))
			log.Panic(err)
		default:
			log.Fatal(color.Danger.Sprint("ERROR: "), err)
		}
	}
}

// Filepath is the absolute path and filename of the error log file.
func Filepath() string {
	fp, err := scope.LogPath(Filename)
	if err != nil {
		h, _ := os.UserHomeDir()
		return path.Join(h, Filename)
	}
	return fp
}

// Log an error but do not exit to the operating system.
func Log(err error) {
	if err != nil {
		save(err, "")
		switch Panic {
		case true:
			log.Println(fmt.Sprintf("error type %T\t: %v", err, err))
			log.Panic(err)
		default:
			log.Printf("%s %s", color.Danger.Sprint("!"), err)
		}
	}
}

// Path returns a file or directory path with all missing elements marked in red.
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

// Print obeys the --quiet flag or formats using the default formats for its operands and writes to standard output.
func Print(a ...interface{}) {
	if !Quiet {
		_, err := fmt.Print(a...)
		Log(err)
	}
}

// Printcr obeys the --quiet flag or otherwise erases the current line and writes to standard output.
func Printcr(a ...interface{}) {
	if !Quiet {
		cols := int(termSize())
		fmt.Printf("\r%s\r", strings.Repeat(" ", cols))
		_, err := fmt.Print(a...)
		Log(err)
	}
}

// Printf obeys the --quiet flag or formats according to a format specifier and writes to standard output.
func Printf(format string, a ...interface{}) {
	if !Quiet {
		_, err := fmt.Printf(format, a...)
		Log(err)
	}
}

// Println obeys the --quiet flag or formats using the default formats for its operands and writes to standard output.
func Println(a ...interface{}) {
	if !Quiet {
		_, err := fmt.Println(a...)
		Log(err)
	}
}

// Printcrf obeys the --quiet flag or otherwise erases the current line and formats according to a format specifier.
func Printcrf(format string, a ...interface{}) {
	if !Quiet {
		cols := int(termSize())
		fmt.Printf("\r%s\r", strings.Repeat(" ", cols))
		_, err := fmt.Printf(format, a...)
		Log(err)
	}
}

// Simulate prints the --simulate=false flag info.
func Simulate() {
	Println(color.Notice.Sprint("use the --simulate=false flag to apply these fixes"))
}

func check(e error) {
	if e != nil {
		log.Printf("%s %s", color.Danger.Sprint("!"), e)
		os.Exit(1)
	}
}

// save an error to the logs.
// path is available for unit tests.
func save(err error, path string) (ok bool) {
	if err == nil || fmt.Sprintf("%v", err) == "" {
		return false
	}
	// use UTC date and times in the log file
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	if path == "" {
		path = Filepath()
	}
	p := filepath.Dir(path)
	if _, e := os.Stat(p); os.IsNotExist(e) {
		e2 := os.MkdirAll(p, 0700)
		check(e2)
	}
	file, e := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	check(e)
	defer file.Close()
	log.SetOutput(file)
	log.Print(err)
	log.SetOutput(os.Stderr)
	return true
}
