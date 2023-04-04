// Package zwt handles the .rar for the group Zero Waiting Time (ZWT).
package zwt

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const Name = "Zero Waiting Time"

func DizDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	rx := regexp.MustCompile(`\[(\d\d\d\d)\-(\d\d)\-(\d\d)\]`)
	f := rx.FindStringSubmatch(body)

	const expected = 4
	if len(f) != expected {
		return 0, 0, 0
	}
	y, _ := strconv.Atoi(f[1])
	m, _ := strconv.Atoi(f[2])
	d, _ := strconv.Atoi(f[3])
	return y, time.Month(m), d
}

func DizTitle(body string) ( //nolint:nonamedreturns
	title string, pub string,
) {
	if body == "" {
		return "", ""
	}
	t, p := "", ""
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
		if p == "" {
			p = s
			break
		}
	}
	return t, p
}
