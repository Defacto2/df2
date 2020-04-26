// +build windows

package logs

func getWinCol() uint16 {
	return uint16(80) // 80 column fallback
}
