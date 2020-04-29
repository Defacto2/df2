// +build windows

package logs

func getWinCol() (columns uint16) {
	return uint16(80) // 80 column fallback
}
