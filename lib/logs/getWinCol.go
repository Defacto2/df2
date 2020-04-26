// +build !windows

package logs

import "os"

func getWinCol() uint16 {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return uint16(80) // 80 column fallback
	}
	return ws.Col
}
