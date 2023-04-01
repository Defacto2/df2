// Package prompt are functions that parse stardard input loops.
package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var ErrReader = errors.New("reader io cannot be nil")

// Read and trim the reader and return the results.
func Read(r io.Reader) (string, error) {
	if r == nil {
		return "", ErrReader
	}
	reader := bufio.NewReader(r)
	s, err := reader.ReadString('\n')
	s = strings.TrimSpace(s) // remove the newline caused from the Enter keypress
	if err != nil && err != io.EOF {
		return s, err
	}
	return s, nil
}

// YN asks the user for a yes or no input.
func YN(w io.Writer, s string, defaultY bool) (bool, error) {
	if w == nil {
		w = io.Discard
	}
	y, n := "Y", "n"
	if !defaultY {
		y, n = "y", "N"
	}
	fmt.Fprintf(w, "%s? [%s/%s] ", s, y, n)
	input, err := Read(os.Stdin)
	if err != nil {
		return false, fmt.Errorf("prompt yn input: %w", err)
	}
	switch input {
	case "":
		if defaultY {
			return true, nil
		}
	case "yes", "y":
		return true, nil
	}
	return false, nil
}
