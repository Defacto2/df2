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
func Count(name string) (count int) {
	db := database.Connect()
	defer db.Close()
	n := name
	s := "SELECT COUNT(*) FROM files WHERE " +
		fmt.Sprintf("group_brand_for='%v' OR group_brand_for LIKE '%v,%%' OR group_brand_for LIKE '%%, %v,%%' OR group_brand_for LIKE '%%, %v'", n, n, n, n) +
		fmt.Sprintf(" OR group_brand_by='%v' OR group_brand_by LIKE '%v,%%' OR group_brand_by LIKE '%%, %v,%%' OR group_brand_by LIKE '%%, %v'", n, n, n, n)
	row := db.QueryRow(s)
	err := row.Scan(&count)
	logs.Check(err)
	return count
}

// Cronjob is used for system automation to generate dynamic HTML pages.
func Cronjob() {
	tags := []string{"bbs", "ftp", "group", "magazine"}
	for i := range tags {
		name := tags[i] + ".htm"
		if update := database.FileUpdate(path.Join(viper.GetString("directory.html"), name), database.LastUpdate()); !update {
			logs.Println(name + " has nothing to update")
		} else {
			Request{tags[i], true, true, false}.HTML(name)
		}
	}
}

// DataList prints an auto-complete list for HTML input elements.
func (r Request) DataList(filename string) {
	// <option value="Bitchin ANSI Design" label="BAD (Bitchin ANSI Design)">
	tpl := `{{range .}}{{if .Initialism}}<option value="{{.Name}}" label="{{.Initialism}} ({{.Name}})">{{end}}<option value="{{.Name}}" label="{{.Name}}">{{end}}`
	r.parse(filename, tpl)
}

// HTML prints a snippet listing links to each group, with an optional file count.
func (r Request) HTML(filename string) {
	// <h2><a href="/g/13-omens">13 OMENS</a> 13O</h2><hr>
	tpl := `{{range .}}{{if .Hr}}<hr>{{end}}<h2><a href="/g/{{.ID}}">{{.Name}}</a>{{if .Initialism}} ({{.Initialism}}){{end}}{{if .Count}} <small>({{.Count}})</small>{{end}}</h2>{{end}}`
	r.parse(filename, tpl)
}

func (r Request) parse(filename string, tpl string) {
	grp, x := list(r.Filter)
	f := r.Filter
	if f == "" {
		f = "all"
	}
	logs.Println(x, "matching", f, "records found")
	data := make([]Group, len(grp))
	cap := ""
	hr := false
	total := len(grp)
	for i := range grp {
		if !logs.Quiet && r.Progress {
			logs.ProgressPct(r.Filter, i+1, total)
		}
		n := grp[i]
		// hr element
		switch c := n[:1]; {
		case cap == "":
			cap = c
		case c != cap:
			cap = c
			hr = true
		default:
			hr = false
		}
		// file totals
		c := 0
		if r.Counts {
			c = Count(n)
		}
		// initialism
		in := ""
		if r.Initialisms {
			in = initialism(n)
		}
		data[i] = Group{
			ID:         MakeSlug(n),
			Name:       n,
			Count:      c,
			Initialism: in,
			Hr:         hr,
		}
	}
	t, err := template.New("h2").Parse(tpl)
	logs.Check(err)
	switch {
	case filename == "":
		err = t.Execute(os.Stdout, data)
		logs.Check(err)
	case r.Filter == "bbs", r.Filter == "ftp", r.Filter == "group", r.Filter == "magazine":
		f, err := os.Create(path.Join(viper.GetString("directory.html"), filename))
		logs.Check(err)
		defer f.Close()
		err = t.Execute(f, data)
		logs.Check(err)
	default:
		logs.Check(fmt.Errorf("groups parse: invalid filter %q used", r.Filter))
	}
}

// list all organizations or filtered groups.
func list(filter string) (groups []string, total int) {
	db := database.Connect()
	defer db.Close()
	s, err := sqlGroups(filter, false)
	logs.Check(err)
	fmt.Printf("%s\n\n", s)
	total = database.Total(&s)
	// interate through records
	rows, err := db.Query(s)
	logs.Check(err)
	var grp sql.NullString
	for rows.Next() {
		err = rows.Scan(&grp)
		logs.Check(err)
		if _, err = grp.Value(); err != nil {
			continue
		}
		groups = append(groups, fmt.Sprintf("%v", grp.String))
	}
	return groups, total
}

