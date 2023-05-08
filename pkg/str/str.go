// Package str has string and stdout functions.
package str

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gookit/color"
)

const (
	NothingToDo = "all good, there is nothing to do"
)

func TimeTaken(w io.Writer, elapsed float64) {
	if w == nil {
		w = io.Discard
	}
	fmt.Fprintf(w, "\ttime taken %.1f seconds\n", elapsed)
}

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

// X returns a red ✗ cross mark.
func X() string {
	return color.Danger.Sprint("✗")
}

// Y returns a green ✓ tick mark.
func Y() string {
	return color.Success.Sprint("✓")
}

// PathTitle returns the title of the release extracted from the
// named directory path of the release. This is intended as a fallback
// when a file_id.diz cannot be parsed.
func PathTitle(name string) string {
	n := strings.LastIndex(name, "-")
	t := name
	if n > -1 {
		t = name[0:n]
	}
	// match v1.0.0
	r := regexp.MustCompile(`v(\d+)\.(\d+)\.(\d+)`)
	t = r.ReplaceAllString(t, "v$1-$2-$3")
	// match v1.0
	r = regexp.MustCompile(`v(\d+)\.(\d+)`)
	t = r.ReplaceAllString(t, "v$1-$2")
	// word fixes
	words := strings.Split(t, ".")
	for i, word := range words {
		switch strings.ToLower(word) {
		case "incl":
			words[i] = "including"
		case "keymaker":
			words[i] = "keymaker"
		}
	}
	t = strings.Join(words, " ")
	// restore v1.0.0
	r = regexp.MustCompile(`v(\d+)-(\d+)-(\d+)`)
	t = r.ReplaceAllString(t, "v$1.$2.$3")
	// restore v1.0
	r = regexp.MustCompile(`v(\d+)-(\d+)`)
	t = r.ReplaceAllString(t, "v$1.$2")
	return strings.TrimSpace(t)
}

func YearAbbr(s string) int {
	y, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	const twentieth, twentyfirst = 1900, 2000
	if y <= time.Now().Year() {
		return twentyfirst + y
	}
	return twentieth + y
}

func GetTerminalWidth() (int, error) {
	cmd := exec.Command("tput", "cols")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Convert the output to an integer
	width, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, err
	}

	return width, nil
}

func RemoveLine() string {
	const def, space = 80, ` `
	w, err := GetTerminalWidth()
	if err != nil {
		return strings.Repeat(space, def)
	}
	return strings.Repeat(space, w)
}
