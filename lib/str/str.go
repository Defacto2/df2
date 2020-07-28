package str

import (
	"unicode/utf8"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
)

// Sec prints a secondary notice.
func Sec(s string) string {
	return color.Secondary.Sprint(s)
}

// Simulate prints the --simulate=false flag info.
func Simulate() {
	logs.Println(color.Notice.Sprint("use the --simulate=false flag to apply these fixes"))
}

// Truncate shortens a string to len characters.
func Truncate(text string, len int) string {
	if len < 1 {
		return text
	}
	const new string = "…"
	if utf8.RuneCountInString(text) <= len {
		return text
	}
	return text[0:len-utf8.RuneCountInString(new)] + new
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
