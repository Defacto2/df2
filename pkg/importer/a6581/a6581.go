package a6581

import (
	"regexp"
	"strconv"
	"time"
)

// Go doesn't permit numeric prefixes in package names.

const Name = "6581"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// Release.Date....: 03-2023
	rx := regexp.MustCompile(`Release.Date\.+: (\d\d)\-(\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)
	const expected = 3
	if len(f) == expected {
		y, _ := strconv.Atoi(f[2])
		m, _ := strconv.Atoi(f[1])
		d := 1 // we must include a valid day
		return y, time.Month(m), d
	}
	return 0, 0, 0
}
