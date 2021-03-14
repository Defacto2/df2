// Package groups deals with group names and their initialisms.
package groups

import (
	"bufio"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/spf13/viper"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
)

const htm = ".htm"

// Filter group by role or function.
type Filter int

const (
	// None returns all groups.
	None Filter = iota
	// BBS boards.
	BBS
	// FTP sites.
	FTP
	// Group generic roles.
	Group
	// Magazine publishers.
	Magazine
)

func (f Filter) String() string {
	switch f {
	case BBS:
		return "bbs"
	case FTP:
		return "ftp"
	case Group:
		return "group"
	case Magazine:
		return "magazine"
	case None:
		return ""
	}
	return ""
}

func filter(s string) Filter {
	switch strings.ToLower(s) {
	case "bbs":
		return BBS
	case "ftp":
		return FTP
	case "group":
		return Group
	case "magazine":
		return Magazine
	case "":
		return None
	}
	return -1
}

var ErrFilter = errors.New("invalid filter used")

// Request flags for group functions.
type Request struct {
	// Filter groups by category.
	Filter string
	// Counts the group's total files.
	Counts bool
	// Initialisms and acronyms for groups.
	Initialisms bool
	// Progress counter when requesting database data.
	Progress bool
}

// Result on a group.
type Result struct {
	// ID used in URLs to link to the group.
	ID string
	// Name of the group.
	Name string
	// Count file totals.
	Count int
	// Initialism or acronym.
	Initialism string
	// Inject a HR element to separate a collection of groups.
	Hr bool
}

// Count returns the number of file entries associated with a group.
func Count(name string) (count int, err error) {
	db := database.Connect()
	defer db.Close()
	n := name
	row := db.QueryRow("SELECT COUNT(*) FROM files WHERE group_brand_for=? OR "+
		"group_brand_for LIKE '?,%%' OR group_brand_for LIKE '%%, ?,%%' OR "+
		"group_brand_for LIKE '%%, ?' OR group_brand_by=? OR group_brand_by "+
		"LIKE '?,%%' OR group_brand_by LIKE '%%, ?,%%' OR group_brand_by LIKE '%%, ?'", n, n)
	if err = row.Scan(&count); err != nil {
		return 0, fmt.Errorf("group count row scan: %w", err)
	}
	return count, db.Close()
}

// Cronjob is used for system automation to generate dynamic HTML pages.
func Cronjob(force bool) error {
	for _, tag := range Wheres() {
		last, err := database.LastUpdate()
		if err != nil {
			return fmt.Errorf("group cronjob last update: %w", err)
		}
		f := tag + htm
		n := path.Join(viper.GetString("directory.html"), f)
		update := true
		if !force {
			update, err = database.FileUpdate(n, last)
		}
		switch {
		case err != nil:
			return fmt.Errorf("group cronjob file update: %w", err)
		case !update:
			logs.Println(f + " has nothing to update")
		default:
			r := Request{tag, true, true, false}
			if err := r.HTML(f); err != nil {
				return fmt.Errorf("group cronjob html: %w", err)
			}
		}
	}
	return nil
}

// DataList prints an auto-complete list for HTML input elements.
func (r Request) DataList(filename string) error {
	// <option value="Bitchin ANSI Design" label="BAD (Bitchin ANSI Design)">
	tpl := `{{range .}}{{if .Initialism}}<option value="{{.Name}}" label="{{.Initialism}} ({{.Name}})">{{end}}<option value="{{.Name}}" label="{{.Name}}">{{end}}`
	if err := r.parse(filename, tpl); err != nil {
		return fmt.Errorf("datalist parse template: %w", err)
	}
	return nil
}

// HTML prints a snippet listing links to each group, with an optional file count.
func (r Request) HTML(filename string) error {
	// <h2><a href="/g/13-omens">13 OMENS</a> 13O</h2><hr>
	tpl := `{{range .}}{{if .Hr}}<hr>{{end}}<h2><a href="/g/{{.ID}}">{{.Name}}</a>{{if .Initialism}} ({{.Initialism}}){{end}}{{if .Count}} <small>({{.Count}})</small>{{end}}</h2>{{end}}`
	if err := r.parse(filename, tpl); err != nil {
		return fmt.Errorf("html parse template: %w", err)
	}
	return nil
}

func hrElement(letter, group string) (string, bool) {
	hr := false
	if group == "" {
		return "", hr
	}
	switch g := group[:1]; {
	case letter == "":
		letter = g
	case letter != g:
		letter = g
		hr = true
	}
	return letter, hr
}

func (r Request) files(group string) (total int, err error) {
	if r.Counts {
		total, err = Count(group)
		if err != nil {
			return 0, fmt.Errorf("request total files for %q: %w", group, err)
		}
		return total, nil
	}
	return 0, nil
}

