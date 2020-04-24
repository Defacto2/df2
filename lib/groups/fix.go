package groups

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/shomali11/parallelizer"
	"gopkg.in/gookit/color.v1"
)

var simulate bool = true

// Fix any malformed group names found in the database.
func Fix(sim bool) {
	simulate = sim
	names, nc := list("")
	c := 0
	start := time.Now()
	group := parallelizer.NewGroup(parallelizer.WithJobQueueSize(nc))
	defer group.Close()
	for _, name := range names {
		group.Add(func() {
			if r := toClean(name); r {
				c++
			}
		})
	}
	err := group.Wait()
	logs.Check(err)
	switch {
	case c > 0 && simulate:
		logs.Printcr(c, "fixes required")
		logs.Simulate()
	case c == 1:
		logs.Printcr("1 fix applied")
	case c > 0:
		logs.Printcr(c, "fixes applied")
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
	if len(s) < 2 {
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
	a := strings.Split(g, " ")
	if len(a) < 2 {
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
	if len(s) < 4 {
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
func toClean(g string) bool {
	f := ToClean(g)
	if f == g {
		return false
	}
	if simulate {
		logs.Printf("%s %q %s %s\n", color.Question.Sprint("?"), g, color.Question.Sprint("!="), color.Info.Sprint(f))
		return true
	}
	s := logs.Y()
	r := true
	var _, err = database.RenGroup(f, g)
	if err != nil {
		s = logs.X()
		r = false
	}
	logs.Printf("%s %q %s %s\n", s, g, color.Question.Sprint("⟫"), color.Info.Sprint(f))
	return r
}
