package people

import (
	"bufio"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/campoy/unique"
	"github.com/spf13/viper"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
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

type Role int

const (
	Everyone Role = iota
	Artists
	Coders
	Musicians
	Writers
)

const all = "all"

func (r Role) String() string {
	switch r {
	case Everyone:
		return all
	case Artists:
		return "artists"
	case Coders:
		return "coders"
	case Musicians:
		return "musicians"
	case Writers:
		return "writers"
	default:
		return ""
	}
}

var (
	ErrFilter = errors.New("invalid filter used")
	ErrRole   = errors.New("unknown role")
)

// DataList prints an auto-complete list for HTML input elements.
func DataList(filename string, r Request) error {
	// <option value="$YN (Syndicate BBS)" label="$YN (Syndicate BBS)">
	tpl := `{{range .}}<option value="{{.Nick}}" label="{{.Nick}}">{{end}}`
	if err := parse(filename, tpl, r); err != nil {
		return fmt.Errorf("datalist: %w", err)
	}
	return nil
}

// Filters is a Role slice for use with the Cobra filterFlag.
func Filters() []string {
	return []string{Artists.String(), Coders.String(), Musicians.String(), Writers.String()}
}

// HTML prints a snippet listing links to each person.
func HTML(filename string, r Request) error {
	// <h2><a href="/p/ben">Ben</a></h2><hr>
	tpl := `{{range .}}{{if .Hr}}<hr>{{end}}<h2><a href="/p/{{.ID}}">{{.Nick}}</a></h2>{{end}}`
	if err := parse(filename, tpl, r); err != nil {
		return fmt.Errorf("html: %w", err)
	}
	return nil
}

// List people filtered by a role.
func List(role Role) (people []string, total int, err error) {
	db := database.Connect()
	defer db.Close()
	s := peopleStmt(role, false)
	if s == "" {
		return nil, 0, fmt.Errorf("list statement %v: %w", role, ErrRole)
	}
	if total, err = database.Total(&s); err != nil {
		return nil, 0, fmt.Errorf("list totals: %w", err)
	}
	// interate through records
	rows, err := db.Query(s)
	if err != nil {
		return nil, 0, fmt.Errorf("list query: %w", err)
	} else if rows.Err() != nil {
		return nil, 0, fmt.Errorf("list rows: %w", rows.Err())
	}
	defer rows.Close()
	var persons sql.NullString
	i := 0
	for rows.Next() {
		if err = rows.Scan(&persons); err != nil {
			return nil, 0, fmt.Errorf("list row scan: %w", err)
		}
		if _, err = persons.Value(); err != nil {
			continue
		}
		i++
		people = append(people, fmt.Sprintf("%v", persons.String))
	}
	return people, total, nil
}

// Print lists people filtered by a role and summaries the results.
func Print(r Request) error {
	ppl, total, err := List(roles(r.Filter))
	if err != nil {
		return fmt.Errorf("print request: %w", err)
	}
	logs.Println(total, "matching", r.Filter, "records found")
	a := make([]string, total)
	for i, p := range ppl {
		if r.Progress {
			str.Progress(r.Filter, i+1, total)
		}
		// role
		x := strings.Split(p, ",")
		a = append(a, x...)
	}
	// title and sort names
	for i := range a {
		if r.Progress {
			str.Progress(r.Filter, i+1, total)
		}
		a[i] = strings.Title(a[i])
	}
	sort.Strings(a)
	// remove duplicates
	less := func(i, j int) bool { return a[i] < a[j] }
	unique.Slice(&a, less)
	// remove empty val
	if a[0] == "" {
		a = a[1:]
	}
	fmt.Printf("\n%s\nTotal authors %d\n", strings.Join(a, ", "), len(a))
	return nil
}

// Roles or jobs of people.
func Roles() string {
	return strings.Join(Filters(), ",")
}

func parse(filename, tpl string, r Request) error {
	grp, x, err := List(roles(r.Filter))
	if err != nil {
		return fmt.Errorf("parse list: %w", err)
	}
	f := r.Filter
	if f == "" {
		f = all
	}
	if !str.Piped() {
		logs.Println(x, "matching", f, "records found")
	}
	data, s, hr := make([]Person, len(grp)), "", false
	total := len(grp)
	for i := range grp {
		if r.Progress {
			str.Progress(r.Filter, i+1, total)
		}
		n := grp[i]
		// hr element
		switch c := n[:1]; {
		case s == "":
			s = c
		case c != s:
			s = c
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
		return fmt.Errorf("parse h2 template: %w", err)
	}
	if filename == "" {
		var buf bytes.Buffer
		wr := bufio.NewWriter(&buf)
		if err = t.Execute(wr, data); err != nil {
			return fmt.Errorf("parse h2 execute template: %w", err)
		}
		if err := wr.Flush(); err != nil {
			return fmt.Errorf("parse writer flush: %w", err)
		}
		fmt.Println(buf.String())
		return nil
	}
	switch roles(r.Filter) {
	case Artists, Coders, Musicians, Writers:
		f, err := os.Create(path.Join(viper.GetString("directory.html"), filename))
		if err != nil {
			return fmt.Errorf("parse create: %w", err)
		}
		defer f.Close()
		if err = t.Execute(f, data); err != nil {
			return fmt.Errorf("parse template execute: %w", err)
		}
	case Everyone:
		return fmt.Errorf("parse %v: %w", r.Filter, ErrFilter)
	}
	return nil
}

// peopleStmt returns a complete SQL WHERE statement where the people are filtered by a role.
func peopleStmt(role Role, softDel bool) (stmt string) {
	del := func() string {
		if softDel {
			return ""
		}
		return "AND `deletedat` IS NULL"
	}
	d := func(s string) string {
		return fmt.Sprintf(" UNION (SELECT DISTINCT %s AS pubCombined FROM files WHERE Length(%s) <> 0 %s)",
			s, s, del())
	}
	switch role {
	case Writers:
		stmt += d("credit_text")
	case Musicians:
		stmt += d("credit_audio")
	case Coders:
		stmt += d("credit_program")
	case Artists:
		stmt += d("credit_illustration")
	case Everyone:
		stmt += d("credit_text")
		stmt += d("credit_audio")
		stmt += d("credit_program")
		stmt += d("credit_illustration")
	}
	if stmt == "" {
		return stmt
	}
	stmt += " ORDER BY pubCombined"
	stmt = strings.Replace(stmt, "UNION (SELECT DISTINCT ", "(SELECT DISTINCT ", 1)
	return stmt
}

func roles(r string) Role {
	switch r {
	case "writers", "w":
		return Writers
	case "musicians", "m":
		return Musicians
	case "coders", "c":
		return Coders
	case "artists", "a":
		return Artists
	case "", all:
		return Everyone
	}
	return -1
}
