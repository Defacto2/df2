// Package prompt are functions that parse stardard input loops.
package prompt

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Defacto2/df2/pkg/prompt/internal/input"
)

const (
	PortMin = input.PortMin // PortMin is the lowest permitted network port.
	PortMax = input.PortMax // PortMax is the largest permitted network port.
)

// Dir asks the user for a directory path and saves it.
func Dir() string {
	s, err := input.Dir(os.Stdin)
	if errors.Is(err, input.ErrEmpty) {
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return s
}

// Port asks the user for a port configuration value and returns the input.
func Port() int64 {
	i, err := input.Port(os.Stdin)
	if err != nil {
		os.Exit(1)
	}
	return i
}

// IsPort reports if the port is usable.
func IsPort(port int) bool {
	if port < PortMin || port > PortMax {
		return false
	}
	return true
}

// String asks the user for a string configuration value and saves it.
func String(keep string) string {
	fmt.Fprintln(os.Stdout, keep)
	s, err := input.String(os.Stdin)
	if errors.Is(err, input.ErrEmpty) {
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return s
}

// YN asks the user for a yes or no input.
func YN(query string, yes bool) bool {
	y, n := "Y", "n"
	if !yes {
		y, n = "y", "N"
	}
	fmt.Fprintf(os.Stdout, "%s? [%s/%s] ", query, y, n)
	in, err := input.Read(os.Stdin)
	if err != nil {
		log.Print(fmt.Errorf("prompt yn input: %w", err))
		return false
	}
	return input.YN(in, yes)
}