func (r Request) initialism(group string) (name string, err error) {
	if r.Initialisms {
		name, err = initialism(group)
		if err != nil {
			return "", fmt.Errorf("request initialism for %q: %w", group, err)
		}
		return name, nil
	}
	return "", nil
}

func (r Request) iterate(groups ...string) (g *[]Result, err error) {
	piped := str.Piped()
	total := len(groups)
	data := make([]Result, total)
	lastLetter, hr := "", false
	for i, grp := range groups {
		if !piped && !logs.Quiet && r.Progress {
			str.Progress(r.Filter, i+1, total)
		}
		lastLetter, hr = hrElement(lastLetter, grp)
		c, err := r.files(grp)
		if err != nil {
			return nil, fmt.Errorf("iterate group file %q: %w", grp, err)
		}
		init, err := r.initialism(grp)
		if err != nil {
			return nil, fmt.Errorf("iterate group initialism %q: %w", grp, err)
		}
		data[i] = Result{
			ID:         MakeSlug(grp),
			Name:       grp,
			Count:      c,
			Initialism: init,
			Hr:         hr,
		}
	}
	return &data, nil
}

func (r Request) parse(filename, templ string) (err error) {
	groups, total, err := list(r.Filter)
	if err != nil {
		return fmt.Errorf("parse group: %w", err)
	}
	if !str.Piped() {
		if f := r.Filter; f == "" {
			logs.Println(total, "matching (all) records found")
		} else {
			logs.Println(total, "matching", f, "records found")
		}
	}
	data, err := r.iterate(groups...)
	if err != nil {
		return fmt.Errorf("parse iterate: %w", err)
	}
	t, err := template.New("h2").Parse(templ)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	if filename == "" {
		var buf bytes.Buffer
		wr := bufio.NewWriter(&buf)
		if err = t.Execute(wr, &data); err != nil {
			return fmt.Errorf("parse template execute: %w", err)
		}
		if err := wr.Flush(); err != nil {
			return fmt.Errorf("parse writer flush: %w", err)
		}
		fmt.Println(buf.String())
		return nil
	}
	switch filter(r.Filter) {
	case BBS, FTP, Group, Magazine:
		f, err := os.Create(path.Join(viper.GetString("directory.html"), filename))
		if err != nil {
			return fmt.Errorf("parse create: %w", err)
		}
		defer f.Close()
		if err = t.Execute(f, &data); err != nil {
			return fmt.Errorf("parse t execute: %w", err)
		}
	case None:
		return fmt.Errorf("parse %q: %w", r.Filter, ErrFilter)
	}
	return nil
}

// list all organizations or filtered groups.
func list(f string) (groups []string, total int, err error) {
	db := database.Connect()
	defer db.Close()
	s, err := groupsStmt(filter(f), false)
	if err != nil {
		return nil, 0, fmt.Errorf("list groups statement: %w", err)
	}
	total, err = database.Total(&s)
	if err != nil {
		return nil, 0, fmt.Errorf("list groups total: %w", err)
	}
	// interate through records
	rows, err := db.Query(s)
	if err != nil {
		return nil, 0, fmt.Errorf("list groups query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list groups rows: %w", rows.Err())
	}
	defer rows.Close()
	var grp sql.NullString
	for rows.Next() {
		if err = rows.Scan(&grp); err != nil {
			return nil, 0, fmt.Errorf("list groups rows scan: %w", err)
		}
		if _, err = grp.Value(); err != nil {
			continue
		}
		groups = append(groups, fmt.Sprintf("%v", grp.String))
	}
	return groups, total, nil
}

