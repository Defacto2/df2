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
func List(role string) (people []string, total int, err error) {
	db := database.Connect()
	defer db.Close()
	s := peopleStmt(role, false)
	if s == "" {
		return nil, 0, nil
	}
	if total, err = database.Total(&s); err != nil {
		return nil, 0, err
	}
	// interate through records
	rows, err := db.Query(s)
	if err != nil {
		return nil, 0, err
	} else if rows.Err() != nil {
		return nil, 0, rows.Err()
	}
	var persons sql.NullString
	i := 0
	for rows.Next() {
		if err := rows.Scan(&persons); err != nil {
			return nil, 0, err
		}
		if _, err = persons.Value(); err != nil {
			continue
		}
		i++
		people = append(people, fmt.Sprintf("%v", persons.String))
	}
	return people, total, nil
}

// DataList prints an auto-complete list for HTML input elements.
func DataList(filename string, r Request) error {
	// <option value="$YN (Syndicate BBS)" label="$YN (Syndicate BBS)">
	tpl := `{{range .}}<option value="{{.Nick}}" label="{{.Nick}}">{{end}}`
	if err := parse(filename, tpl, r); err != nil {
		return err
	}
	return nil
}

// HTML prints a snippet listing links to each person.
func HTML(filename string, r Request) error {
	// <h2><a href="/p/ben">Ben</a></h2><hr>
	tpl := `{{range .}}{{if .Hr}}<hr>{{end}}<h2><a href="/p/{{.ID}}">{{.Nick}}</a></h2>{{end}}`
	if err := parse(filename, tpl, r); err != nil {
		return err
	}
	return nil
}

func parse(filename string, tpl string, r Request) error {
	grp, x, err := List(r.Filter)
	if err != nil {
		return err
	}
	f := r.Filter
	if f == "" {
		f = "all"
	}
	logs.Println(x, "matching", f, "records found")
	data, cap, hr := make([]Person, len(grp)), "", false
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
	if err != nil {
		return err
	}
	switch {
	case filename == "":
		if err = t.Execute(os.Stdout, data); err != nil {
			return err
		}
	case r.Filter == "artists", r.Filter == "coders", r.Filter == "musicians", r.Filter == "writers":
		f, err := os.Create(path.Join(viper.GetString("directory.html"), filename))
		if err != nil {
			return err
		}
		defer f.Close()
		if err = t.Execute(f, data); err != nil {
			return err
		}
	default:
		return fmt.Errorf("people html parse: invalid filter %q used", r.Filter)
	}
	return nil
}

// Print lists people filtered by a role and summaries the results.
func Print(r Request) error {
	ppl, total, err := List(r.Filter)
	if err != nil {
		return err
	}
	logs.Println(total, "matching", r.Filter, "records found")
	var a = make([]string, total)
	for i := range ppl {
		if r.Progress {
			logs.ProgressPct(r.Filter, i+1, total)
		}
		// role
		x := strings.Split(ppl[i], ",")
		a = append(a, x...)
	}
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
	fmt.Printf("\n\n%s\nTotal authors %d", strings.Join(a, ","), len(a))
	return nil
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

// peopleStmt returns a complete SQL WHERE statement where the people are filtered by a role.
func peopleStmt(role string, softDel bool) (stmt string) {
	var del = func() string {
		if softDel {
			return ""
		}
		return "AND `deletedat` IS NULL"
	}
	var d = func(s string) string {
		return fmt.Sprintf(" UNION (SELECT DISTINCT %s AS pubCombined FROM files WHERE Length(%s) <> 0 %s)",
			s, s, del())
	}
	f := roles(role)
	if strings.ContainsAny(f, "w") {
		stmt += d("credit_text")
	}
	if strings.ContainsAny(f, "m") {
		stmt += d("credit_audio")
	}
	if strings.ContainsAny(f, "c") {
		stmt += d("credit_program")
	}
	if strings.ContainsAny(f, "a") {
		stmt += d("credit_illustration")
	}
	if stmt == "" {
		return stmt
	}
	stmt += " ORDER BY pubCombined"
	stmt = strings.Replace(stmt, "UNION (SELECT DISTINCT ", "(SELECT DISTINCT ", 1)
	return stmt
}
