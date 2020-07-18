// +build windows

package logs

func termSize() (columns uint16) {
	const falback = 80
	return uint16(falback)
}
