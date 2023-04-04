//go:build !windows

// Package terminal returns information on the host Linux terminal.
package terminal

import (
	"os"

	"golang.org/x/sys/unix"
)

// Columns returns the character width of the terminal.
func Columns() uint16 {
	const fallback = uint16(80)
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return fallback
	}
	return ws.Col
}
