package groups

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/viper"
)

const htm = ".htm"

// Filters are group categories.
const Filters = "bbs,ftp,group,magazine"

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

// Group data.
type Group struct {
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
	s := "SELECT COUNT(*) FROM files WHERE " +
		fmt.Sprintf("group_brand_for='%v' OR group_brand_for LIKE '%v,%%' OR group_brand_for LIKE '%%, %v,%%' OR group_brand_for LIKE '%%, %v'", n, n, n, n) +
		fmt.Sprintf(" OR group_brand_by='%v' OR group_brand_by LIKE '%v,%%' OR group_brand_by LIKE '%%, %v,%%' OR group_brand_by LIKE '%%, %v'", n, n, n, n)
	row := db.QueryRow(s)
	if err = row.Scan(&count); err != nil {
		return 0, err
	}
	return count, db.Close()
}

// Cronjob is used for system automation to generate dynamic HTML pages.
func Cronjob() error {
	tags := []string{"bbs", "ftp", "group", "magazine"}
	for i := range tags {
		last, err := database.LastUpdate()
		if err != nil {
			return err
		}
		f := tags[i] + htm
		n := path.Join(viper.GetString("directory.html"), f)
		if update, err := database.FileUpdate(n, last); err != nil {
			return err
		} else if !update {
			logs.Println(f + " has nothing to update")
		} else {
			r := Request{tags[i], true, true, false}
			if err := r.HTML(f); err != nil {
				return err
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
		return err
	}
	return nil
}

// HTML prints a snippet listing links to each group, with an optional file count.
func (r Request) HTML(filename string) error {
	// <h2><a href="/g/13-omens">13 OMENS</a> 13O</h2><hr>
	tpl := `{{range .}}{{if .Hr}}<hr>{{end}}<h2><a href="/g/{{.ID}}">{{.Name}}</a>{{if .Initialism}} ({{.Initialism}}){{end}}{{if .Count}} <small>({{.Count}})</small>{{end}}</h2>{{end}}`
	if err := r.parse(filename, tpl); err != nil {
		return err
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
		return Count(group)
	}
	return 0, nil
}

func (r Request) initialism(group string) (name string, err error) {
	if r.Initialisms {
		return initialism(group)
	}
	return "", nil
}

func (r Request) iterate(groups []string) (g *[]Group, err error) {
	total := len(groups)
	data := make([]Group, total)
	lastLetter, hr := "", false
	for i, grp := range groups {
		if !logs.Quiet && r.Progress {
			logs.ProgressPct(r.Filter, i+1, total)
		}
		lastLetter, hr = hrElement(lastLetter, grp)
		c, err := r.files(grp)
		if err != nil {
			return nil, err
		}
		init, err := r.initialism(grp)
		if err != nil {
			return nil, err
		}
		data[i] = Group{
			ID:         MakeSlug(grp),
			Name:       grp,
			Count:      c,
			Initialism: init,
			Hr:         hr,
		}
	}
	return &data, nil
}

func (r Request) parse(filename string, templ string) (err error) {
	groups, total, err := list(r.Filter)
	if err != nil {
		return err
	}
	if f := r.Filter; f == "" {
		logs.Println(total, "matching (all) records found")
	} else {
		logs.Println(total, "matching", f, "records found")
	}
	data, err := r.iterate(groups)
	if err != nil {
		return err
	}
	t, err := template.New("h2").Parse(templ)
	if err != nil {
		return err
	}
	switch {
	case filename == "":
		if err = t.Execute(os.Stdout, &data); err != nil {
			return err
		}
	case r.Filter == "bbs", r.Filter == "ftp", r.Filter == "group", r.Filter == "magazine":
		f, err := os.Create(path.Join(viper.GetString("directory.html"), filename))
		if err != nil {
			return err
		}
		defer f.Close()
		if err = t.Execute(f, &data); err != nil {
			return err
		}
	default:
		return fmt.Errorf("groups parse: invalid filter %q used", r.Filter)
	}
	return nil
}

// list all organizations or filtered groups.
func list(filter string) (groups []string, total int, err error) {
	db := database.Connect()
	defer db.Close()
	s, err := groupsStmt(filter, false)
	if err != nil {
		return nil, 0, err
	}
	total, err = database.Total(&s)
	if err != nil {
		return nil, 0, err
	}
	// interate through records
	rows, err := db.Query(s)
	if err != nil {
		return nil, 0, err
	} else if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var grp sql.NullString
	for rows.Next() {
		if err = rows.Scan(&grp); err != nil {
			return nil, 0, err
		}
		if _, err = grp.Value(); err != nil {
			continue
		}
		groups = append(groups, fmt.Sprintf("%v", grp.String))
	}
	return groups, total, err
}

// MakeSlug takes a name and makes it into a URL friendly slug.
func MakeSlug(name string) string {
	n := trimSP(name)
	n = remInitialism(n)
	n = strings.ReplaceAll(n, "-", "_")
	n = strings.ReplaceAll(n, ", ", "*")
	n = strings.ReplaceAll(n, " & ", " ampersand ")
	re := regexp.MustCompile(` ([0-9])`)
	n = re.ReplaceAllString(n, `-$1`)
	re = regexp.MustCompile(`[^A-Za-z0-9 \-\+\.\_\*]`) // remove all chars except these
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
		return 0, err
	}
	logs.Println(total, "matching", r.Filter, "records found")
	a := make([]string, total)
	for i := range grp {
		if r.Progress {
			logs.ProgressPct(r.Filter, i+1, total)
		}
		// name
		n := grp[i]
		s := n
		// initialism
		if r.Initialisms {
			if in, err := initialism(n); err != nil {
				return 0, err
			} else if in != "" {
				s = fmt.Sprintf("%v [%s]", s, in)
			}
		}
		// file totals
		if r.Counts {
			c, err := Count(n)
			if err != nil {
				return 0, err
			}
			if c > 0 {
				s = fmt.Sprintf("%v (%d)", s, c)
			}
		}
		a = append(a, s)
	}
	logs.Printf("\n%s\nTotal groups %d", strings.Join(a, ", "), total)
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
		return nil, err
	}
	return vars, nil
}

