//go:build windows

// Package terminal returns information on the host Windows terminal.
package terminal

// Columns on Windows will always returns 80 characters.
func Columns() uint16 {
	const fallback = 80
	return uint16(fallback)
}
