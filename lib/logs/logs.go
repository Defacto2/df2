package logs

import (
	"fmt"
	"log"

	"gopkg.in/gookit/color.v1"
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

// ProgressPct returns the count of total remaining as a percentage.
func ProgressPct(name string, count int, total int) float64 {
	r := float64(count) / float64(total) * 100
	switch r {
	case 100:
		fmt.Printf("\rQuerying %s %.0f %%  ", name, r)
	default:
		fmt.Printf("\rQuerying %s %.2f %%", name, r)
	}
	return r
}

// ProgressSum returns the count of total remaining.
// func ProgressSum(count int, total int) {
// 	fmt.Printf("\rBuilding %d/%d", count, total)
// }

// Sec prints a secondary notice.
func Sec(s string) {
	color.Secondary.Println(s)
}
// Warn prints a warning notice.
func Warn(s string) {
	color.Warn.Println(s)
}
