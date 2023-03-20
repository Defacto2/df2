// Package people deals with people, person names, aliases and their roles.
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
	"github.com/Defacto2/df2/pkg/logger"
	"github.com/Defacto2/df2/pkg/people/internal/person"
	"github.com/Defacto2/df2/pkg/people/internal/role"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/campoy/unique"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	ErrCJDir  = errors.New("cronjob directory does not exist")
	ErrCJFile = errors.New("cronjob file to save html does not exist")
	ErrCfg    = errors.New("the directory.html setting is empty")
)

const htm = ".htm"

// Request flags for people functions.
type Request struct {
	Filter   string // Filter people by category.
	Counts   bool   // Counts the person's total files.
	Progress bool   // Progress counter when requesting database data.
}

// HTML prints a snippet listing links to each group, with an optional file count.
func (r Request) HTML(db *sql.DB, w io.Writer, name, directory string) error {
	return HTML(db, w, name, directory, r)
}

// Cronjob is used for system automation to generate dynamic HTML pages.
func Cronjob(db *sql.DB, w io.Writer, directory string, force bool) error {
	// as the jobs take time, check the locations before querying the database
	for _, tag := range Wheres() {
		f := tag + htm
		d := directory
		n := path.Join((d), f)
		if d == "" {
			return fmt.Errorf("cronjob: %w", ErrCfg)
		}
		if _, err := os.Stat(d); errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("cronjob: %w: %s", ErrCJDir, d)
		}
		if _, err := os.Stat(n); errors.Is(err, fs.ErrNotExist) {
			if err1 := directories.Touch(n); err1 != nil {
				return fmt.Errorf("cronjob: %w: %s", err1, n)
			}
		}
	}
	for _, tag := range Wheres() {
		if err := cronjob(db, w, directory, tag, force); err != nil {
			return err
		}
	}
	return nil
}

func cronjob(db *sql.DB, w io.Writer, directory, tag string, force bool) error {
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
		r := Request{
			Filter:   tag,
			Counts:   true,
			Progress: false,
		}
		if force {
			r.Progress = true
		}
		if err := r.HTML(db, w, f, directory); err != nil {
			return fmt.Errorf("group cronjob html: %w", err)
		}
	}
	return nil
}

// DataList prints an auto-complete list for HTML input elements.
func DataList(db *sql.DB, w io.Writer, filename, directory string, r Request) error {
	// <option value="$YN (Syndicate BBS)" label="$YN (Syndicate BBS)">
	tpl := `{{range .}}<option value="{{.Nick}}" label="{{.Nick}}">{{end}}`
	if err := parse(db, w, filename, directory, tpl, r); err != nil {
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
func HTML(db *sql.DB, w io.Writer, filename, directory string, r Request) error {
	// <h2><a href="/p/ben">Ben</a></h2><hr>
	tpl := `{{range .}}{{if .HR}}<hr>{{end}}<h2><a href="/p/{{.ID}}">{{.Nick}}</a></h2>{{end}}`
	if err := parse(db, w, filename, directory, tpl, r); err != nil {
		return fmt.Errorf("html: %w", err)
	}
	return nil
}

// Print lists people filtered by a role and summaries the results.
func Print(db *sql.DB, w io.Writer, r Request) error {
	ppl, total, err := role.List(db, w, role.Roles(r.Filter))
	if err != nil {
		return fmt.Errorf("print request: %w", err)
	}
	fmt.Fprintf(w, "%d matching %s records found\n", total, r.Filter)
	var a []string
	//	a := make([]string, total)
	for i, p := range ppl {
		if r.Progress {
			str.Progress(w, r.Filter, i+1, total)
		}
		// role
		a = append(a, strings.Split(p, ",")...)
	}
	// title and sort names
	title := cases.Title(language.English, cases.NoLower)
	for i := range a {
		if r.Progress {
			str.Progress(w, r.Filter, i+1, total)
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

func parse(db *sql.DB, w io.Writer, filename, directory, tpl string, r Request) error {
	grp, x, err := role.List(db, w, role.Roles(r.Filter))
	if err != nil {
		return fmt.Errorf("parse list: %w", err)
	}
	f := r.Filter
	if f == "" {
		f = "all"
	}
	if !str.Piped() {
		fmt.Fprintf(w, "%d matching %s records found\n", x, f)
	}
	data, s, hr := make(person.Persons, len(grp)), "", false
	total := len(grp)
	for i := range grp {
		if r.Progress {
			str.Progress(w, r.Filter, i+1, total)
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
			ID:   groups.Slug(n),
			Nick: n,
			HR:   hr,
		}
	}
	return data.Template(w, filename, directory, tpl, r.Filter)
}

// Fix any malformed group names found in the database.
func Fix(db *sql.DB, w io.Writer) error {
	c, start := 0, time.Now()
	for _, r := range []role.Role{role.Artists, role.Coders, role.Musicians, role.Writers} {
		credits, _, err := role.List(db, w, r)
		if err != nil {
			return err
		}
		for _, credit := range credits {
			if r := role.Clean(db, w, credit, r); r {
				c++
			}
		}
	}
	switch {
	case c == 1:
		logger.Printcr(w, "1 fix applied")
	case c > 0:
		logger.Printcrf(w, "%d fixes applied", c)
	default:
		logger.Printcr(w, "no people fixes needed")
	}
	elapsed := time.Since(start).Seconds()
	fmt.Fprintf(w, ", time taken %.1f seconds\n", elapsed)
	return nil
}

// Wheres are group categories.
func Wheres() []string {
	return []string{
		role.Artists.String(),
		role.Coders.String(),
		role.Musicians.String(),
		role.Writers.String(),
	}
}
