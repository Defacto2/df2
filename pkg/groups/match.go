package groups

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"
)

const (
	approx     = "≈"
	bbs        = " BBS"
	ftp        = " FTP"
	ignoreCase = "(?i)"
)

// Contains returns true whenever sorted contains x.
// The slice value must be stored in increasing order.
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

// MatchStdOut scans over the groups and attempts to match possible misnamed duplicates.
// The results are printed to stdout in realtime.
func MatchStdOut() error { //nolint:funlen
	tick := time.Now()

	const (
		n0, n1, n2, n3, n4, n5, n6, n7, n8, n9, n10, n11, n12 = 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12
	)

	list, total, err := List()
	if err != nil {
		return err
	}
	sort.Strings(list)

	l := 0
	var matches []string
	a0, a1, a2, b0, b1, b2, c0, c1, d0, d1, d2, d3, d4 := "", "", "", "", "", "", "", "", "", "", "", "", ""
	e0, e1, e2, e3, e4, e5, e6, e7, e8, e9, e10, e11, e12 := "", "", "", "", "", "", "", "", "", "", "", "", ""
	f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12 := "", "", "", "", "", "", "", "", "", "", "", ""
	g0, g1, g2, g3, g4, g5, g6, g7, g8 := "", "", "", "", "", "", "", "", ""
	h0, h1, h2, h3, h4, h5, h6, h7, h8 := "", "", "", "", "", "", "", "", ""
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
		d0 = SwapOne(group, "ph", "f")
		d1 = SwapOne(group, "ight", "ite")
		d2 = SwapOne(group, "oul", "ul")
		d3 = SwapOne(group, "ool", "ewl")
		d4 = SwapOne(group, "culd", "suld")
		e0 = SwapNumeral(group, n0)
		e1 = SwapNumeral(group, n1)
		e2 = SwapNumeral(group, n2)
		e3 = SwapNumeral(group, n3)
		e4 = SwapNumeral(group, n4)
		e5 = SwapNumeral(group, n5)
		e6 = SwapNumeral(group, n6)
		e7 = SwapNumeral(group, n7)
		e8 = SwapNumeral(group, n8)
		e9 = SwapNumeral(group, n9)
		e10 = SwapNumeral(group, n10)
		e11 = SwapNumeral(group, n11)
		e12 = SwapNumeral(group, n12)
		f1 = SwapNumeral(group, n1)
		f2 = SwapNumeral(group, n2)
		f3 = SwapNumeral(group, n3)
		f4 = SwapNumeral(group, n4)
		f5 = SwapNumeral(group, n5)
		f6 = SwapNumeral(group, n6)
		f7 = SwapNumeral(group, n7)
		f8 = SwapNumeral(group, n8)
		f9 = SwapNumeral(group, n9)
		f10 = SwapNumeral(group, n10)
		f11 = SwapNumeral(group, n11)
		f12 = SwapNumeral(group, n12)
		g0 = SwapAll(group, "0", "o")
		h0 = SwapAll(group, "o", "0")
		g1 = SwapAll(group, "1", "l")
		h1 = SwapAll(group, "l", "1")
		g2 = SwapAll(group, "1", "i")
		h2 = SwapAll(group, "i", "q")
		g3 = SwapAll(group, "i", "l")
		h3 = SwapAll(group, "l", "i")
		g4 = SwapAll(group, "3", "e")
		h4 = SwapAll(group, "e", "3")
		g5 = SwapAll(group, "4", "a")
		h5 = SwapAll(group, "a", "4")
		g6 = SwapAll(group, "6", "g")
		h6 = SwapAll(group, "g", "6")
		g7 = SwapAll(group, "8", "b")
		h7 = SwapAll(group, "b", "8")
		g8 = SwapAll(group, "9", "g")
		h8 = SwapAll(group, "g", "9")

		for _, match := range list {
			if Contains(match, matches) {
				continue
			}
			switch match {
			case group, "":
				continue
			case a0, a1, a2, b0, b1, b2, c0, c1, d0, d1, d2, d3, d4,
				e0, e1, e2, e3, e4, e5, e6, e7, e8, e9, e10, e11, e12,
				f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12,
				g0, g1, g2, g3, g4, g5, g6, g7, g8,
				h0, h1, h2, h3, h4, h5, h6, h7, h8:
				g, err1 := Count(group)
				m, err2 := Count(match)
				fmt.Fprintf(os.Stdout, "%s %s %s (%d%s%d)\n", group, approx, match,
					g, approx, m)
				if err1 != nil {
					fmt.Fprintln(os.Stdout, err1)
				}
				if err2 != nil {
					fmt.Fprintln(os.Stdout, err2)
				}
				matches = append(matches, match)
				sort.Strings(matches)
				continue
			}
		}
	}
	l = len(matches)
	elapsed := time.Since(tick)
	w := os.Stdout
	fmt.Fprintf(w, "\nProcessing time %s\n", elapsed)
	switch l {
	case 0:
		fmt.Fprintf(w, "\nGreat, there are no known duplicate names from %d groups\n", total)
	default:
		color.Primary.Printf("\n%d matches from %d groups\n", l, total)
		fmt.Fprintf(w, "To rename groups: df2 fix rename \"group name\" \"replacement name\"\n")
		fmt.Fprintf(w, "Example: df2 fix rename %q %q\n", "defacto ii", "defacto2")
	}
	return nil
}

