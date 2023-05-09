package spirit

import (
	"regexp"
	"strconv"
	"time"
)

const Name = "Spirit"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}
	const (
		ddmmyy = 4
		yyyymm = 3
	)
	// DATE: 02/27/2007
	// DATE: 11-21-2005
	rx := regexp.MustCompile(`DATE: (\d\d)[\/|\-](\d\d)[\/|\-](\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)
	if len(f) == ddmmyy {
		y, _ := strconv.Atoi(f[3])
		m, _ := strconv.Atoi(f[1])
		d, _ := strconv.Atoi(f[2])
		return y, time.Month(m), d
	}
	// DATE: 13.12.2006 (note the different DD.MM position)
	rx = regexp.MustCompile(`DATE: (\d\d)\.(\d\d)\.(\d\d\d\d)`)
	f = rx.FindStringSubmatch(body)
	if len(f) == ddmmyy {
		y, _ := strconv.Atoi(f[3])
		m, _ := strconv.Atoi(f[2])
		d, _ := strconv.Atoi(f[1])
		return y, time.Month(m), d
	}
	// DATE.....: 19.04.2010
	rx = regexp.MustCompile(`DATE.....: (\d\d)\.(\d\d)\.(\d\d\d\d)`)
	f = rx.FindStringSubmatch(body)
	if len(f) == ddmmyy {
		y, _ := strconv.Atoi(f[3])
		m, _ := strconv.Atoi(f[2])
		d, _ := strconv.Atoi(f[1])
		return y, time.Month(m), d
	}
	// DATE: 2006-10-xx
	rx = regexp.MustCompile(`DATE: (\d\d\d\d)\-(\d\d)\-xx`)
	f = rx.FindStringSubmatch(body)
	if len(f) == yyyymm {
		y, _ := strconv.Atoi(f[1])
		m, _ := strconv.Atoi(f[2])
		d := 0
		return y, time.Month(m), d
	}
	// DATE: 2006-08-08
	// DATE....: 2007-07-13
	rx = regexp.MustCompile(`DATE[.]{0,}: (\d\d\d\d)\-(\d\d)\-(\d\d)`)
	f = rx.FindStringSubmatch(body)
	if len(f) == ddmmyy {
		y, _ := strconv.Atoi(f[1])
		m, _ := strconv.Atoi(f[2])
		d, _ := strconv.Atoi(f[3])
		return y, time.Month(m), d
	}
	return 0, 0, 0
}
