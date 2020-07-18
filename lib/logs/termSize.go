// +build !windows

package logs

import (
	"os"

	"golang.org/x/sys/unix"
)

func termSize() (columns uint16) {
	const falback = 80
	columns = uint16(falback)
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return columns
	}
	columns = ws.Col
	return columns
}
