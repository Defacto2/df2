package bean

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/importer/months"
)

const Name = "Bean"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// Date : May 01, 2008
	// Date : May 1, 2008
	// Date : May 1,2008
	rx := regexp.MustCompile(`Date : ([a-zA-Z]{3,}) (\d{1,2})\, ?(\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)
	const expected = 4
	if len(f) == expected {
		y, _ := strconv.Atoi(f[3])
		s := strings.ToUpper(f[1][:3])
		m := months.Months()[s]
		d, _ := strconv.Atoi(f[2])
		return y, time.Month(m), d
	}
	return 0, 0, 0
}
