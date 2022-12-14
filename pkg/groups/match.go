package groups

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	approx = "≈"
	bbs    = " BBS"
	ftp    = " FTP"
)

// Contains returns true whenever sorted contains x.
func Contains(x string, sorted []string) bool {
	l := len(sorted)
	if l == 0 {
		return false
	}
	o := sort.SearchStrings(sorted, x)
	if o >= l {
		return false
	}
	return sorted[o] == x
}

func MatchStd() error {
	// defer profile.Start(
	// 	profile.CPUProfile,
	// 	//profile.GoroutineProfile,
	// 	//profile.MemProfileHeap,
	// 	//profile.MemProfileAllocs,
	// 	profile.ProfilePath(".")).Stop()

	tick := time.Now()

	list, total, err := List()
	if err != nil {
		return err
	}
	sort.Strings(list)

	l := 0
	var matches []string
	a0, a1, a2, b0, b1, b2, c0, c1, d0, d1, d2, d3, d4 :=
		"", "", "", "", "", "", "", "", "", "", "", "", ""

	for _, group := range list {
		l = len(group)
		if l == 0 {
			continue
		}

		a0 = SwapSuffix(group, "s", "z")
		a1 = group + "s"
		a2 = group + "z"
		b0, b1, b2 = TrimSP(group)
		c0 = SwapSuffix(group, "er", "a")
		c1 = SwapPrefix(group, "th", "da")
		d0 = SwapPhonetic(group, "ph", "f")
		d1 = SwapPhonetic(group, "ight", "ite")
		d2 = SwapPhonetic(group, "oul", "ul")
		d3 = SwapPhonetic(group, "ool", "ewl")
		d4 = SwapPhonetic(group, "culd", "suld")

		for _, match := range list {
			if Contains(match, matches) {
				continue
			}
			switch match {
			case group, "":
				continue
			case a0, a1, a2, b0, b1, b2, c0, c1, d0, d1, d2, d3, d4:
				fmt.Printf("%q %s %q\n", group, approx, match)
				matches = append(matches, match)
				sort.Strings(matches)
				continue
			}
		}
	}
	list = nil
	l = len(matches)
	switch l {
	case 0:
		fmt.Printf("\nGreat, there are no known duplicate names from %d groups\n", total)
	default:
		fmt.Printf("\n%d matches from %d groups\n", l, total)
	}
	elapsed := time.Since(tick)
	log.Printf("taking %s", elapsed)
	return nil
}

func SwapPhonetic(group, phonetic, swap string) string {
	re := regexp.MustCompile(phonetic)
	s := re.ReplaceAllString(group, swap)
	return Format(s)
}

func SwapPrefix(group, prefix, swap string) string {
	prefix = Format(prefix)
	if strings.HasPrefix(group, prefix) {
		return strings.TrimPrefix(group, prefix) + swap
	}
	if strings.HasPrefix(group, prefix+bbs) {
		return strings.TrimPrefix(group, prefix+bbs) + swap + bbs
	}
	if strings.HasPrefix(group, prefix+ftp) {
		return strings.TrimPrefix(group, prefix+ftp) + swap + ftp
	}
	return ""
}

func SwapSuffix(group, suffix, swap string) string {
	if strings.HasSuffix(group, suffix) {
		return strings.TrimSuffix(group, suffix) + swap
	}
	if strings.HasSuffix(group, suffix+bbs) {
		return strings.TrimSuffix(group, suffix+bbs) + swap + bbs
	}
	if strings.HasSuffix(group, suffix+ftp) {
		return strings.TrimSuffix(group, suffix+ftp) + swap + ftp
	}
	return ""
}

func TrimSP(group string) (string, string, string) {
	if strings.HasSuffix(group, bbs) {
		s := strings.TrimSuffix(group, bbs)
		s = strings.ReplaceAll(s, " ", "")
		s = Format(s)
		x, y, z := s+bbs, s+"s"+bbs, s+"z"+bbs
		return x, y, z
	}
	if strings.HasSuffix(group, ftp) {
		s := strings.TrimSuffix(group, ftp)
		s = strings.ReplaceAll(s, " ", "")
		s = Format(s)
		x, y, z := s+ftp, s+"s"+ftp, s+"z"+ftp
		return x, y, z
	}
	s := strings.ReplaceAll(group, " ", "")
	s = Format(s)
	x, y, z := s, s+"s", s+"z"
	return x, y, z
}
