package audioutopia

import (
	"regexp"
	"strconv"
	"time"

	"github.com/Defacto2/df2/pkg/importer/months"
)

const Name = "AudioUtopia"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// Date.................: 19.DEC.2015
	// Date.................: 27.SEPT.2015
	rx := regexp.MustCompile(`Date\.+: (\d\d)\.([a-zA-Z]{3,})\.(\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)
	const expected = 4
	if len(f) == expected {
		y, _ := strconv.Atoi(f[3])
		m := months.Months()[f[2][:3]]
		d, _ := strconv.Atoi(f[1])
		return y, time.Month(m), d
	}
	return 0, 0, 0
}
