package logs

// os.Exit() = 1x

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	gap "github.com/muesli/go-app-paths"
	"gopkg.in/gookit/color.v1"
)

const (
	// PortMin is the lowest permitted network port
	PortMin int = 0
	// PortMax is the largest permitted network port
	PortMax int = 65535
	// AED is the ANSI Erase in Display
	AED string = "\r\003[2J"
	// AEL is the ANSI Erase in Line sequence
	AEL string = "\r\033[0K"
	// Filename is the default error log filename
	Filename string = "errors.log"
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
	os.Exit(10)
}

// Check logs any errors and exits to the operating system with error code 1.
func Check(err error) {
	if err != nil {
		save(err)
		switch Panic {
		case true:
			println(fmt.Sprintf("error type: %T\tmsg: %v", err, err))
			log.Panic(err)
		default:
			log.Fatal(color.Danger.Sprint("ERROR: "), err)
		}
	}
}

// EL prints the ANSI command Erase in Line.
func EL() {
	Print(AEL)
}

// Log an error but do not exit to the operating system.
func Log(err error) {
	if err != nil {
		save(err)
		switch Panic {
		case true:
			println(fmt.Sprintf("error type: %T\tmsg: %v", err, err))
			log.Panic(err)
		default:
			log.Printf("%s %s", color.Danger.Sprint("!"), err)
		}
	}
}

func save(err error) {
	if err == nil {
		return
	}
	path := Filepath()
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
}

func check(e error) {
	if e != nil {
		log.Printf("%s %s", color.Danger.Sprint("!"), e)
		os.Exit(19)
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

// Print obeys the --quiet flag or formats using the default formats for its operands and writes to standard output.
func Print(a ...interface{}) {
	switch Quiet {
	case false:
		_, err := fmt.Print(a...)
		Log(err)
	}
}

// Printf obeys the --quiet flag or formats according to a format specifier and writes to standard output.
func Printf(format string, a ...interface{}) {
	switch Quiet {
	case false:
		_, err := fmt.Printf(format, a...)
		Log(err)
	}
}

// Println obeys the --quiet flag or formats using the default formats for its operands and writes to standard output.
func Println(a ...interface{}) {
	switch Quiet {
	case false:
		_, err := fmt.Println(a...)
		Log(err)
	}
}

// ProgressPct returns the count of total remaining as a percentage.
func ProgressPct(name string, count, total int) float64 {
	r := float64(count) / float64(total) * 100
	switch r {
	case 100:
		fmt.Printf("\rquerying %s %.0f %%  ", name, r)
	default:
		fmt.Printf("\rquerying %s %.2f %%", name, r)
	}
	return r
}

// ProgressSum returns the count of total remaining. TODO toggle with a configuration setting
func ProgressSum(count, total int) {
	fmt.Printf("\rBuilding %d/%d", count, total)
}

// Sec prints a secondary notice.
func Sec(s string) string {
	return color.Secondary.Sprint(s)
}

// Warn prints a warning notice.
func Warn(s string) string {
	return color.Warn.Sprint(s)
}

// X returns a red ✗ cross mark.
func X() string {
	return color.Danger.Sprint("✗")
}

// Y returns a green ✓ tick mark.
func Y() string {
	return color.Success.Sprint("✓")
}

// File is a logger for common os package functions.
// config is an optional configuration path used by cmd.config.
func File(config string, err error) {
	var pathError *os.PathError
	if errors.As(err, &pathError) {
		log.Println(X(), "failed to create or open file:", Path(pathError.Path))
		if config != "" {
			Println("  to fix run:", color.Info.Sprintf("config set --name %v", config))
		}
		if Panic {
			log.Panic(err)
		}
		os.Exit(11)
	} else {
		Log(err)
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

// Port reports if the value is valid.
func Port(port int) bool {
	if port < PortMin || port > PortMax {
		return false
	}
	return true
}

// promptCheck asks the user for a string configuration value and saves it.
func promptCheck(cnt int) {
	switch {
	case cnt == 2:
		fmt.Println("Ctrl+C to keep the existing port")
	case cnt >= 4:
		os.Exit(12)
	}
}

func scannerCheck(s *bufio.Scanner) {
	if err := s.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
		os.Exit(13)
	}
}

// Simulate prints the --simulate=false flag info.
func Simulate() {
	Println(color.Notice.Sprint("use the --simulate=false flag to apply these fixes"))
}

// PromptDir asks the user for a directory path and saves it.
func PromptDir() string {
	// allow multiple word user input
	scanner := bufio.NewScanner(os.Stdin)
	var save string
	for scanner.Scan() {
		txt := scanner.Text()
		switch txt {
		case "":
			os.Exit(0)
		case "-":
			save = ""
		default:
			save = txt
		}
		if _, err := os.Stat(save); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "will not save the change as this directory is not found:", Path(save))
			os.Exit(14)
		}
		break // exit loop if the directory is found
	}
	scannerCheck(scanner)
	return save
}

// PromptPort asks the user for a port configuration value and returns the input.
func PromptPort() int64 {
	var input string
	cnt := 0
	for {
		input = ""
		cnt++
		fmt.Scanln(&input)
		if input == "" {
			promptCheck(cnt)
			continue
		}
		i, err := strconv.ParseInt(input, 10, 0)
		if err != nil && input != "" {
			fmt.Printf("%s %v\n", X(), input)
			promptCheck(cnt)
			continue
		}
		// check that the input a valid port
		if v := Port(int(i)); !v {
			fmt.Printf("%s %q is out of range\n", X(), input)
			promptCheck(cnt)
			continue
		}
		return i
	}
}

// PromptString asks the user for a string configuration value and saves it.
func PromptString(keep string) string {
	// allow multiple word user input
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		txt := scanner.Text()
		switch txt {
		case "":
			os.Exit(0)
		case "-":
			return ""
		default:
			return txt
		}
	}
	scannerCheck(scanner)
	os.Exit(0)
	return ""
}

// PromptYN asks the user for a yes or no input.
func PromptYN(query string, yesDefault bool) bool {
	var input string
	var y, n string = "Y", "n"
	if !yesDefault {
		y, n = "y", "N"
	}
	fmt.Printf("%s? [%s/%s] ", query, y, n)
	fmt.Scanln(&input)
	switch input {
	case "":
		if yesDefault {
			return true
		}
	case "yes", "y":
		return true
	}
	return false
}

// Truncate shortens a string to len characters.
func Truncate(text string, len int) string {
	var new string = "…"
	if utf8.RuneCountInString(text) <= len {
		return text
	}
	return text[0:len-utf8.RuneCountInString(new)] + new
}
