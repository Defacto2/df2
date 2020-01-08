package logs

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/gookit/color.v1"
)

const (
	// PortMin is the lowest permitted network port
	PortMin int = 0
	// PortMax is the largest permitted network port
	PortMax int = 65535
)

var (
	// Panic uses the panic function to handle all error logs.
	Panic = false
	// Quiet stops most writing to the standard output.
	Quiet = false
)

// Arg returns instructions for invalid arguments.
func Arg(args []string) {
	Check(fmt.Errorf("invalid command %q please use one of the available config commands", args[0]))
}

// Check logs any errors and exits to the operating system with error code 1.
func Check(err error) {
	if err != nil {
		switch Panic {
		case true:
			fmt.Printf("error type: %T\tmsg: %v\n", err, err)
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

// Out writes the string to the standard output.
func Out(s string) {
	switch Quiet {
	case false:
		fmt.Print(s)
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

// X returns a red cross mark.
func X() string {
	return color.Danger.Sprint("✗")
}

// Y returns a green tick mark.
func Y() string {
	return color.Success.Sprint("✓")
}

// File is a logger for common os package functions.
// config is an optional configuration path used by cmd.config.
func File(config string, err error) {
	var pathError *os.PathError
	if errors.As(err, &pathError) {
		fmt.Println(X(), "failed to create or open file:", Path(pathError.Path))
		if config != "" {
			fmt.Println("  to fix run:", color.Info.Sprintf("config set --name %v", config))
		}
		if Panic {
			log.Panic(err)
		}
		os.Exit(1)
	} else {
		Log(err)
	}
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
		fmt.Println("ctrl+C to keep the existing port")
	case cnt >= 4:
		os.Exit(1)
	}
}

func scannerCheck(s *bufio.Scanner) {
	if err := s.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
		os.Exit(1)
	}
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
			os.Exit(1)
		}
		return save
	}
	scannerCheck(scanner)
	return ""
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
	y := "Y"
	n := "n"
	if !yesDefault {
		y = "y"
		n = "N"
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
