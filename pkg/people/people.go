// Package people deals with people, person names, aliases and their roles.
package people

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/people/internal/person"
	"github.com/Defacto2/df2/pkg/people/internal/role"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/campoy/unique"
	"github.com/spf13/viper"
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
func (r Request) HTML(name string) error {
	return HTML(name, r)
}

// Cronjob is used for system automation to generate dynamic HTML pages.
func Cronjob(force bool) error {
	// as the jobs take time, check the locations before querying the database
	for _, tag := range Wheres() {
		f := tag + htm
		d := viper.GetString("directory.html")
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
		if err := cronjob(tag, force); err != nil {
			return err
		}
	}
	return nil
}

func cronjob(tag string, force bool) error {
	last, err := database.LastUpdate()
	if err != nil {
		return fmt.Errorf("cronjob lastupdate: %w", err)
	}
	f := tag + htm
	n := path.Join(viper.GetString("directory.html"), f)
	update := true
	if !force {
		update, err = database.FileUpdate(n, last)
	}
	switch {
	case err != nil:
		return fmt.Errorf("cronjob fileupdate: %w", err)
	case !update:
		logs.Printf("%s has nothing to update (%s)\n", tag, n)
	default:
		r := Request{
			Filter:   tag,
			Counts:   true,
			Progress: false,
		}
		if force {
			r.Progress = true
		}
		if err := r.HTML(f); err != nil {
			return fmt.Errorf("group cronjob html: %w", err)
		}
	}
	return nil
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
		role.Writers.String(),
	}
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
	logs.Printf("%d matching %s records found\n", total, r.Filter)
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
	title := cases.Title(language.English, cases.NoLower)
	for i := range a {
		if r.Progress {
			str.Progress(r.Filter, i+1, total)
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
	logs.Printf("\n%s\nTotal authors %d\n", strings.Join(a, ", "), len(a))
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
		logs.Printf("%d matching %s records found\n", x, f)
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
			ID:   groups.Slug(n),
			Nick: n,
			Hr:   hr,
		}
	}
	return data.Template(filename, tpl, r.Filter)
}

// Fix any malformed group names found in the database.
func Fix() error {
	c, start := 0, time.Now()
	for _, r := range []role.Role{role.Artists, role.Coders, role.Musicians, role.Writers} {
		credits, _, err := role.List(r)
		if err != nil {
			return err
		}
		for _, credit := range credits {
			if r := role.Clean(credit, r); r {
				c++
			}
		}
	}
	switch {
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

// Wheres are group categories.
func Wheres() []string {
	return []string{
		role.Artists.String(),
		role.Coders.String(),
		role.Musicians.String(),
		role.Writers.String(),
	}
}
