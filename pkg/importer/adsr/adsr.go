package adsr

import (
	"regexp"
	"strconv"
	"time"
)

const Name = "Attack Decay Sustain Release"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// RELEASE DATE................................................: 06-2018
	rx := regexp.MustCompile(`RELEASE DATE\.+: (\d\d)\-(\d\d\d\d)`)
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
