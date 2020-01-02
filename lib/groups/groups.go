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
)

const source = "/Users/ben/github/df2"

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
		if update := database.FileUpdate(path.Join(source, name), database.LastUpdate()); !update {
			println(name + " has nothing to update")
		} else {
			HTML(tags[i], true, false, name)
		}
	}
}

// CronThreads is a multithread Cronjob but cannot be used as this func is not thread-safe.
func CronThreads() {
	var (
		count = true
		pct   = false
	)
	// make these 4 image tasks multithread
	c := make(chan bool)
	go func() { HTML("bbs", count, pct, "bbs.htm"); c <- true }()
	go func() { HTML("ftp", count, pct, "ftp.htm"); c <- true }()
	go func() { HTML("group", count, pct, "group.htm"); c <- true }()
	go func() { HTML("magazine", count, pct, "magazine.htm"); c <- true }()
	<-c // sync 4 tasks
}

// HTML prints a snippet listing links to each group, with an optional file count.
func HTML(name string, count bool, countIndicator bool, filename string) {
	// <h2><a href="/g/13-omens">13 OMENS</a> 13O</h2><hr>
	tpl := `{{range .}}{{if .Hr}}<hr>{{end}}<h2><a href="/g/{{.ID}}">{{.Name}}</a>{{if .Initialism}} ({{.Initialism}}){{end}}{{if .Count}} <small>({{.Count}})</small>{{end}}</h2>{{end}}`
	type Group struct {
		ID         string
		Name       string
		Count      int
		Initialism string
		Hr         bool
	}
	grp, x := List(name)
	println(x, "matching", name, "records found")
	data := make([]Group, len(grp))
	cap := ""
	hr := false
	var cnt int
	total := len(grp)
	for i := range grp {
		if countIndicator {
			progressPct(name, i+1, total)
		}
		n := grp[i]
		switch c := n[:1]; {
		case cap == "":
			cap = c
		case c != cap:
			cap = c
			hr = true
		default:
			hr = false
		}
		switch count {
		case false:
			cnt = 0
		case true:
			cnt = Count(n)
		}
		data[i] = Group{
			ID:         MakeSlug(n),
			Name:       n,
			Count:      cnt,
			Initialism: initialism(n),
			Hr:         hr,
		}
	}
	t, err := template.New("h2").Parse(tpl)
	logs.Check(err)
	switch {
	case filename == "":
		err = t.Execute(os.Stdout, data)
		logs.Check(err)
	case name == "bbs", name == "ftp", name == "group", name == "magazine":
		f, err := os.Create(path.Join(source, filename))
		logs.Check(err)
		defer f.Close()
		err = t.Execute(f, data)
		logs.Check(err)
	default:
		logs.Check(fmt.Errorf("invalid name %q used", name))
	}
}

// List organizations or groups filtered by a name.
func List(name string) ([]string, int) {
	db := database.Connect()
	defer db.Close()
	s := sqlGroups(name, false)
	total := database.Total(s)
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

func removeInitialism(s string) string {
	s = strings.TrimSpace(s)
	a := strings.Split(s, " ")
	l := a[len(a)-1]
	if l[:1] == "(" && l[len(l)-1:] == ")" {
		return strings.Join(a[:len(a)-1], " ")
	}
	return s
}

// MakeSlug takes a name and makes it into a URL friendly slug.
func MakeSlug(name string) string {
	n := removeInitialism(name)
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
func Print(name string, count bool) {
	g, i := List(name)
	fmt.Println(strings.Join(g, ", "))
	fmt.Println("Total groups", i)
}

// progressPct returns the count of total remaining as a percentage.
func progressPct(name string, count int, total int) {
	fmt.Printf("\rQuerying %s %.2f %%", name, float64(count)/float64(total)*100)
}

// progressSum returns the count of total remaining.
// func progressSum(count int, total int) {
// 	fmt.Printf("\rBuilding %d/%d", count, total)
// }

func wheres() [4]string {
	return [4]string{"bbs", "ftp", "group", "magazine"}
}

func sqlInitialisms(name string, includeSoftDeletes bool) string {
	inc := includeSoftDeletes
	var sql string
	sql = "SELECT pubValue, (SELECT CONCAT(pubCombined, ' ', '(', initialisms, ')') "
	sql += "FROM groups WHERE pubName = pubCombined AND Length(initialisms) <> 0) AS pubCombined"
	sql += " FROM (SELECT TRIM(group_brand_for) AS pubValue, group_brand_for AS pubCombined " + sqlGroupsWhere(name, inc)
	sql += "FROM files WHERE Length(group_brand_for) <> 0"
	sql += " UNION SELECT TRIM(group_brand_by) AS pubValue, group_brand_by AS pubCombined " + sqlGroupsWhere(name, inc)
	sql += "FROM files WHERE Length(group_brand_by) <> 0) AS pubTbl"
	return sql + " ORDER BY pubCombined"
}

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

// sqlGroups returns a complete SQL WHERE statement where the groups are filtered by name.
func sqlGroups(name string, includeSoftDeletes bool) string {
	inc := includeSoftDeletes
	skip := false
	for _, a := range wheres() {
		if a == name {
			skip = true
		}
	}
	var sql string
	switch skip {
	case true: // disable group_brand_by listings for BBS, FTP, group, magazine filters
		sql = "SELECT DISTINCT group_brand_for AS pubCombined "
		sql += "FROM files WHERE Length(group_brand_for) <> 0 " + sqlGroupsWhere(name, inc)
	default:
		sql = "(SELECT DISTINCT group_brand_for AS pubCombined "
		sql += "FROM files WHERE Length(group_brand_for) <> 0 " + sqlGroupsWhere(name, inc) + ")"
		sql += " UNION "
		sql += "(SELECT DISTINCT group_brand_by AS pubCombined "
		sql += "FROM files WHERE Length(group_brand_by) <> 0 " + sqlGroupsWhere(name, inc) + ")"
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
		fmt.Printf("%q|", sql[l-4:])
		return sql[:l-4]
	}
	return sql
}
