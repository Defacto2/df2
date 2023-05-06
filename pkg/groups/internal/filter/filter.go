// Package filter returns the groups based on a role or activity filter.
package filter

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups/internal/acronym"
)

var ErrFilter = errors.New("invalid filter used")

const (
	bbs = "bbs"
	ftp = "ftp"
	grp = "group"
	mag = "magazine"
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

// Count returns the number of file entries associated with a named group.
func Count(db *sql.DB, name string) (int, error) {
	if db == nil {
		return 0, database.ErrDB
	}
	if name == "" {
		return 0, nil
	}
	n, count := name, 0
	row := db.QueryRow("SELECT COUNT(*) FROM files WHERE group_brand_for=? OR "+
		"group_brand_for LIKE '?,%%' OR group_brand_for LIKE '%%, ?,%%' OR "+
		"group_brand_for LIKE '%%, ?' OR group_brand_by=? OR group_brand_by "+
		"LIKE '?,%%' OR group_brand_by LIKE '%%, ?,%%' OR group_brand_by LIKE '%%, ?'", n, n)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}

// List all organisations or groups filtered by the named string.
func List(db *sql.DB, w io.Writer, name string) ([]string, int, error) {
	if db == nil {
		return nil, 0, database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	r, err := SQLSelect(w, Get(name), false)
	if err != nil {
		return nil, 0, fmt.Errorf("list statement: %w", err)
	}
	total, err := database.Total(db, w, &r)
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
	grp := sql.NullString{}
	groups := []string{}
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
func SQLSelect(w io.Writer, f Filter, includeSoftDeletes bool) (string, error) {
	if w == nil {
		w = io.Discard
	}
	where, err := SQLWhere(w, f, includeSoftDeletes)
	if err != nil {
		return "", fmt.Errorf("sql select %q: %w", f.String(), err)
	}
	s := "(SELECT DISTINCT group_brand_for AS pubCombined " +
		"FROM files WHERE Length(group_brand_for) <> 0 " + where + ")" +
		" UNION " +
		"(SELECT DISTINCT group_brand_by AS pubCombined " +
		"FROM files WHERE Length(group_brand_by) <> 0 " + where + ")"
	switch f {
	case BBS, FTP, Group, Magazine:
		s = "SELECT DISTINCT group_brand_for AS pubCombined " +
			"FROM files WHERE Length(group_brand_for) <> 0 " + where
	case None:
	default:
	}
	return s + " ORDER BY pubCombined", nil
}

// SQLWhere returns a partial SQL WHERE statement where groups are filtered.
func SQLWhere(w io.Writer, f Filter, includeSoftDeletes bool) (string, error) {
	if w == nil {
		w = io.Discard
	}
	deleted := includeSoftDeletes
	s, err := SQLFilter(f)
	if err != nil {
		return "", fmt.Errorf("sql where: %w", err)
	}
	switch {
	case s != "" && deleted:
		s = "AND " + s
	case s == "" && deleted: // do nothing
	case s != "" && !deleted:
		s = "AND " + s + " `deletedat` IS NULL"
	default:
		s = "AND `deletedat` IS NULL"
	}
	const andLen = 4
	l := len(s)
	if l > andLen && s[l-andLen:] == " AND" {
		fmt.Fprintf(w, "%q|", s[l-andLen:])
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
