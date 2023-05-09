package hexwars

import (
	"regexp"
	"strconv"
	"time"
)

const Name = "Hexwars"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// DATE : 25.05.2018
	// DATE : 27-01-2015
	rx := regexp.MustCompile(`DATE : (\d\d)[\.|\-](\d\d)[\.|\-](\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)
	const ddmmyy = 4
	if len(f) == ddmmyy {
		y, _ := strconv.Atoi(f[3])
		m, _ := strconv.Atoi(f[2])
		d, _ := strconv.Atoi(f[1])
		return y, time.Month(m), d
	}
	return 0, 0, 0
}
