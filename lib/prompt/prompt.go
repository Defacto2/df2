// Package prompt are functions that parse stardard input loops.
package prompt

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
)

const (
	// PortMin is the lowest permitted network port
	PortMin = 0
	// PortMax is the largest permitted network port
	PortMax = 65535
)

// Dir asks the user for a directory path and saves it.
func Dir() string {
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
			fmt.Fprintln(os.Stderr, "will not save the change as this directory is not found:", logs.Path(save))
			os.Exit(1)
		} else {
			break // exit loop if the directory is found
		}
	}
	return save
}

// Port asks the user for a port configuration value and returns the input.
func Port() int64 {
	var input string
	cnt := 0
	for {
		input = ""
		cnt++
		fmt.Scanln(&input)
		if input == "" {
			check(cnt)
			continue
		}
		i, err := strconv.ParseInt(input, 10, 0)
		if err != nil && input != "" {
			fmt.Printf("%s %v\n", str.X(), input)
			check(cnt)
			continue
		}
		// check that the input a valid port
		if v := port(int(i)); !v {
			fmt.Printf("%s %q is out of range\n", str.X(), input)
			check(cnt)
			continue
		}
		return i
	}
}

// port reports if the value is valid.
func port(port int) bool {
	if port < PortMin || port > PortMax {
		return false
	}
	return true
}

// String asks the user for a string configuration value and saves it.
func String(keep string) string {
	fmt.Println(keep)
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
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
		os.Exit(1)
	}
	os.Exit(0)
	return ""
}

// YN asks the user for a yes or no input.
func YN(query string, yes bool) bool {
	var y, n string = "Y", "n"
	if !yes {
		y, n = "y", "N"
	}
	fmt.Printf("%s? [%s/%s] ", query, y, n)
	input, err := read(os.Stdin)
	if err != nil {
		log.Fatal(fmt.Errorf("prompt yn input: %w", err))
	}
	return parseyn(input, yes)
}

// check asks the user for a string configuration value and saves it.
func check(cnt int) {
	const help, max = 2, 4
	switch {
	case cnt == help:
		fmt.Println("Ctrl+C to keep the existing port")
	case cnt >= max:
		os.Exit(1)
	}
}

func parseyn(input string, yes bool) bool {
	switch input {
	case "":
		if yes {
			return true
		}
	case "yes", "y":
		return true
	}
	return false
}

func read(stdin io.Reader) (input string, err error) {
	reader := bufio.NewReader(stdin)
	input, err = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if err != nil && err != io.EOF {
		return input, err
	}
	return input, nil
}