// SwapNumeral finds any occurrences of i within a group name and swaps it for a cardinal number.
func SwapNumeral(group string, i int) string {
	num := []string{
		"zero", "one", "two", "three", "four", "five",
		"six", "seven", "eight", "nine", "ten", "eleven", "twelve",
	}
	if i > len(num) {
		return ""
	}
	re := regexp.MustCompile(strconv.Itoa(i))
	s := re.ReplaceAllString(group, num[i])
	return Format(s)
}

// SwapOrdinal finds any occurrences of i within a group name and swaps it for a ordinal number.
func SwapOrdinal(group string, i int) string {
	num := []string{
		"0", "1st", "2nd", "3rd", "4th", "5th",
		"6th", "7th", "8th", "9th", "10th", "11th", "12th",
	}
	if i > len(num) {
		return ""
	}
	re := regexp.MustCompile(strconv.Itoa(i))
	s := re.ReplaceAllString(group, num[i])
	return Format(s)
}

func replaceOne(str string) string {
	// regex source: https://stackoverflow.com/questions/16703501/replace-one-occurrence-with-regexp
	return ignoreCase + "^(.*?)" + str + "(.*)$"
}

func replOne(swap string) string {
	return "${1}" + swap + "$2"
}

// SwapOne finds the first occurrence of str within a group name and replaces it with swap.
func SwapOne(group, str, swap string) string {
	re := regexp.MustCompile(replaceOne(str))
	s := re.ReplaceAllString(group, replOne(swap))
	return Format(s)
}

// SwapAll finds all occurrences of str within a group name and replaces it with swap.
func SwapAll(group, str, swap string) string {
	re := regexp.MustCompile(ignoreCase + str)
	s := re.ReplaceAllString(group, swap)
	return Format(s)
}

// SwapPrefix replaces the prefix value at the start of a group name and replaces it with swap.
// An empty string is returned if the prefix does not exist in the name.
func SwapPrefix(group, prefix, swap string) string {
	s := ""
	prefix = strings.ToLower(prefix)
	group = strings.ToLower(group)
	if strings.HasPrefix(group, prefix) {
		s = swap + strings.TrimPrefix(group, prefix)
		return Format(s)
	}
	return ""
}

// SwapSuffix replaces the suffix value at the end of a group name and replaces it with swap.
// An empty string is returned if the suffix does not exist in the name.
func SwapSuffix(group, suffix, swap string) string {
	s := ""
	suffix = strings.ToLower(suffix)
	group = strings.ToLower(group)
	if strings.HasSuffix(group, suffix) {
		s = strings.TrimSuffix(group, suffix) + swap
		return Format(s)
	}
	suf := strings.ToLower(suffix + bbs)
	if strings.HasSuffix(group, suf) {
		s = strings.TrimSuffix(group, suf) + swap + bbs
		return Format(s)
	}
	suf = strings.ToLower(suffix + ftp)
	if strings.HasSuffix(group, suf) {
		s = strings.TrimSuffix(group, suf) + swap + ftp
		return Format(s)
	}
	return ""
}

// TrimSP removes all spaces from a named group, BBS or FTP site.
// The first string returns the spaceless name.
// The second string returns a spaceless name appended with the plural "s".
// The third string is spaceless name appended with the plural "z".
func TrimSP(name string) (string, string, string) {
	if strings.HasSuffix(name, bbs) {
		s := strings.TrimSuffix(name, bbs)
		s = strings.ReplaceAll(s, " ", "")
		s = Format(s)
		x, y, z := s+bbs, s+"s"+bbs, s+"z"+bbs
		return x, y, z
	}
	if strings.HasSuffix(name, ftp) {
		s := strings.TrimSuffix(name, ftp)
		s = strings.ReplaceAll(s, " ", "")
		s = Format(s)
		x, y, z := s+ftp, s+"s"+ftp, s+"z"+ftp
		return x, y, z
	}
	s := strings.ReplaceAll(name, " ", "")
	s = Format(s)
	x, y, z := s, s+"s", s+"z"
	return x, y, z
}
