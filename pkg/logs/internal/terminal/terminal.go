//go:build !windows

package terminal

import (
	"os"

	"golang.org/x/sys/unix"
)

// Size returns the character width of the terminal.
func Size() (columns uint16) {
	const falback = 80
	columns = uint16(falback)
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return columns
	}
	columns = ws.Col
	return columns
}
