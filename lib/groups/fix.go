package groups

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"gopkg.in/gookit/color.v1"
)

var sim bool = true

// Fix any malformed group names found in the database.
func Fix(simulate bool) {
	sim = simulate
	names, _, err := list("")
	logs.Check(err)
	c := 0
	start := time.Now()
	for _, name := range names {
		if r := toClean(name); r {
			c++
		}
	}
	switch {
	case c > 0 && simulate:
		logs.Printcrf("%d fixes required", c)
		logs.Simulate()
	case c == 1:
		logs.Printcr("1 fix applied")
	case c > 0:
		logs.Printcrf("%d fixes applied", c)
	default:
		logs.Printcr("no fixes needed")
	}
	elapsed := time.Since(start)
	logs.Print(fmt.Sprintf(", time taken %s\n", elapsed))
}

// ToClean fixes any malformed strings.
func ToClean(g string) string {
	f := remDupeSpaces(g)
	f = strings.TrimSpace(f)
	f = dropThe(f)
	f = format(f)
	return f
}

// dropDot removes any trailing dots from a string.
func dropDot(s string) string {
	const short = 2
	if len(s) < short {
		return s
	}
	l := s[len(s)-1:]
	if l == "." {
		return s[:len(s)-1]
	}
	return s
}

// dropThe removes a 'the' prefix from a string.
func dropThe(g string) string {
	const short = 2
	a := strings.Split(g, " ")
	if len(a) < short {
		return g
	}
	l := a[len(a)-1]
	if strings.ToLower(a[0]) == "the" && (l == "BBS" || l == "FTP") {
		return strings.Join(a[1:], " ") // drop "the" prefix
	}
	return g
}

// format returns a copy of the string with custom formatting.
func format(s string) string {
	const acronym = 3
	if len(s) <= acronym {
		return strings.ToUpper(s)
	}
	groups := strings.Split(s, ",")
	for j, g := range groups {
		words := strings.Split(g, " ")
		last := len(words) - 1
		for i, w := range words {
			w = strings.ToLower(w)
			w = dropDot(w)
			if i > 0 && i < last {
				switch w {
				case "a", "and", "by", "of", "for", "from", "in", "is", "or", "the", "to":
					words[i] = strings.ToLower(w)
					continue
				}
			}
			switch w {
			case "3d", "ansi", "bbs", "cd", "cgi", "dox", "eu", "ftp", "fx", "hq", "id", "ii", "iii", "iso", "pc", "pcb", "pda", "st", "uk", "us", "uss", "ussr", "vcd", "whq":
				words[i] = strings.ToUpper(w)
				continue
			}
			words[i] = strings.Title(w)
		}
		groups[j] = strings.Join(words, " ")
	}
	return strings.Join(groups, ",")
}

// remDupeSpaces removes duplicate spaces from a string.
func remDupeSpaces(s string) string {
	r := regexp.MustCompile(`\s+`)
	return r.ReplaceAllString(s, " ")
}

// toClean fixes and saves a malformed group name.
func toClean(g string) (ok bool) {
	f := ToClean(g)
	if f == g {
		return false
	}
	if sim {
		logs.Printf("\n%s %q %s %s", color.Question.Sprint("?"), g, color.Question.Sprint("!="), color.Info.Sprint(f))
		return true
	}
	s := logs.Y()
	ok = true
	var _, err = database.RenGroup(f, g)
	if err != nil {
		s = logs.X()
		ok = false
	}
	logs.Printf("\n%s %q %s %s", s, g, color.Question.Sprint("⟫"), color.Info.Sprint(f))
	return ok
}
