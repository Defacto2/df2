//go:build windows

package terminal

func Size() (columns uint16) {
	const falback = 80
	return uint16(falback)
}
