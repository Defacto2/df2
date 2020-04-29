// +build !windows

package logs

import (
	"os"

	"golang.org/x/sys/unix"
)

func getWinCol() (columns uint16) {
	columns = uint16(80) // 80 column fallback
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return columns
	}
	columns = ws.Col
	return columns
}
