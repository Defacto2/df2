package people

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/campoy/unique"
	"github.com/spf13/viper"
)

// Request flags for people functions.
type Request struct {
	// Filter people by category.
	Filter string
	// Counts the person's total files.
	Counts bool
	// Progress counter when requesting database data.
	Progress bool
}

// Person data.
type Person struct {
	// ID used in URLs to link to the person.
	ID string
	// Nick of the person.
	Nick string
	// Inject a HR element to separate a collection of groups.
	Hr bool
}

// Filters are peoples' roles.
const Filters = "artists,coders,musicians,writers"

// List people filtered by a role.
func List(role string) ([]string, int) {
	db := database.Connect()
	defer db.Close()
	s := sqlPeople(role, false)
	total := database.Total(&s)
	// interate through records
	rows, err := db.Query(s)
	logs.Check(err)
	var persons sql.NullString
	i := 0
	g := ""
	ppl := []string{}
	for rows.Next() {
		err = rows.Scan(&persons)
		logs.Check(err)
		if _, err = persons.Value(); err != nil {
			continue
		}
		i++
		g = fmt.Sprintf("%v", persons.String)
		ppl = append(ppl, g)
	}
	return ppl, total
}

// DataList prints an auto-complete list for HTML input elements.
func DataList(filename string, r Request) {
	// <option value="$YN (Syndicate BBS)" label="$YN (Syndicate BBS)">
	tpl := `{{range .}}<option value="{{.Nick}}" label="{{.Nick}}">{{end}}`
	parse(filename, tpl, r)
}

// HTML prints a snippet listing links to each person.
func HTML(filename string, r Request) {
	// <h2><a href="/p/ben">Ben</a></h2><hr>
	tpl := `{{range .}}{{if .Hr}}<hr>{{end}}<h2><a href="/p/{{.ID}}">{{.Nick}}</a></h2>{{end}}`
	parse(filename, tpl, r)
}

func parse(filename string, tpl string, r Request) {
	grp, x := List(r.Filter)
	f := r.Filter
	if f == "" {
		f = "all"
	}
	println(x, "matching", f, "records found")
	data := make([]Person, len(grp))
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
		data[i] = Person{
			ID:   groups.MakeSlug(n),
			Nick: n,
			Hr:   hr,
		}
	}
	t, err := template.New("h2").Parse(tpl)
	logs.Check(err)
	switch {
	case filename == "":
		err = t.Execute(os.Stdout, data)
		logs.Check(err)
	case r.Filter == "artists", r.Filter == "coders", r.Filter == "musicians", r.Filter == "writers":
		f, err := os.Create(path.Join(viper.GetString("directory.html"), filename))
		logs.Check(err)
		defer f.Close()
		err = t.Execute(f, data)
		logs.Check(err)
	default:
		logs.Check(fmt.Errorf("invalid filter %q used", r.Filter))
	}
}

// Print lists people filtered by a role and summaries the results.
func Print(r Request) {
	ppl, total := List(r.Filter)
	println(total, "matching", r.Filter, "records found")
	var a []string
	for i := range ppl {
		if r.Progress {
			logs.ProgressPct(r.Filter, i+1, total)
		}
		// role
		x := strings.Split(ppl[i], ",")
		a = append(a, x...)
	}
	//ppl = nil
	// title and sort names
	for i := range a {
		if r.Progress {
			logs.ProgressPct(r.Filter, i+1, total)
		}
		a[i] = strings.Title(a[i])
	}
	sort.Strings(a)
	// remove duplicates
	less := func(i, j int) bool { return a[i] < a[j] }
	unique.Slice(&a, less)
	logs.Println()
	logs.Println(strings.Join(a, ","))
	logs.Println("Total authors", len(a))
}

// Wheres are group categories.
func Wheres() []string {
	return strings.Split(Filters, ",")
}

func roles(r string) string {
	switch r {
	case "writers", "w":
		return "w"
	case "musicians", "m":
		return "m"
	case "coders", "c":
		return "c"
	case "artists", "a":
		return "a"
	case "", "all":
		return "wmca"
	}
	return ""
}

// sqlPeople returns a complete SQL WHERE statement where the people are filtered by a role.
func sqlPeople(role string, includeSoftDeletes bool) string {
	inc := includeSoftDeletes
	f := roles(role)
	var sql string
	if strings.ContainsAny(f, "w") {
		sql += " UNION (SELECT DISTINCT credit_text AS pubCombined FROM files WHERE Length(credit_text) <> 0 " + sqlPeopleDel(inc) + ")"
	}
	if strings.ContainsAny(f, "m") {
		sql += " UNION (SELECT DISTINCT credit_audio AS pubCombined FROM files WHERE Length(credit_audio) <> 0 " + sqlPeopleDel(inc) + ")"
	}
	if strings.ContainsAny(f, "c") {
		sql += " UNION (SELECT DISTINCT credit_program AS pubCombined FROM files WHERE Length(credit_program) <> 0 " + sqlPeopleDel(inc) + ")"
	}
	if strings.ContainsAny(f, "a") {
		sql += " UNION (SELECT DISTINCT credit_illustration AS pubCombined FROM files WHERE Length(credit_illustration) <> 0 " + sqlPeopleDel(inc) + ")"
	}
	sql += " ORDER BY pubCombined"
	sql = strings.Replace(sql, "UNION (SELECT DISTINCT ", "(SELECT DISTINCT ", 1)
	return sql
}

// sqlPeopleDel returns a partial SQL WHERE to handle soft deleted entries.
func sqlPeopleDel(includeSoftDeletes bool) string {
	sql := ""
	if includeSoftDeletes {
		return sql
	}
	sql += "AND `deletedat` IS NULL"
	return sql
}
