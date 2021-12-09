// Package people deals with people, person names, aliases and their roles.
package people

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/groups"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/people/internal/person"
	"github.com/Defacto2/df2/lib/people/internal/role"
	"github.com/Defacto2/df2/lib/str"
	"github.com/campoy/unique"
)

// Request flags for people functions.
type Request struct {
	Filter   string // Filter people by category.
	Counts   bool   // Counts the person's total files.
	Progress bool   // Progress counter when requesting database data.
}

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
	return []string{
		role.Artists.String(),
		role.Coders.String(),
		role.Musicians.String(),
		role.Writers.String()}
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

// Print lists people filtered by a role and summaries the results.
func Print(r Request) error {
	ppl, total, err := role.List(role.Roles(r.Filter))
	if err != nil {
		return fmt.Errorf("print request: %w", err)
	}
	logs.Println(total, "matching", r.Filter, "records found")
	var a []string
	//	a := make([]string, total)
	for i, p := range ppl {
		if r.Progress {
			str.Progress(r.Filter, i+1, total)
		}
		// role
		a = append(a, strings.Split(p, ",")...)
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
	grp, x, err := role.List(role.Roles(r.Filter))
	if err != nil {
		return fmt.Errorf("parse list: %w", err)
	}
	f := r.Filter
	if f == "" {
		f = "all"
	}
	if !str.Piped() {
		logs.Println(x, "matching", f, "records found")
	}
	data, s, hr := make(person.Persons, len(grp)), "", false
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
		data[i] = person.Person{
			ID:   groups.MakeSlug(n),
			Nick: n,
			Hr:   hr,
		}
	}
	return data.Template(filename, tpl, r.Filter)
}

// Fix any malformed group names found in the database.
func Fix(simulate bool) error {
	c, start := 0, time.Now()
	for _, r := range []role.Role{role.Artists, role.Coders, role.Musicians, role.Writers} {
		credits, _, err := role.List(r)
		if err != nil {
			return err
		}
		for _, credit := range credits {
			if r := role.Clean(credit, r, simulate); r {
				c++
			}
		}
	}
	switch {
	case c > 0 && simulate:
		logs.Printcrf("%d fixes required", c)
		logs.Simulate()
	case c == 1:
		logs.Printcr("1 fix applied")
	case c > 0:
		logs.Printcrf("%d fixes applied", c)
	default:
		logs.Printcr("no people fixes needed")
	}
	elapsed := time.Since(start).Seconds()
	logs.Print(fmt.Sprintf(", time taken %.1f seconds\n", elapsed))

	return nil
}
