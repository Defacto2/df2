package request

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/Defacto2/df2/pkg/groups/internal/acronym"
	"github.com/Defacto2/df2/pkg/groups/internal/group"
	"github.com/Defacto2/df2/pkg/str"
)

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
	Hr         bool   // Inject a HR element to separate a collection of groups.
}

// DataList prints an auto-complete list for HTML input elements.
func (r Flags) DataList(w io.Writer, name, directory string) error {
	// <option value="Bitchin ANSI Design" label="BAD (Bitchin ANSI Design)">
	tpl := `{{range .}}{{if .Initialism}}<option value="{{.Name}}" label="{{.Initialism}} ({{.Name}})">{{end}}`
	tpl += `<option value="{{.Name}}" label="{{.Name}}">{{end}}`
	if err := r.Parse(w, name, directory, tpl); err != nil {
		return fmt.Errorf("template: %w", err)
	}
	return nil
}

// HTML prints a snippet listing links to each group, with an optional file count.
func (r Flags) HTML(w io.Writer, name, directory string) error {
	// <h2><a href="/g/13-omens">13 OMENS</a> 13O</h2><hr>
	tpl := `{{range .}}{{if .Hr}}<hr>{{end}}<h2><a href="/g/{{.ID}}">{{.Name}}</a>`
	tpl += `{{if .Initialism}} ({{.Initialism}}){{end}}{{if .Count}} <small>({{.Count}})</small>{{end}}</h2>{{end}}`
	if err := r.Parse(w, name, directory, tpl); err != nil {
		return fmt.Errorf("template: %w", err)
	}
	return nil
}

// Files returns the number of files associated with the named group.
func (r Flags) Files(w io.Writer, name string) (int, error) {
	if !r.Counts {
		return 0, nil
	}
	total, err := group.Count(w, name)
	if err != nil {
		return 0, fmt.Errorf("files %q: %w", name, err)
	}
	return total, nil
}

// Initialism returns the initialism of the named group.
func (r Flags) Initialism(w io.Writer, name string) (string, error) {
	if !r.Initialisms {
		return "", nil
	}
	s, err := acronym.Get(w, name)
	if err != nil {
		return "", fmt.Errorf("initialism %q: %w", name, err)
	}
	return s, nil
}

func (r Flags) iterate(w io.Writer, groups ...string) (*[]Result, error) {
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
		lastLetter, hr = group.UseHr(lastLetter, grp)
		c, err := r.Files(w, grp)
		if err != nil {
			return nil, fmt.Errorf("iterate files %q: %w", grp, err)
		}
		init, err := r.Initialism(w, grp)
		if err != nil {
			return nil, fmt.Errorf("iterate initialism %q: %w", grp, err)
		}
		data[i] = Result{
			ID:         group.Slug(grp),
			Name:       grp,
			Count:      c,
			Initialism: init,
			Hr:         hr,
		}
	}
	return &data, nil
}

// Parse the group template and save it to the named file.
// If the named file is empty, the results will be sent to stdout.
// The HTML returned to stdout is different to the markup saved
// to a file.
func (r Flags) Parse(w io.Writer, name, directory, tmpl string) error {
	groups, total, err := group.List(w, r.Filter)
	if err != nil {
		return fmt.Errorf("parse list: %w", err)
	}
	if !str.Piped() {
		if f := r.Filter; f == "" {
			fmt.Fprintln(w, total, "matching (all) records found")
		} else {
			p := path.Join(directory, name)
			fmt.Fprintf(w, "%d matching %s records found (%s)\n", total, f, p)
		}
	}
	data, err := r.iterate(w, groups...)
	if err != nil {
		return fmt.Errorf("parse iterate: %w", err)
	}
	t, err := template.New("h2").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	if name == "" {
		return noname(w, t, data)
	}
	return r.parse(name, directory, total, t, data)
}

func (r Flags) parse(name, directory string, total int, t *template.Template, data *[]Result) error {
	switch group.Get(r.Filter) {
	case group.BBS, group.FTP, group.Group, group.Magazine:
		html := path.Join(directory, name)
		f, err := os.Create(html)
		if err != nil {
			if _, _ = os.Stat(directory); errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("parse create: parent directory is missing: %w", err)
			}
			return fmt.Errorf("parse create: %w", err)
		}
		defer f.Close()
		if _, err = f.WriteString(r.prependHTML(total)); err != nil {
			return fmt.Errorf("prepend writestring: %w", err)
		}
		// html template
		if err = t.Execute(f, &data); err != nil {
			return fmt.Errorf("parse execute: %w", err)
		}
		// append html
		if _, err := f.WriteString("</div>\n"); err != nil {
			return fmt.Errorf("append writestring: %w", err)
		}
	case group.None:
		return fmt.Errorf("parse %q: %w", r.Filter, group.ErrFilter)
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

func noname(w io.Writer, t *template.Template, data *[]Result) error {
	var buf bytes.Buffer
	wr := bufio.NewWriter(&buf)
	if err := t.Execute(wr, &data); err != nil {
		return fmt.Errorf("parse execute: %w", err)
	}
	if err := wr.Flush(); err != nil {
		return fmt.Errorf("parse flush: %w", err)
	}
	fmt.Fprintln(w, buf.String())
	return nil
}

// Print list organisations or groups filtered by a name and summaries the results.
func Print(w io.Writer, r Flags) (int, error) {
	grp, total, err := group.List(w, r.Filter)
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
			if in, err := acronym.Get(w, g); err != nil {
				return 0, fmt.Errorf("print initialism: %w", err)
			} else if in != "" {
				s = fmt.Sprintf("%v [%s]", s, in)
			}
		}
		// file totals
		if r.Counts {
			c, err := group.Count(w, g)
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
