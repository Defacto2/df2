package str

import (
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/gookit/color"

	"github.com/Defacto2/df2/lib/logs"
)

// Piped detects whether the program text is being piped to another operating
// system command or sent to stdout.
func Piped() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		logs.Fatal(err)
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// Progress returns the count of total remaining as a percentage.
func Progress(name string, count, total int) float64 {
	const fin = 100
	r := float64(count) / float64(total) * fin
	switch r {
	case fin:
		fmt.Printf("\rquerying %s %.0f %%  ", name, r)
	default:
		fmt.Printf("\rquerying %s %.2f %%", name, r)
	}
	return r
}

// Sec prints a secondary notice.
func Sec(s string) string {
	return color.Secondary.Sprint(s)
}

// Truncate shortens a string to length characters.
func Truncate(text string, length int) string {
	if length < 1 {
		return text
	}
	const s = "…"
	if utf8.RuneCountInString(text) <= length {
		return text
	}
	return text[0:length-utf8.RuneCountInString(s)] + s
}

// Warn prints a warning notice.
func Warn(s string) string {
	return color.Warn.Sprint(s)
}

// X returns a red ✗ cross mark.
func X() string {
	return color.Danger.Sprint("✗")
}

// Y returns a green ✓ tick mark.
func Y() string {
	return color.Success.Sprint("✓")
}
