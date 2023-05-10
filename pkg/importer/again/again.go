package again

import (
	"regexp"
	"strconv"
	"time"
)

const Name = "Again"

func DizDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// released on 12/10/2004
	// Released on 12/10/2004
	rx := regexp.MustCompile(`(?i)released on (\d\d)\/(\d\d)\/(\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)

	const expected = 4
	if len(f) != expected {
		return 0, 0, 0
	}
	y, _ := strconv.Atoi(f[3])
	m, _ := strconv.Atoi(f[2])
	d, _ := strconv.Atoi(f[1])
	return y, time.Month(m), d
}

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// Date........: 14/09/2008
	rx := regexp.MustCompile(`Date........: (\d\d)\/(\d\d)\/(\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)

	const expected = 4
	if len(f) != expected {
		return 0, 0, 0
	}
	y, _ := strconv.Atoi(f[3])
	m, _ := strconv.Atoi(f[2])
	d, _ := strconv.Atoi(f[1])
	return y, time.Month(m), d
}
