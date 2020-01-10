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
	"gopkg.in/gookit/color.v1"
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
func Count(name string) int {
	db := database.Connect()
	defer db.Close()
	n := name
	var count int
	s := "SELECT COUNT(*) FROM files WHERE "
	s += fmt.Sprintf("group_brand_for='%v' OR group_brand_for LIKE '%v,%%' OR group_brand_for LIKE '%%, %v,%%' OR group_brand_for LIKE '%%, %v'", n, n, n, n)
	s += fmt.Sprintf(" OR group_brand_by='%v' OR group_brand_by LIKE '%v,%%' OR group_brand_by LIKE '%%, %v,%%' OR group_brand_by LIKE '%%, %v'", n, n, n, n)
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
			println(name + " has nothing to update")
		} else {
			Request{tags[i], true, true, false}.HTML(name)
		}
	}
}

// HTML prints a snippet listing links to each group, with an optional file count.
func (r Request) HTML(filename string) {
	// <h2><a href="/g/13-omens">13 OMENS</a> 13O</h2><hr>
	tpl := `{{range .}}{{if .Hr}}<hr>{{end}}<h2><a href="/g/{{.ID}}">{{.Name}}</a>{{if .Initialism}} ({{.Initialism}}){{end}}{{if .Count}} <small>({{.Count}})</small>{{end}}</h2>{{end}}`
	grp, x := List(r.Filter)
	f := r.Filter
	if f == "" {
		f = "all"
	}
	println(x, "matching", f, "records found")
	data := make([]Group, len(grp))
	cap := ""
	hr := false
	total := len(grp)
	for i := range grp {
		if r.Progress {
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
		logs.Check(fmt.Errorf("invalid filter %q used", r.Filter))
	}
}

// Fix the formatting of group names.
func Fix(simulate bool) {
	grp, _ := List("")
	c := 0
	for i := range grp {
		c = c + fixApply(simulate, grp[i])
	}
	switch {
	case c > 0 && simulate:
		println(c, "fixes required")
		color.Notice.Println("use the --simulate=false flag to apply these fixes")
	case c > 0:
		println(c, "fixes applied")
	default:
		println("no fixes applied")
	}
}

// FixSpaces removes duplicate spaces from a string.
func FixSpaces(s string) string {
	r := regexp.MustCompile(`\s+`)
	return r.ReplaceAllString(s, " ")
}

// List organizations or groups filtered by a name.
func List(name string) ([]string, int) {
	db := database.Connect()
	defer db.Close()
	s := sqlGroups(name, false)
	total := database.Total(&s)
	// interate through records
	rows, err := db.Query(s)
	logs.Check(err)
	var grp sql.NullString
	i := 0
	g := ""
	grps := []string{}
	for rows.Next() {
		err = rows.Scan(&grp)
		logs.Check(err)
		if _, err = grp.Value(); err != nil {
			continue
		}
		i++
		g = fmt.Sprintf("%v", grp.String)
		grps = append(grps, g)
	}
	return grps, total
}

// MakeSlug takes a name and makes it into a URL friendly slug.
func MakeSlug(name string) string {
	n := FixSpaces(name)
	n = removeInitialism(n)
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
	grp, total := List(r.Filter)
	println(total, "matching", r.Filter, "records found")
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

// Wheres are group categories.
func Wheres() []string {
	return strings.Split(Filters, ",")
}

func fixes(g string) string {
	f := FixSpaces(g)
	f = strings.TrimSpace(f)
	return fixThe(f)
}

func fixApply(simulate bool, g string) int {
	f := fixes(g)
	v := 0
	if f != g && simulate {
		logs.Printf("%s %q %s %s\n", color.Question.Sprint("?"), g, color.Question.Sprint("!="), color.Info.Sprint(f))
		v++
	} else if f != g {
		s := logs.Y()
		v++
		if x := fixGroup(g); !x {
			s = logs.X()
			v--
		}
		logs.Printf("%s %q %s %s\n", s, g, color.Question.Sprint("⟫"), color.Info.Sprint(f))
	}
	return v
}

func fixThe(g string) string {
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

func fixGroup(g string) bool {
	db := database.Connect()
	defer db.Close()
	var sql = [2]string{
		"UPDATE files SET group_brand_for=? WHERE group_brand_for=?",
		"UPDATE files SET group_brand_by=? WHERE group_brand_by=?",
	}
	for i := range sql {
		update, err := db.Prepare(sql[i])
		logs.Check(err)
		if err != nil {
			return false
		}
		_, err = update.Exec(fixes(g), g)
		logs.Check(err)
		if err != nil {
			return false
		}
	}

	return true
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

// removeInitialism removes a (bracketed initialism) from a string.
// For example "Defacto2 (DF2)" would return "Defacto2".
func removeInitialism(s string) string {
	s = strings.TrimSpace(s)
	a := strings.Split(s, " ")
	l := a[len(a)-1]
	if l[:1] == "(" && l[len(l)-1:] == ")" {
		return strings.Join(a[:len(a)-1], " ")
	}
	return s
}

// sqlGroups returns a complete SQL WHERE statement where the groups are filtered by name.
func sqlGroups(name string, includeSoftDeletes bool) string {
	var inc, skip bool = includeSoftDeletes, false
	for _, a := range Wheres() {
		if a == name {
			skip = true
		}
	}
	var sql string
	var where = sqlGroupsWhere(name, inc)
	switch skip {
	case true: // disable group_brand_by listings for BBS, FTP, group, magazine filters
		sql = "SELECT DISTINCT group_brand_for AS pubCombined "
		sql += "FROM files WHERE Length(group_brand_for) <> 0 " + where
	default:
		sql = "(SELECT DISTINCT group_brand_for AS pubCombined "
		sql += "FROM files WHERE Length(group_brand_for) <> 0 " + where + ")"
		sql += " UNION "
		sql += "(SELECT DISTINCT group_brand_by AS pubCombined "
		sql += "FROM files WHERE Length(group_brand_by) <> 0 " + where + ")"
	}
	return sql + " ORDER BY pubCombined"
}

// sqlGroupsFilter returns a partial SQL WHERE statement to filer groups by name.
func sqlGroupsFilter(name string) string {
	var sql string
	switch name {
	case "magazine":
		sql = "section = 'magazine' AND"
	case "bbs":
		sql = "RIGHT(group_brand_for,4) = ' BBS' AND"
	case "ftp":
		sql = "RIGHT(group_brand_for,4) = ' FTP' AND"
	case "group": // only display groups who are listed under group_brand_for, group_brand_by only groups will be ignored
		sql = "RIGHT(group_brand_for,4) != ' FTP' AND RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine' AND"
	}
	return sql
}

// sqlGroupsWhere returns a partial SQL WHERE statement where groups are filtered by name.
func sqlGroupsWhere(name string, includeSoftDeletes bool) string {
	sql := sqlGroupsFilter(name)
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
		return sql[:l-4]
	}
	return sql
}
