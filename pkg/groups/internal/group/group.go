package group

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups/internal/acronym"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/gookit/color"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var ErrFilter = errors.New("invalid filter used")

const (
	bbs   = "bbs"
	ftp   = "ftp"
	grp   = "group"
	mag   = "magazine"
	space = " "
)

// Filter group by role or function.
type Filter int

const (
	None     Filter = iota // None returns all groups.
	BBS                    // BBS boards.
	FTP                    // FTP sites.
	Group                  // Group generic roles.
	Magazine               // Magazine publishers.
)

func (f Filter) String() string {
	switch f {
	case BBS:
		return bbs
	case FTP:
		return ftp
	case Group:
		return grp
	case Magazine:
		return mag
	case None:
		return ""
	}
	return ""
}

// Get the Filter type from s.
func Get(s string) Filter {
	switch strings.ToLower(s) {
	case bbs:
		return BBS
	case ftp:
		return FTP
	case grp:
		return Group
	case mag:
		return Magazine
	case "":
		return None
	}
	return -1
}

// List all organisations or groups filtered by s.
func List(s string) (groups []string, total int, err error) {
	db := database.Connect()
	defer db.Close()
	r, err := SQLSelect(Get(s), false)
	if err != nil {
		return nil, 0, fmt.Errorf("list statement: %w", err)
	}
	total, err = database.Total(&r)
	if err != nil {
		return nil, 0, fmt.Errorf("list total: %w", err)
	}
	// interate through records
	rows, err := db.Query(r)
	if err != nil {
		return nil, 0, fmt.Errorf("list query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list rows: %w", rows.Err())
	}
	defer rows.Close()
	var grp sql.NullString
	for rows.Next() {
		if err = rows.Scan(&grp); err != nil {
			return nil, 0, fmt.Errorf("list rows scan: %w", err)
		}
		if _, err = grp.Value(); err != nil {
			continue
		}
		groups = append(groups, fmt.Sprintf("%v", grp.String))
	}
	return groups, total, nil
}

// SQLSelect returns a complete SQL WHERE statement where the groups are filtered.
func SQLSelect(f Filter, includeSoftDeletes bool) (string, error) {
	inc, skip := includeSoftDeletes, false
	if f > -1 {
		skip = true
	}
	where, err := SQLWhere(f, inc)
	if err != nil {
		return "", fmt.Errorf("sql select %q: %w", f.String(), err)
	}
	var s string
	switch skip {
	case true: // disable group_brand_by listings for BBS, FTP, group, magazine filters
		s = "SELECT DISTINCT group_brand_for AS pubCombined " +
			"FROM files WHERE Length(group_brand_for) <> 0 " + where
	default:
		s = "(SELECT DISTINCT group_brand_for AS pubCombined " +
			"FROM files WHERE Length(group_brand_for) <> 0 " + where + ")" +
			" UNION " +
			"(SELECT DISTINCT group_brand_by AS pubCombined " +
			"FROM files WHERE Length(group_brand_by) <> 0 " + where + ")"
	}
	return s + " ORDER BY pubCombined", nil
}

// SQLWhere returns a partial SQL WHERE statement where groups are filtered.
func SQLWhere(f Filter, softDel bool) (string, error) {
	s, err := SQLFilter(f)
	if err != nil {
		return "", fmt.Errorf("sql where: %w", err)
	}
	switch {
	case s != "" && softDel:
		s = "AND " + s
	case s == "" && softDel: // do nothing
	case s != "" && !softDel:
		s = "AND " + s + " `deletedat` IS NULL"
	default:
		s = "AND `deletedat` IS NULL"
	}
	const andLen = 4
	l := len(s)
	if l > andLen && s[l-andLen:] == " AND" {
		logs.Printf("%q|", s[l-andLen:])
		return s[:l-andLen], nil
	}
	return s, nil
}

// SQLFilter returns a partial SQL WHERE statement to filter groups.
func SQLFilter(f Filter) (string, error) {
	var s string
	switch f {
	case None:
		return "", nil
	case Magazine:
		s = "section = 'magazine' AND"
	case BBS:
		s = "RIGHT(group_brand_for,4) = ' BBS' AND"
	case FTP:
		s = "RIGHT(group_brand_for,4) = ' FTP' AND"
	case Group: // only display groups who are listed under group_brand_for, group_brand_by only groups will be ignored
		s = "RIGHT(group_brand_for,4) != ' FTP' AND RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine' AND"
	default:
		return "", fmt.Errorf("sql filter %q: %w", f.String(), ErrFilter)
	}
	return s, nil
}

// Request flags for group functions.
type Request struct {
	Filter      string // Filter groups by category.
	Counts      bool   // Counts the group's total files.
	Initialisms bool   // Initialisms and acronyms for groups.
	Progress    bool   // Progress counter when requesting database data.
}

// Clean a malformed group name and save the fix to the database.
func Clean(name string, sim bool) (ok bool) {
	f := CleanS(name)
	if f == name {
		return false
	}
	if sim {
		logs.Printf("\n%s %q %s %s", color.Question.Sprint("?"), name,
			color.Question.Sprint("!="), color.Info.Sprint(f))
		return true
	}
	s := str.Y()
	ok = true
	if _, err := rename(f, name); err != nil {
		s = str.X()
		ok = false
	}
	logs.Printf("\n%s %q %s %s", s, name, color.Question.Sprint("⟫"), color.Info.Sprint(f))
	return ok
}

// CleanS fixes the malformed string.
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
	title := cases.Title(language.English, cases.NoLower)
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
			words[i] = title.String(w)
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
		return 0, fmt.Errorf("rename statement: %w", err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(newName, newName, group, group)
	if err != nil {
		return 0, fmt.Errorf("rename exec: %w", err)
	}
	count, err = res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rename rows affected: %w", err)
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

// Use a HR element to mark out the groups alphabetically.
func UseHr(prevLetter, group string) (string, bool) {
	if group == "" {
		return "", false
	}
	switch g := group[:1]; {
	case prevLetter == "":
		return g, false
	case prevLetter != g:
		return g, true
	}
	return prevLetter, false
}

// Slug takes a string and makes it into a URL friendly slug.
func Slug(s string) string {
	n := database.TrimSP(s)
	n = acronym.Trim(n)
	n = strings.ReplaceAll(n, "-", "_")
	n = strings.ReplaceAll(n, ", ", "*")
	n = strings.ReplaceAll(n, " & ", " ampersand ")
	re := regexp.MustCompile(` (\d)`)
	n = re.ReplaceAllString(n, `-$1`)
	re = regexp.MustCompile(`[^A-Za-z0-9 \-\+.\_\*]`) // remove all chars except these
	n = re.ReplaceAllString(n, ``)
	n = strings.ToLower(n)
	re = regexp.MustCompile(` ([a-z])`)
	n = re.ReplaceAllString(n, `-$1`)
	return n
}

// Count returns the number of file entries associated with a named group.
func Count(name string) (int, error) {
	if name == "" {
		return 0, nil
	}
	db := database.Connect()
	defer db.Close()
	n, count := name, 0
	row := db.QueryRow("SELECT COUNT(*) FROM files WHERE group_brand_for=? OR "+
		"group_brand_for LIKE '?,%%' OR group_brand_for LIKE '%%, ?,%%' OR "+
		"group_brand_for LIKE '%%, ?' OR group_brand_by=? OR group_brand_by "+
		"LIKE '?,%%' OR group_brand_by LIKE '%%, ?,%%' OR group_brand_by LIKE '%%, ?'", n, n)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, db.Close()
}
