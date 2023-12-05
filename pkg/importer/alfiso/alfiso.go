package alfiso

import (
	"regexp"
	"strconv"
	"time"
)

const Name = "AlfISO"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// Release date : 25.05.18
	rx := regexp.MustCompile(`Release date : (\d\d)\.(\d\d)\.(\d\d)`)
	f := rx.FindStringSubmatch(body)
	const ddmmyy = 4
	if len(f) == ddmmyy {
		y, _ := strconv.Atoi(f[3])
		y = 2000 + y // append century
		m, _ := strconv.Atoi(f[2])
		d, _ := strconv.Atoi(f[1])
		return y, time.Month(m), d
	}
	// ReL. DaTe : 2004-04-11
	rx = regexp.MustCompile(`ReL. DaTe : (\d\d\d\d)\-(\d\d)\-(\d\d)`)
	f = rx.FindStringSubmatch(body)
	if len(f) == ddmmyy {
		y, _ := strconv.Atoi(f[1])
		m, _ := strconv.Atoi(f[2])
		d, _ := strconv.Atoi(f[3])
		return y, time.Month(m), d
	}
	// REL. DATE : 28-03-2003
	rx = regexp.MustCompile(`REL. DATE : (\d\d)\-(\d\d)\-(\d\d\d\d)`)
	f = rx.FindStringSubmatch(body)
	if len(f) == ddmmyy {
		y, _ := strconv.Atoi(f[3])
		m, _ := strconv.Atoi(f[2])
		d, _ := strconv.Atoi(f[1])
		return y, time.Month(m), d
	}
	return 0, 0, 0
}
