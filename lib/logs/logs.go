package logs

import (
	"fmt"
	"log"
)

var (
	// Panic uses the panic function to handle all error logs.
	Panic = false
	// Quiet stops most writing to the standard output.
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

// Cli writes the string to the standard output.
func Cli(s string) {
	switch Quiet {
	case false:
		fmt.Print(s)
	}
}

// Log any errors.
func Log(err error) {
	if err != nil {
		log.Printf("! %v", err)
	}
}
