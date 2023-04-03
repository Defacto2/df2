// Package str has custom print to terminal, display and colour functions.
package str

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/gookit/color"
)

// Piped detects whether the program text is being piped to another operating
// system command or sent to stdout.
func Piped() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		log.Fatal(err)
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// Progress returns the count of total remaining as a percentage.
func Progress(w io.Writer, name string, count, total int) float64 {
	if w == nil {
		w = io.Discard
	}
	const fin = 100
	r := float64(count) / float64(total) * fin
	switch r {
	case fin:
		fmt.Fprintf(w, "\rquerying %s %s %.0f %%  \n", name, bar(r), r)
	default:
		fmt.Fprintf(w, "\rquerying %s %s %.2f %% ", name, bar(r), r)
	}
	return r
}

func bar(r float64) string {
	const (
		c     = "\u15e7" // ᗧ
		pad   = "\u2022" // •
		done  = "\u00b7" // ·
		width = 5
		start = 0
		end   = 100
	)
	pos, max := math.Max(0, r/width), end/width
	switch {
	case pos == start:
		return fmt.Sprintf("(%s%s)", c, strings.Repeat(pad, max))
	case r == end, pos > float64(max):
		return fmt.Sprintf("(%s☺)", strings.Repeat(pad, max))
	default:
		return fmt.Sprintf("(%s%s%s)",
			strings.Repeat(done, int(pos)), c,
			strings.Repeat(pad, max-int(pos)))
	}
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