// MakeSlug takes a name and makes it into a URL friendly slug.
func MakeSlug(name string) string {
	n := database.TrimSP(name)
	n = remInitialism(n)
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

// Print list organizations or groups filtered by a name and summaries the results.
func Print(r Request) (total int, err error) {
	grp, total, err := list(r.Filter)
	if err != nil {
		return 0, fmt.Errorf("print list groups: %w", err)
	}
	logs.Println(total, "matching", r.Filter, "records found")
	a := make([]string, total)
	for i, g := range grp {
		if r.Progress {
			str.Progress(r.Filter, i+1, total)
		}
		s := g
		// initialism
		if r.Initialisms {
			if in, err := initialism(g); err != nil {
				return 0, fmt.Errorf("print initialism: %w", err)
			} else if in != "" {
				s = fmt.Sprintf("%v [%s]", s, in)
			}
		}
		// file totals
		if r.Counts {
			c, err := Count(g)
			if err != nil {
				return 0, fmt.Errorf("print counts: %w", err)
			}
			if c > 0 {
				s = fmt.Sprintf("%v (%d)", s, c)
			}
		}
		a[i] = s
	}
	// remove empty val
	if a[0] == "" {
		a = a[1:]
	}
	logs.Printf("\n%s\nTotal groups %d\n", strings.Join(a, ", "), total)
	return total, nil
}

// Variations creates format variations for a named group.
func Variations(name string) (vars []string, err error) {
	if name == "" {
		return vars, nil
	}
	name = strings.ToLower(name)
	vars = append(vars, name)
	s := strings.Split(name, " ")
	a := strings.Join(s, "")
	if name != a {
		vars = append(vars, a)
	}
	b := strings.Join(s, "-")
	if name != b {
		vars = append(vars, b)
	}
	c := strings.Join(s, "_")
	if name != c {
		vars = append(vars, c)
	}
	d := strings.Join(s, ".")
	if name != d {
		vars = append(vars, d)
	}
	if init, err := Initialism(name); err == nil && init != "" {
		vars = append(vars, strings.ToLower(init))
	} else if err != nil {
		return nil, fmt.Errorf("variations %q: %w", name, err)
	}
	return vars, nil
}

// Wheres are group categories.
func Wheres() []string {
	return []string{BBS.String(), FTP.String(), Group.String(), Magazine.String()}
}

// initialism returns a group's initialism or acronym.
// For example "Defacto2" would return "df2".
func initialism(name string) (string, error) {
	db := database.Connect()
	defer db.Close()
	var i string
	row := db.QueryRow("SELECT `initialisms` FROM groups WHERE `pubname` = ?", name)
	if err := row.Scan(&i); err != nil &&
		strings.Contains(err.Error(), "no rows in result set") {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("initialism %q: %w", name, err)
	}
	return i, db.Close()
}

// remInitialism removes a (bracketed initialism) from a string.
// For example "Defacto2 (DF2)" would return "Defacto2".
func remInitialism(s string) string {
	s = strings.TrimSpace(s)
	a := strings.Split(s, " ")
	l := a[len(a)-1]
	if l[:1] == "(" && l[len(l)-1:] == ")" {
		return strings.Join(a[:len(a)-1], " ")
	}
	return s
}

// groupsStmt returns a complete SQL WHERE statement where the groups are filtered.
func groupsStmt(f Filter, includeSoftDeletes bool) (stmt string, err error) {
	var inc, skip bool = includeSoftDeletes, false
	if f > -1 {
		skip = true
	}
	where, err := groupsWhere(f, inc)
	if err != nil {
		return "", fmt.Errorf("group statement %q: %w", f.String(), err)
	}
	switch skip {
	case true: // disable group_brand_by listings for BBS, FTP, group, magazine filters
		stmt = "SELECT DISTINCT group_brand_for AS pubCombined " +
			"FROM files WHERE Length(group_brand_for) <> 0 " + where
	default:
		stmt = "(SELECT DISTINCT group_brand_for AS pubCombined " +
			"FROM files WHERE Length(group_brand_for) <> 0 " + where + ")" +
			" UNION " +
			"(SELECT DISTINCT group_brand_by AS pubCombined " +
			"FROM files WHERE Length(group_brand_by) <> 0 " + where + ")"
	}
	return stmt + " ORDER BY pubCombined", nil
}

// groupsFilter returns a partial SQL WHERE statement to filter groups.
func groupsFilter(f Filter) (stmt string, err error) {
	switch f {
	case None:
		stmt = ""
	case Magazine:
		stmt = "section = 'magazine' AND"
	case BBS:
		stmt = "RIGHT(group_brand_for,4) = ' BBS' AND"
	case FTP:
		stmt = "RIGHT(group_brand_for,4) = ' FTP' AND"
	case Group: // only display groups who are listed under group_brand_for, group_brand_by only groups will be ignored
		stmt = "RIGHT(group_brand_for,4) != ' FTP' AND RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine' AND"
	default:
		return "", fmt.Errorf("group filter %q: %w", f.String(), ErrFilter)
	}
	return stmt, nil
}

// groupsWhere returns a partial SQL WHERE statement where groups are filtered.
func groupsWhere(f Filter, softDel bool) (stmt string, err error) {
	stmt, err = groupsFilter(f)
	if err != nil {
		return "", fmt.Errorf("groups where: %w", err)
	}
	switch {
	case stmt != "" && softDel:
		stmt = "AND " + stmt
	case stmt == "" && softDel: // do nothing
	case stmt != "" && !softDel:
		stmt = "AND " + stmt + " `deletedat` IS NULL"
	default:
		stmt = "AND `deletedat` IS NULL"
	}
	l := len(stmt)
	if l > 4 && stmt[l-4:] == " AND" {
		logs.Printf("%q|", stmt[l-4:])
		return stmt[:l-4], nil
	}
	return stmt, nil
}
