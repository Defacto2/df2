package zwt

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const Name = "Zero Waiting Time"

func DizDate(r io.Reader) (year int, month time.Month, day int) {
	if r == nil {
		return 0, 0, 0
	}
	buf := &strings.Builder{}
	_, _ = io.Copy(buf, r)

	rx := regexp.MustCompile(`\[(\d\d\d\d)\-(\d\d)\-(\d\d)\]`)
	f := rx.FindStringSubmatch(buf.String())

	const expected = 4
	if len(f) != expected {
		return 0, 0, 0
	}
	y, _ := strconv.Atoi(f[1])
	m, _ := strconv.Atoi(f[2])
	d, _ := strconv.Atoi(f[3])
	return y, time.Month(m), d
}

func DizTitle(r io.Reader) (title string, pub string) {
	if r == nil {
		return "", ""
	}
	scanner := bufio.NewScanner(r)
	t, p := "", ""
	for scanner.Scan() {
		s := scanner.Text()
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			continue
		}
		if t == "" {
			t = strings.TrimSpace(s)
			continue
		}
		if p == "" {
			p = strings.TrimSpace(s)
			break
		}
	}
	return t, p
}
