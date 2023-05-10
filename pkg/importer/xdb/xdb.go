package xdb

import (
	"regexp"
	"strconv"
	"time"
)

const Name = "Xdb"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}
	// Release Date...: 11.03.2013
	// Release Date......: 10.09.2012
	rx := regexp.MustCompile(`Release Date[.]{1,}: (\d\d)\.(\d\d)\.(\d\d\d\d)`)
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
