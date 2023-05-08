package arcade

import (
	"regexp"
	"strconv"
	"time"

	"github.com/Defacto2/df2/pkg/str"
)

const Name = "Arcade"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// DATE ......: 02/2011
	// Date..: 12/2009
	// Date.: 12/2008
	rx := regexp.MustCompile(`(?i)DATE[ ]{0,}[.]{1,}: (\d?\d)\/(\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)
	const mmyyyy = 3
	if len(f) == mmyyyy {
		y, _ := strconv.Atoi(f[2])
		m, _ := strconv.Atoi(f[1])
		d := 0
		return y, time.Month(m), d
	}

	// Date.: 05.01.06
	// DATE:          08.31.04
	rx = regexp.MustCompile(`(?i)Date[.]{0,}:[ ]{1,}(\d\d)\.(\d\d)\.(\d\d)`)
	f = rx.FindStringSubmatch(body)
	const ddmmyy = 4
	if len(f) == ddmmyy {
		y := str.YearAbbr(f[3])
		m, _ := strconv.Atoi(f[1])
		d, _ := strconv.Atoi(f[2])
		return y, time.Month(m), d
	}
	return 0, 0, 0
}
