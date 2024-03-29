package assign

import (
	"regexp"
	"strconv"
	"time"

	"github.com/Defacto2/df2/pkg/importer/months"
)

const Name = "Assign"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// DATE : 01 MARCH 2015
	// DATE : 15 APRiL 2010
	rx := regexp.MustCompile(`DATE : (\d\d) ([a-zA-Z]{3,}) (\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)
	const expected = 4
	if len(f) == expected {
		y, _ := strconv.Atoi(f[3])
		s := f[2][:3]
		m := months.Months()[s]
		d, _ := strconv.Atoi(f[1])
		return y, time.Month(m), d
	}
	return 0, 0, 0
}
