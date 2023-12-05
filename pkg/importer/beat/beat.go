package beat

import (
	"regexp"
	"strconv"
	"time"
)

const Name = "Beat"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// ░█  DATE  ▓░ \   1013 2004
	//  Date.: 0319 2007
	// Date :     1211 2003
	rx := regexp.MustCompile(`(\d\d)(\d\d) (\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)
	const expected = 4
	if len(f) == expected {
		y, _ := strconv.Atoi(f[3])
		m, _ := strconv.Atoi(f[1])
		d, _ := strconv.Atoi(f[2])
		return y, time.Month(m), d
	}
	return 0, 0, 0
}
