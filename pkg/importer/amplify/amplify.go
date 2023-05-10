package amplify

import (
	"regexp"
	"strconv"
	"time"

	"github.com/Defacto2/df2/pkg/str"
)

const Name = "Amplify"

func DizDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}
	// -[08.15.2008]
	rx := regexp.MustCompile(`\-\[(\d\d)\.(\d\d)\.(\d\d\d\d)\]`)
	f := rx.FindStringSubmatch(body)
	const expected = 4
	if len(f) == expected {
		y, _ := strconv.Atoi(f[3])
		m, _ := strconv.Atoi(f[1])
		d, _ := strconv.Atoi(f[2])
		return y, time.Month(m), d
	}
	// [ DATE 05-16-2006 or  [ date 09-14-2006
	rx = regexp.MustCompile(`(?i)\[ DATE (\d\d)\-(\d\d)\-(\d\d\d\d)`)
	f = rx.FindStringSubmatch(body)
	if len(f) == expected {
		y, _ := strconv.Atoi(f[3])
		m, _ := strconv.Atoi(f[1])
		d, _ := strconv.Atoi(f[2])
		return y, time.Month(m), d
	}
	//  -[2009 Jan]===========[01/28]-
	rx = regexp.MustCompile(`\-\[(\d\d\d\d) (\b\w{3,}\b)\]`)
	f = rx.FindStringSubmatch(body)
	const yyyyMMM = 3
	if len(f) == yyyyMMM {
		y, _ := strconv.Atoi(f[1])
		m := str.Month(f[2])
		return y, time.Month(m), 0
	}
	// -[XMAS2008]=
	rx = regexp.MustCompile(`\-\[XMAS(\d\d\d\d)\]=`)
	f = rx.FindStringSubmatch(body)
	const yyyy = 2
	if len(f) == yyyy {
		y, _ := strconv.Atoi(f[1])
		m := 12
		return y, time.Month(m), 0
	}
	// -[2008]===========[01/01]-
	rx = regexp.MustCompile(`\-\[(\d\d\d\d)\]`)
	f = rx.FindStringSubmatch(body)
	if len(f) == yyyy {
		y, _ := strconv.Atoi(f[1])
		return y, time.Month(0), 0
	}
	return 0, 0, 0
}

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// REL-DATE..:  01.30.2008
	rx := regexp.MustCompile(`REL-DATE..:  (\d\d)\.(\d\d)\.(\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)
	const ddmmyy = 4
	if len(f) == ddmmyy {
		y, _ := strconv.Atoi(f[3])
		m, _ := strconv.Atoi(f[1])
		d, _ := strconv.Atoi(f[2])
		return y, time.Month(m), d
	}
	return 0, 0, 0
}
