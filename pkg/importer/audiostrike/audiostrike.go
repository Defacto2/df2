package audiostrike

import (
	"regexp"
	"strconv"
	"time"
)

const Name = "AudioStrike"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// DATE....: 29/05/2015
	rx := regexp.MustCompile(`DATE\.+: (\d\d)\/(\d\d)\/(\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)
	const expected = 4
	if len(f) == expected {
		y, _ := strconv.Atoi(f[3])
		m, _ := strconv.Atoi(f[2])
		d, _ := strconv.Atoi(f[1])
		return y, time.Month(m), d
	}
	return 0, 0, 0
}
