package logs

import (
	"fmt"
	"strings"
)

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
