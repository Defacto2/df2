package arctic

import (
	"regexp"
	"strconv"
	"time"

	"github.com/Defacto2/df2/pkg/str"
)

const Name = "Arctic"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// Date        : 03-16-03
	rx := regexp.MustCompile(`Date        : (\d\d)\-(\d\d)\-(\d\d)`)
	f := rx.FindStringSubmatch(body)
	const ddmmyy = 4
	if len(f) == ddmmyy {
		y := str.YearAbbr(f[3])
		m, _ := strconv.Atoi(f[1])
		d, _ := strconv.Atoi(f[2])
		return y, time.Month(m), d
	}
	return 0, 0, 0
}
