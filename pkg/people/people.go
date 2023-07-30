// Package people handles scene persons, their names, aliases and roles.
package people

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/Defacto2/df2/pkg/people/internal/person"
	"github.com/Defacto2/df2/pkg/people/internal/role"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/campoy/unique"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	ErrCronDir = errors.New("cronjob directory does not exist")
	ErrHTMLDir = errors.New("the directory.html setting is empty")
)

const htm = ".htm"

// Flags for people functions.
type Flags struct {
	Filter   string // Filter people by category.
	Counts   bool   // Counts the person's total files.
	Progress bool   // Progress counter when requesting database data.
}

// HTML prints a snippet listing links to each group, with an optional file count.
func (f Flags) HTML(db *sql.DB, w io.Writer, dest string) error {
	return HTML(db, w, dest, f)
}

// Cronjob is used for system automation to generate dynamic HTML pages.
func Cronjob(db *sql.DB, w io.Writer, directory string, force bool) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	// as the jobs take time, check the locations before querying the database
	for _, tag := range Tags() {
		f := tag + htm
		d := directory
		n := path.Join((d), f)
		if d == "" {
			return fmt.Errorf("cronjob: %w", ErrHTMLDir)
		}
		if _, err := os.Stat(d); errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("cronjob: %w: %s", ErrCronDir, d)
		}
		if _, err := os.Stat(n); errors.Is(err, fs.ErrNotExist) {
			if err1 := directories.Touch(n); err1 != nil {
				return fmt.Errorf("cronjob: %w: %s", err1, n)
			}
		}
	}
	for _, tag := range Tags() {
		if err := cronjob(db, w, directory, tag, force); err != nil {
			return err
		}
	}
	return nil
}

func cronjob(db *sql.DB, w io.Writer, directory, tag string, force bool) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	last, err := database.LastUpdate(db)
	if err != nil {
		return fmt.Errorf("cronjob lastupdate: %w", err)
	}
	f := tag + htm
	n := path.Join(directory, f)
	update := true
	if !force {
		update, err = database.FileUpdate(n, last)
	}
	switch {
	case err != nil:
		return fmt.Errorf("cronjob fileupdate: %w", err)
	case !update:
		fmt.Fprintf(w, "%s has nothing to update (%s)\n", tag, n)
	default:
		r := Flags{
			Filter:   tag,
			Counts:   true,
			Progress: false,
		}
		if force {
			r.Progress = true
		}
		if err := r.HTML(db, w, n); err != nil {
			return fmt.Errorf("group cronjob html: %w", err)
		}
	}
	return nil
}

// DataList prints an auto-complete list for HTML input elements.
func DataList(db *sql.DB, w io.Writer, dest string, f Flags) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	// <option value="$YN (Syndicate BBS)" label="$YN (Syndicate BBS)">
	tpl := `{{range .}}<option value="{{.Nick}}" label="{{.Nick}}">{{end}}`
	if err := parse(db, w, dest, tpl, f); err != nil {
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
		role.Writers.String(),
	}
}

// HTML prints a snippet listing links to each person.
func HTML(db *sql.DB, w io.Writer, dest string, f Flags) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	// <h2><a href="/p/ben">Ben</a></h2><hr>
	tpl := `{{range .}}{{if .HR}}<hr>{{end}}<h2><a href="/p/{{.ID}}">{{.Nick}}</a></h2>{{end}}`
	if err := parse(db, w, dest, tpl, f); err != nil {
		return fmt.Errorf("html: %w", err)
	}
	return nil
}

// Print lists people filtered by a role and summaries the results.
func Print(db *sql.DB, w io.Writer, f Flags) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	ppl, ppls, err := role.List(db, w, role.Roles(f.Filter))
	if err != nil {
		return fmt.Errorf("print request: %w", err)
	}
	fmt.Fprintf(w, "%d matching %s records found\n", ppls, f.Filter)
	var a []string
	//	a := make([]string, total)
	for i, p := range ppl {
		if f.Progress {
			str.Progress(w, f.Filter, i+1, ppls)
		}
		// role
		a = append(a, strings.Split(p, ",")...)
	}
	// title and sort names
	title := cases.Title(language.English, cases.NoLower)
	for i := range a {
		if f.Progress {
			str.Progress(w, f.Filter, i+1, ppls)
		}
		a[i] = title.String(a[i])
	}
	sort.Strings(a)
	// remove duplicates
	less := func(i, j int) bool { return a[i] < a[j] }
	unique.Slice(&a, less)
	// remove empty val
	if a[0] == "" {
		a = a[1:]
	}
	fmt.Fprintf(w, "\n%s\nTotal authors %d\n", strings.Join(a, ", "), len(a))
	return nil
}

// Roles or jobs of people.
func Roles() string {
	return strings.Join(Filters(), ",")
}

func parse(db *sql.DB, w io.Writer, dest, tmpl string, f Flags) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	ppl, ppls, err := role.List(db, w, role.Roles(f.Filter))
	if err != nil {
		return fmt.Errorf("parse list: %w", err)
	}
	x := f.Filter
	if x == "" {
		x = "all"
	}
	if !str.Piped() {
		fmt.Fprintf(w, "%d matching %s records found\n", ppls, x)
	}
	data, s, hr := make(person.Persons, len(ppl)), "", false
	total := len(ppl)
	for i := range ppl {
		if f.Progress {
			str.Progress(w, f.Filter, i+1, total)
		}
		n := ppl[i]
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
			ID:   groups.Slug(n),
			Nick: n,
			HR:   hr,
		}
	}
	if dest == "" {
		return data.TemplateW(w, tmpl)
	}
	return data.Template(dest, tmpl)
}

// Fix any malformed names found in the database.
func Fix(db *sql.DB, w io.Writer) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	c, start := 0, time.Now()
	for _, r := range []role.Role{role.Artists, role.Coders, role.Musicians, role.Writers} {
		credits, _, err := role.List(db, w, r)
		if err != nil {
			return err
		}
		for _, credit := range credits {
			if r, err := role.Clean(db, w, credit, r); err != nil {
				return err
			} else if r {
				c++
			}
		}
	}
	str.Total(w, c, "people fixes")
	str.TimeTaken(w, time.Since(start).Seconds())
	return nil
}

// Tags are categories of people.
func Tags() []string {
	return []string{
		role.Artists.String(),
		role.Coders.String(),
		role.Musicians.String(),
		role.Writers.String(),
	}
}
