package air

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const Name = "AiR"

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// DATE ......: 10/2008
	// DATE......: 10/2006
	rx := regexp.MustCompile(`DATE[ ]{0,}......: (\d?\d)\/(\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)

	const mmyyyy = 3
	if len(f) == mmyyyy {
		y, _ := strconv.Atoi(f[2])
		m, _ := strconv.Atoi(f[1])
		d := 0
		return y, time.Month(m), d
	}

	// released....: 19 March, 1999
	rx = regexp.MustCompile(`released....: (\d\d) (\b\w{3,}\b), (\d\d\d\d)`)
	f = rx.FindStringSubmatch(body)

	const ddMMMyyyy = 4
	fmt.Println(f)
	if len(f) == ddMMMyyyy {
		y, _ := strconv.Atoi(f[3])
		m := Month(f[2])
		d, _ := strconv.Atoi(f[1])
		return y, time.Month(m), d
	}

	return 0, 0, 0
}

func Month(month string) int {
	if len(month) < 3 {
		return 0
	}
	const dec = 12
	for i := 1; i <= dec; i++ {
		mon := strings.ToLower(month)[0:3]
		mmm := strings.ToLower(time.Month(i).String())[0:3]
		if mon == mmm {
			return i
		}
	}
	return 0
}
