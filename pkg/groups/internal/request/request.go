// Package request obtains and writes the data of the group to various formats.
package request

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups/internal/acronym"
	"github.com/Defacto2/df2/pkg/groups/internal/filter"
	"github.com/Defacto2/df2/pkg/str"
)

var ErrPointer = errors.New("pointer value cannot be nil")

// Flags for group functions.
type Flags struct {
	Filter      string // Filter groups by category.
	Counts      bool   // Counts the group's total files.
	Initialisms bool   // Initialisms and acronyms for groups.
	Progress    bool   // Progress counter when requesting database data.
}

// Result on a group.
type Result struct {
	ID         string // ID used in URLs to link to the group.
	Name       string // Name of the group.
	Count      int    // Count file totals.
	Initialism string // Initialism or acronym.
	HR         bool   // Inject a HR element to separate a collection of groups.
}

// DataList saves an auto-complete list for HTML input elements to the dest file path.
func (r Flags) DataList(db *sql.DB, w, dest io.Writer) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	if dest == nil {
		dest = io.Discard
	}
	// <option value="Bitchin ANSI Design" label="BAD (Bitchin ANSI Design)">
	tpl := `{{range .}}{{if .Initialism}}<option value="{{.Name}}" label="{{.Initialism}} ({{.Name}})">{{end}}`
	tpl += `<option value="{{.Name}}" label="{{.Name}}">{{end}}`
	if err := r.Parse(db, w, dest, tpl); err != nil {
		return fmt.Errorf("template: %w", err)
	}
	return nil
}

// HTML prints a snippet listing links to each group, with an optional file count.
// If dest is empty the results will be send to stdout.
func (r Flags) HTML(db *sql.DB, w, dest io.Writer) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	if dest == nil {
		dest = io.Discard
	}
	// <h2><a href="/g/13-omens">13 OMENS</a> 13O</h2><hr>
	tpl := `{{range .}}{{if .HR}}<hr>{{end}}<h2><a href="/g/{{.ID}}">{{.Name}}</a>`
	tpl += `{{if .Initialism}} ({{.Initialism}}){{end}}{{if .Count}} <small>({{.Count}})</small>{{end}}</h2>{{end}}`
	if err := r.Parse(db, w, dest, tpl); err != nil {
		return fmt.Errorf("template: %w", err)
	}
	return nil
}

// Files returns the number of files associated with the named group.
func (r Flags) Files(db *sql.DB, name string) (int, error) {
	if db == nil {
		return 0, database.ErrDB
	}
	if !r.Counts {
		return 0, nil
	}
	total, err := filter.Count(db, name)
	if err != nil {
		return 0, fmt.Errorf("files %q: %w", name, err)
	}
	return total, nil
}

// Initialism returns the initialism of the named filter.
func (r Flags) Initialism(db *sql.DB, name string) (string, error) {
	if db == nil {
		return "", database.ErrDB
	}
	if !r.Initialisms {
		return "", nil
	}
	s, err := acronym.Get(db, name)
	if err != nil {
		return "", fmt.Errorf("initialism %q: %w", name, err)
	}
	return s, nil
}

func (r Flags) iterate(db *sql.DB, w io.Writer, groups ...string) (*[]Result, error) {
	if db == nil {
		return nil, database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	piped := str.Piped()
	total := len(groups)
	data := make([]Result, total)
	var (
		hr         bool
		lastLetter string
	)
	for i, grp := range groups {
		if !piped && r.Progress {
			str.Progress(w, r.Filter, i+1, total)
		}
		lastLetter, hr = filter.UseHr(lastLetter, grp)
		c, err := r.Files(db, grp)
		if err != nil {
			return nil, fmt.Errorf("iterate files %q: %w", grp, err)
		}
		init, err := r.Initialism(db, grp)
		if err != nil {
			return nil, fmt.Errorf("iterate initialism %q: %w", grp, err)
		}
		data[i] = Result{
			ID:         filter.Slug(grp),
			Name:       grp,
			Count:      c,
			Initialism: init,
			HR:         hr,
		}
	}
	return &data, nil
}

// Parse the group template and save it to the named file.
// If the dest is empty, the results will be sent to stdout.
// The HTML returned to stdout is different to the markup saved
// to a file.
func (r Flags) Parse(db *sql.DB, w, dest io.Writer, tmpl string) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	if dest == nil {
		dest = io.Discard
	}
	list, count, err := filter.List(db, w, r.Filter)
	if err != nil {
		return fmt.Errorf("parse list: %w", err)
	}
	if !str.Piped() {
		if f := r.Filter; f == "" {
			fmt.Fprintln(w, count, "matching (all) records found")
		} else {
			fmt.Fprintf(w, "%d matching %s records found\n", count, f)
		}
	}
	data, err := r.iterate(db, w, list...)
	if err != nil {
		return fmt.Errorf("parse iterate: %w", err)
	}
	t, err := template.New("h2").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	if dest == os.Stdout {
		return t.Execute(dest, &data)
	}
	return r.parse(t, data, dest, count)
}

func (r Flags) parse(t *template.Template, data *[]Result, dest io.Writer, count int) error {
	if t == nil {
		return fmt.Errorf("%w: t templte", ErrPointer)
	}
	if data == nil {
		return fmt.Errorf("%w: data result", ErrPointer)
	}
	if dest == nil {
		dest = io.Discard
	}
	switch filter.Get(r.Filter) {
	case filter.BBS, filter.FTP, filter.Group, filter.Magazine:
		fmt.Fprint(dest, r.prependHTML(count))
		if err := t.Execute(dest, &data); err != nil {
			return fmt.Errorf("parse execute: %w", err)
		}
		fmt.Fprintln(dest, "</div>")
	case filter.None:
		return fmt.Errorf("parse %q: %w", r.Filter, filter.ErrFilter)
	}
	return nil
}

func (r Flags) prependHTML(total int) string {
	now := time.Now()
	s := "<div class=\"pagination-statistics\"><span class=\"label label-default\">"
	s += fmt.Sprintf("%d %s sites</span>",
		total, r.Filter)
	s += fmt.Sprintf("&nbsp; <span class=\"label label-default\">updated, %s</span>", now.Format("2006 Jan 2"))
	s += "</div><div class=\"columns-list\" id=\"organisation-drill-down\">"
	return s
}

// Print list organisations or groups filtered by a name and summaries the results.
func Print(db *sql.DB, w io.Writer, r Flags) (int, error) {
	if db == nil {
		return 0, database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	grp, total, err := filter.List(db, w, r.Filter)
	if err != nil {
		return 0, fmt.Errorf("print groups: %w", err)
	}
	fmt.Fprintf(w, "\n\n%d matching %q records found\n", total, r.Filter)
	a := make([]string, total)
	for i, g := range grp {
		if r.Progress {
			str.Progress(w, r.Filter, i+1, total)
		}
		s := g
		if r.Initialisms {
			if in, err := acronym.Get(db, g); err != nil {
				return 0, fmt.Errorf("print initialism: %w", err)
			} else if in != "" {
				s = fmt.Sprintf("%v [%s]", s, in)
			}
		}
		// file totals
		if r.Counts {
			c, err := filter.Count(db, g)
			if err != nil {
				return 0, fmt.Errorf("print counts: %w", err)
			}
			if c > 0 {
				s = fmt.Sprintf("%v (%d)", s, c)
			}
		}
		a[i] = s
	}
	// remove empty val
	if a[0] == "" {
		a = a[1:]
	}
	fmt.Fprintf(w, "\n%s\nTotal groups %d\n", strings.Join(a, ", "), total)
	return total, nil
}
