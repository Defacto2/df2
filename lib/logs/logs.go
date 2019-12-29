package logs

import (
	"fmt"
	"log"
)

// Quiet silences most writing to the standard output.
var (
	Panic = false
	Quiet = false
)

// Check logs any errors and exits to the operating system with error code 1.
func Check(err error) {
	if err != nil {
		switch Panic {
		case true:
			log.Panic(err)
		default:
			log.Fatal("ERROR: ", err)
		}
	}
}

// Log any errors.
func Log(err error) {
	if err != nil {
		log.Printf("! %v", err)
	}
}

// Cli writes the string to the standard output.
func Cli(s string) {
	switch Quiet {
	case false:
		fmt.Print(s)
	}
}
