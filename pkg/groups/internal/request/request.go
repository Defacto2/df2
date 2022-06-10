package request

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/Defacto2/df2/pkg/groups/internal/acronym"
	"github.com/Defacto2/df2/pkg/groups/internal/group"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/spf13/viper"
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
func (r Flags) DataList(name string) error {
	// <option value="Bitchin ANSI Design" label="BAD (Bitchin ANSI Design)">
	tpl := `{{range .}}{{if .Initialism}}<option value="{{.Name}}" label="{{.Initialism}} ({{.Name}})">{{end}}`
	tpl += `<option value="{{.Name}}" label="{{.Name}}">{{end}}`
	if err := r.Parse(name, tpl); err != nil {
		return fmt.Errorf("template: %w", err)
	}
	return nil
}

// HTML prints a snippet listing links to each group, with an optional file count.
func (r Flags) HTML(name string) error {
	// <h2><a href="/g/13-omens">13 OMENS</a> 13O</h2><hr>
	tpl := `{{range .}}{{if .Hr}}<hr>{{end}}<h2><a href="/g/{{.ID}}">{{.Name}}</a>`
	tpl += `{{if .Initialism}} ({{.Initialism}}){{end}}{{if .Count}} <small>({{.Count}})</small>{{end}}</h2>{{end}}`
	if err := r.Parse(name, tpl); err != nil {
		return fmt.Errorf("template: %w", err)
	}
	return nil
}

// Files returns the number of files associated with the named group.
func (r Flags) Files(name string) (int, error) {
	if !r.Counts {
		return 0, nil
	}
	total, err := group.Count(name)
	if err != nil {
		return 0, fmt.Errorf("files %q: %w", name, err)
	}
	return total, nil
}

// Initialism returns the initialism of the named group.
func (r Flags) Initialism(name string) (string, error) {
	if !r.Initialisms {
		return "", nil
	}
	s, err := acronym.Get(name)
	if err != nil {
		return "", fmt.Errorf("initialism %q: %w", name, err)
	}
	return s, nil
}

func (r Flags) iterate(groups ...string) (*[]Result, error) {
	piped := str.Piped()
	total := len(groups)
	data := make([]Result, total)
	var (
		hr         bool
		lastLetter string
	)
	for i, grp := range groups {
		if !piped && !logs.IsQuiet() && r.Progress {
			str.Progress(r.Filter, i+1, total)
		}
		lastLetter, hr = group.UseHr(lastLetter, grp)
		c, err := r.Files(grp)
		if err != nil {
			return nil, fmt.Errorf("iterate files %q: %w", grp, err)
		}
		init, err := r.Initialism(grp)
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
func (r Flags) Parse(name, tmpl string) error { //nolint:funlen
	groups, total, err := group.List(r.Filter)
	if err != nil {
		return fmt.Errorf("parse list: %w", err)
	}
	if !str.Piped() {
		if f := r.Filter; f == "" {
			logs.Println(total, "matching (all) records found")
		} else {
			p := path.Join(viper.GetString("directory.html"), name)
			logs.Printf("%d matching %s records found (%s)\n", total, f, p)
		}
	}
	data, err := r.iterate(groups...)
	if err != nil {
		return fmt.Errorf("parse iterate: %w", err)
	}
	t, err := template.New("h2").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	if name == "" {
		var buf bytes.Buffer
		wr := bufio.NewWriter(&buf)
		if err = t.Execute(wr, &data); err != nil {
			return fmt.Errorf("parse execute: %w", err)
		}
		if err := wr.Flush(); err != nil {
			return fmt.Errorf("parse flush: %w", err)
		}
		fmt.Println(buf.String())
		return nil
	}
	switch group.Get(r.Filter) {
	case group.BBS, group.FTP, group.Group, group.Magazine:
		html := path.Join(viper.GetString("directory.html"), name)
		f, err := os.Create(html)
		if err != nil {
			if _, _ = os.Stat(viper.GetString("directory.html")); errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("parse create: parent directory is missing: %w", err)
			}
			return fmt.Errorf("parse create: %w", err)
		}
		defer f.Close()
		// prepend html
		s := "<div class=\"pagination-statistics\"><span class=\"label label-default\">"
		s += fmt.Sprintf("%d %s sites</span></div><div class=\"columns-list\" id=\"organisation-drill-down\">",
			total, r.Filter)
		if _, err = f.WriteString(s); err != nil {
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

// Print list organisations or groups filtered by a name and summaries the results.
func Print(r Flags) (total int, err error) {
	grp, total, err := group.List(r.Filter)
	if err != nil {
		return 0, fmt.Errorf("print groups: %w", err)
	}
	logs.Println(total, "matching", r.Filter, "records found")
	a := make([]string, total)
	for i, g := range grp {
		if r.Progress {
			str.Progress(r.Filter, i+1, total)
		}
		s := g
		if r.Initialisms {
			if in, err := acronym.Get(g); err != nil {
				return 0, fmt.Errorf("print initialism: %w", err)
			} else if in != "" {
				s = fmt.Sprintf("%v [%s]", s, in)
			}
		}
		// file totals
		if r.Counts {
			c, err := group.Count(g)
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
	logs.Printf("\n%s\nTotal groups %d\n", strings.Join(a, ", "), total)
	return total, nil
}
