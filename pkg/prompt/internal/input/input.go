package input

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/str"
)

var (
	ErrDir      = errors.New("will not save the change as this directory is not found")
	ErrEmpty    = errors.New("empty directory input")
	ErrTooMany  = errors.New("too many stdin tries")
	ErrNoReader = errors.New("reader cannot be nil, it should be os.stdin")
)

const (
	PortMin = 0     // PortMin is the lowest permitted network port.
	PortMax = 65535 // PortMax is the largest permitted network port.
)

// Dir parses the reader for a valid directory input.
// Except for testing, r should always be os.Stdin.
func Dir(r io.Reader) (string, error) {
	if r == nil {
		return "", ErrNoReader
	}
	// allow multiple word user input
	scanner := bufio.NewScanner(r)
	var save string
	for scanner.Scan() {
		txt := scanner.Text()
		switch txt {
		case "":
			return "", ErrEmpty
		case "-":
			save = ""
		default:
			save = txt
		}
		if _, err := os.Stat(save); os.IsNotExist(err) {
			return "", fmt.Errorf("%w: %s", ErrDir, logs.Path(save))
		} else if err == nil {
			break // exit loop if the directory is found
		}
	}
	return save, nil
}

// Port parses the reader for a valid network port.
// Except for testing, r should always be os.Stdin.
func Port(r io.Reader) (int64, error) {
	if r == nil {
		return 0, ErrNoReader
	}
	const decimal = 10
	scanner := bufio.NewScanner(r)
	cnt := 0
	for scanner.Scan() {
		cnt++
		txt := scanner.Text()
		if txt == "" {
			if err := check(cnt); errors.Is(err, ErrTooMany) {
				return 0, err
			}
			continue
		}
		i, err := strconv.ParseInt(txt, decimal, 0)
		if err != nil && txt != "" {
			fmt.Printf("%s %v\n", str.X(), txt)
			if err := check(cnt); errors.Is(err, ErrTooMany) {
				return 0, err
			}
			continue
		}
		// check that the input a valid port
		if ok := IsPort(int(i)); !ok {
			fmt.Printf("%s %q is out of range\n", str.X(), txt)
			if err := check(cnt); errors.Is(err, ErrTooMany) {
				return 0, err
			}
			continue
		}
		return i, nil
	}
	return 0, nil
}

// check the number of attempts at asking for a valid port.
func check(cnt int) error {
	const help, max = 2, 4
	switch {
	case cnt == help:
		fmt.Println("Ctrl+C to keep the existing port")
	case cnt >= max:
		return ErrTooMany
	}
	return nil
}

// IsPort reports if the port is usable.
func IsPort(port int) bool {
	if port < PortMin || port > PortMax {
		return false
	}
	return true
}

// Read and trim the reader and return the results.
func Read(r io.Reader) (string, error) {
	reader := bufio.NewReader(r)
	s, err := reader.ReadString('\n')
	s = strings.TrimSpace(s)
	if err != nil && err != io.EOF {
		return s, err
	}
	return s, nil
}

// String parses the reader, looking for a string and newline.
// Except for testing, r should always be os.Stdin.
func String(r io.Reader) (string, error) {
	if r == nil {
		return "", ErrNoReader
	}
	// allow multiple word user input
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		txt := scanner.Text()
		switch txt {
		case "":
			return "", ErrEmpty
		case "-":
			return "", nil
		default:
			return txt, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", ErrEmpty
}

// YN parse s for a usable boolean value.
// An empty string will return true when emptyIsYes is true.
func YN(s string, emptyIsYes bool) bool {
	switch s {
	case "":
		if emptyIsYes {
			return true
		}
	case "yes", "y":
		return true
	}
	return false
}
