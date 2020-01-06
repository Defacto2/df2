package logs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

// X returns a red cross mark.
func X() string {
	return color.Danger.Sprint("✗")
}

// Y returns a green tick mark.
func Y() string {
	return color.Success.Sprint("✓")
}

// Sec prints a secondary notice.
func Sec(s string) {
	color.Secondary.Println(s)
}

// Warn prints a warning notice.
func Warn(s string) {
	color.Warn.Println(s)
}

// Path returns a file or directory path with all missing elements marked in red.
func Path(name string) string {
	a := strings.Split(name, "/")
	var p string
	var s string
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
