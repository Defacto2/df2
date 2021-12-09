package group

import (
	"fmt"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/gookit/color"
)

const (
	bbs   = "bbs"
	ftp   = "ftp"
	space = " "
)

// Request flags for group functions.
type Request struct {
	Filter      string // Filter groups by category.
	Counts      bool   // Counts the group's total files.
	Initialisms bool   // Initialisms and acronyms for groups.
	Progress    bool   // Progress counter when requesting database data.
}

// cleanGroup fixes and saves a malformed group name.
func Clean(g string, sim bool) (ok bool) {
	f := CleanS(g)
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
func CleanS(s string) string {
	f := database.TrimSP(s)
	f = database.StripChars(f)
	f = database.StripStart(f)
	f = strings.TrimSpace(f)
	f = TrimThe(f)
	f = Format(f)
	return f
}

// Format returns a copy of s with custom formatting.
func Format(s string) string {
	const acronym = 3
	if len(s) <= acronym {
		return strings.ToUpper(s)
	}
	groups := strings.Split(s, ",")
	for j, g := range groups {
		words := strings.Split(g, space)
		last := len(words) - 1
		for i, w := range words {
			w = strings.ToLower(w)
			w = TrimDot(w)
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
		groups[j] = strings.Join(words, space)
	}
	return strings.Join(groups, ",")
}

// rename replaces all instances of the group name with a new group name.
func rename(newName, group string) (count int64, err error) {
	db := database.Connect()
	defer db.Close()
	stmt, err := db.Prepare("UPDATE `files` SET group_brand_for=?," +
		" group_brand_by=? WHERE (group_brand_for=? OR group_brand_by=?)")
	if err != nil {
		return 0, fmt.Errorf("rename group statement: %w", err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(newName, newName, group, group)
	if err != nil {
		return 0, fmt.Errorf("rename group exec: %w", err)
	}
	count, err = res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rename group rows affected: %w", err)
	}
	return count, db.Close()
}

// TrimDot removes any trailing dots from s.
func TrimDot(s string) string {
	const short = 2
	if len(s) < short {
		return s
	}
	if l := s[len(s)-1:]; l == "." {
		return s[:len(s)-1]
	}
	return s
}

// TrimThe removes 'the' prefix from s.
func TrimThe(s string) string {
	const short = 2
	a := strings.Split(s, space)
	if len(a) < short {
		return s
	}
	l := a[len(a)-1]
	if strings.EqualFold(a[0], "the") && (l == "BBS" || l == "FTP") {
		return strings.Join(a[1:], space) // drop "the" prefix
	}
	return s
}
