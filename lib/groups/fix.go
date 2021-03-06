package groups

import (
	"fmt"
	"strings"
	"time"

	"github.com/gookit/color" //nolint:misspell

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
)

const (
	bbs = "bbs"
	ftp = "ftp"
)

// Fix any malformed group names found in the database.
func Fix(simulate bool) error {
	names, _, err := list("")
	if err != nil {
		return err
	}
	c, start := 0, time.Now()
	for _, name := range names {
		if r := cleanGroup(name, simulate); r {
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
		logs.Printcr("no group fixes needed")
	}
	elapsed := time.Since(start).Seconds()
	logs.Print(fmt.Sprintf(", time taken %.1f seconds\n", elapsed))
	return nil
}

// cleanGroup fixes and saves a malformed group name.
func cleanGroup(g string, sim bool) (ok bool) {
	f := cleanString(g)
	if f == g {
		return false
	}
	if sim {
		logs.Printf("\n%s %q %s %s", color.Question.Sprint("?"), g,
			color.Question.Sprint("!="), color.Info.Sprint(f))
		return true
	}
	s := str.Y()
	ok = true
	if _, err := rename(f, g); err != nil {
		s = str.X()
		ok = false
	}
	logs.Printf("\n%s %q %s %s", s, g, color.Question.Sprint("⟫"), color.Info.Sprint(f))
	return ok
}

// cleanString fixes any malformed strings.
func cleanString(s string) string {
	f := database.TrimSP(s)
	f = database.StripChars(f)
	f = database.StripStart(f)
	f = strings.TrimSpace(f)
	f = trimThe(f)
	f = format(f)
	return f
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
			w = trimDot(w)
			if i > 0 && i < last {
				switch w {
				case "a", "and", "by", "of", "for", "from", "in", "is", "or", "the", "to":
					words[i] = strings.ToLower(w)
					continue
				}
			}
			switch w {
			case "3d", "ansi", bbs, "cd", "cgi", "dox", "eu", ftp, "fx", "hq",
				"id", "ii", "iii", "iso", "pc", "pcb", "pda", "st", "uk", "us",
				"uss", "ussr", "vcd", "whq":
				words[i] = strings.ToUpper(w)
				continue
			}
			words[i] = strings.Title(w)
		}
		groups[j] = strings.Join(words, " ")
	}
	return strings.Join(groups, ",")
}

// rename replaces all instances of the group name with a new group name.
func rename(replacement, group string) (count int64, err error) {
	db := database.Connect()
	defer db.Close()
	stmt, err := db.Prepare("UPDATE `files` SET group_brand_for=?, group_brand_by=? WHERE (group_brand_for=? OR group_brand_by=?)")
	if err != nil {
		return 0, fmt.Errorf("rename group statement: %w", err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(replacement, replacement, group, group)
	if err != nil {
		return 0, fmt.Errorf("rename group exec: %w", err)
	}
	count, err = res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rename group rows affected: %w", err)
	}
	return count, db.Close()
}

// trimDot removes any trailing dots from a string.
func trimDot(s string) string {
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

// trimThe removes a 'the' prefix from a string.
func trimThe(s string) string {
	const short = 2
	a := strings.Split(s, " ")
	if len(a) < short {
		return s
	}
	l := a[len(a)-1]
	if strings.EqualFold(a[0], "the") && (l == "BBS" || l == "FTP") {
		return strings.Join(a[1:], " ") // drop "the" prefix
	}
	return s
}
