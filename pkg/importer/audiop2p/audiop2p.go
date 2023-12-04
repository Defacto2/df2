package audiop2p

import (
	"regexp"
	"strconv"
	"time"
)

const Name = "AudioP2P"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// Prepared...: 01-27-2015
	rx := regexp.MustCompile(`Prepared\.+: (\d\d)\-(\d\d)\-(\d\d\d\d)`)
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