// Wheres are group categories.
func Wheres() []string {
	return strings.Split(Filters, ",")
}

// initialism returns a group's initialism or acronym.
// For example "Defacto2" would return "df2".
func initialism(name string) (string, error) {
	db := database.Connect()
	defer db.Close()
	var i string
	s := fmt.Sprintf("SELECT `initialisms` FROM groups WHERE `pubname` = %q", name)
	row := db.QueryRow(s)
	if err := row.Scan(&i); err != nil &&
		strings.Contains(err.Error(), "no rows in result set") {
		return "", nil
	} else if err != nil {
		return "", err
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
func groupsStmt(filter string, includeSoftDeletes bool) (stmt string, err error) {
	var inc, skip bool = includeSoftDeletes, false
	for _, a := range Wheres() {
		if a == filter {
			skip = true
		}
	}
	where, err := groupsWhere(filter, inc)
	if err != nil {
		return "", err
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
func groupsFilter(filter string) (stmt string, err error) {
	switch filter {
	case "":
		return "", nil
	case "magazine":
		stmt = "section = 'magazine' AND"
	case "bbs":
		stmt = "RIGHT(group_brand_for,4) = ' BBS' AND"
	case "ftp":
		stmt = "RIGHT(group_brand_for,4) = ' FTP' AND"
	case "group": // only display groups who are listed under group_brand_for, group_brand_by only groups will be ignored
		stmt = "RIGHT(group_brand_for,4) != ' FTP' AND RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine' AND"
	default:
		err = fmt.Errorf("groups groupsFilter: unsupported filter option %q, leave blank or use either %v", filter, Filters)
		return "", err
	}
	return stmt, nil
}

// groupsWhere returns a partial SQL WHERE statement where groups are filtered.
func groupsWhere(filter string, softDel bool) (stmt string, err error) {
	stmt, err = groupsFilter(filter)
	if err != nil {
		return "", err
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
