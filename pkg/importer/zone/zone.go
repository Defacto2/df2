package zone

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/str"
)

const (
	Name     = "ZONE"
	december = 12
	dd       = 2
	mm       = 1
	yy       = 3
)

func DizDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}
	const (
		type1 = `-=?=\[(\d\d)\/(\d\d)\/(\d\d)\]=?=-` // -=[07/15/00]=- or -==[11/17/00]==-
		type2 = `\[DATE\:(\d?\d)\/(\d?\d)\/(\d\d)\]` // [DATE:06/24/00] or [DATE:2/4/00]
		type3 = `\[(\d\d)\.(\d\d)\.(\d\d)\]`         // [03.28.02]
	)
	regs := []string{type1, type2, type3}
	for _, exp := range regs {
		rx := regexp.MustCompile(exp)
		f := rx.FindStringSubmatch(body)
		const expected = 4
		if len(f) != expected {
			continue
		}
		y := str.YearAbbr(f[yy])
		m, _ := strconv.Atoi(f[mm])
		d, _ := strconv.Atoi(f[dd])
		if m > december {
			// ZONE releases use both DD/MM & MM/DD formats
			return y, time.Month(d), m
		}
		return y, time.Month(m), d
	}
	return 0, 0, 0
}

func DizTitle(body string) ( //nolint:nonamedreturns
	title string,
) {
	if body == "" {
		return ""
	}
	body = strings.TrimSpace(body)

	t := ""
	scanner := bufio.NewScanner(strings.NewReader(body))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		if len(s) == 0 {
			continue
		}
		if t == "" {
			t = s
			continue
		}
	}
	// hack to handle edge cases.
	if strings.HasPrefix(t, "T-Racks") {
		return t
	}
	return str.PathTitle(t)
}

func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// matches
	// DATE  : 05.29.02
	// [DATE: 02/05/01]
	// Date:  03/07/2000
	// date: 2-24-2000
	rx := regexp.MustCompile(`(?i)DATE[ ]{0,}:[ ]{1,}` +
		`(\d?\d)[\/|\.|\-](\d?\d)[\/|\.|\-](\d\d\d?\d?)`)
	f := rx.FindStringSubmatch(body)

	const expected = 4
	if len(f) != expected {
		return 0, 0, 0
	}
	y, _ := strconv.Atoi(f[yy])
	if len(f[3]) == 2 {
		y = str.YearAbbr(f[yy])
	}
	m, _ := strconv.Atoi(f[mm])
	d, _ := strconv.Atoi(f[dd])
	if m > december {
		// ZONE releases use both DD/MM & MM/DD formats
		return y, time.Month(d), m
	}
	return y, time.Month(m), d
}