// MakeSlug takes a name and makes it into a URL friendly slug.
func MakeSlug(name string) string {
	n := remDupeSpaces(name)
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
func Print(r Request) {
	grp, total := list(r.Filter)
	logs.Println(total, "matching", r.Filter, "records found")
	var a []string
	for i := range grp {
		if r.Progress {
			logs.ProgressPct(r.Filter, i+1, total)
		}
		// name
		n := grp[i]
		s := n
		// initialism
		if r.Initialisms {
			if in := initialism(n); in != "" {
				s = fmt.Sprintf("%v [%s]", s, in)
			}
		}
		// file totals
		if r.Counts {
			if c := Count(n); c > 0 {
				s = fmt.Sprintf("%v (%d)", s, c)
			}
		}
		a = append(a, s)
	}
	logs.Println()
	logs.Println(strings.Join(a, ", "))
	logs.Println("Total groups", total)
}

// Variations creates format variations for a named group.
func Variations(name string) (vars []string) {
	if name == "" {
		return vars
	}
	name = strings.ToLower(name)
	vars = append(vars, name)
	if name != "" {
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
		}
	}
	return vars
}

// Wheres are group categories.
func Wheres() []string {
	return strings.Split(Filters, ",")
}

// initialism returns a group's initialism or acronym.
// For example "Defacto2" would return "df2"
func initialism(name string) string {
	db := database.Connect()
	defer db.Close()
	var i string
	s := fmt.Sprintf("SELECT `initialisms` FROM groups WHERE `pubname` = %q", name)
	row := db.QueryRow(s)
	err := row.Scan(&i)
	if err != nil && strings.Contains(err.Error(), "no rows in result set") {
		return ""
	}
	logs.Check(err)
	return i
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

// sqlGroups returns a complete SQL WHERE statement where the groups are filtered.
func sqlGroups(filter string, includeSoftDeletes bool) (sql string, err error) {
	var inc, skip bool = includeSoftDeletes, false
	for _, a := range Wheres() {
		if a == filter {
			skip = true
		}
	}
	where, err := sqlGroupsWhere(filter, inc)
	if err != nil {
		return sql, err
	}
	switch skip {
	case true: // disable group_brand_by listings for BBS, FTP, group, magazine filters
		sql = "SELECT DISTINCT group_brand_for AS pubCombined " +
			"FROM files WHERE Length(group_brand_for) <> 0 " + where
	default:
		sql = "(SELECT DISTINCT group_brand_for AS pubCombined " +
			"FROM files WHERE Length(group_brand_for) <> 0 " + where + ")" +
			" UNION " +
			"(SELECT DISTINCT group_brand_by AS pubCombined " +
			"FROM files WHERE Length(group_brand_by) <> 0 " + where + ")"
	}
	return sql + " ORDER BY pubCombined", err
}

// sqlGroupsFilter returns a partial SQL WHERE statement to filter groups.
func sqlGroupsFilter(filter string) (sql string, err error) {
	switch filter {
	case "":
		return sql, err
	case "magazine":
		sql = "section = 'magazine' AND"
	case "bbs":
		sql = "RIGHT(group_brand_for,4) = ' BBS' AND"
	case "ftp":
		sql = "RIGHT(group_brand_for,4) = ' FTP' AND"
	case "group": // only display groups who are listed under group_brand_for, group_brand_by only groups will be ignored
		sql = "RIGHT(group_brand_for,4) != ' FTP' AND RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine' AND"
	default:
		err = fmt.Errorf("groups sqlGroupsFilter: unsupported filter option %q, leave blank or use either %v", filter, Filters)
	}
	return sql, err
}

// sqlGroupsWhere returns a partial SQL WHERE statement where groups are filtered.
func sqlGroupsWhere(filter string, includeSoftDeletes bool) (sql string, err error) {
	sql, err = sqlGroupsFilter(filter)
	if err != nil {
		return sql, err
	}
	switch {
	case sql != "" && includeSoftDeletes:
		sql = "AND " + sql
	case sql == "" && includeSoftDeletes: // do nothing
	case sql != "" && !includeSoftDeletes:
		sql = "AND " + sql + " `deletedat` IS NULL"
	default:
		sql = "AND `deletedat` IS NULL"
	}
	l := len(sql)
	if l > 4 && sql[l-4:] == " AND" {
		logs.Printf("%q|", sql[l-4:])
		return sql[:l-4], err
	}
	return sql, err
}
