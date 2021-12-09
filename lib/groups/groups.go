// Package groups deals with group names and their initialisms.
package groups

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/groups/internal/acronym"
	"github.com/Defacto2/df2/lib/groups/internal/group"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/spf13/viper"
)

const htm = ".htm"

// Request flags for group functions.
type Request struct {
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

// Count returns the number of file entries associated with a named group.
func Count(name string) (int, error) {
	if name == "" {
		return 0, nil
	}
	db := database.Connect()
	defer db.Close()
	n, count := name, 0
	row := db.QueryRow("SELECT COUNT(*) FROM files WHERE group_brand_for=? OR "+
		"group_brand_for LIKE '?,%%' OR group_brand_for LIKE '%%, ?,%%' OR "+
		"group_brand_for LIKE '%%, ?' OR group_brand_by=? OR group_brand_by "+
		"LIKE '?,%%' OR group_brand_by LIKE '%%, ?,%%' OR group_brand_by LIKE '%%, ?'", n, n)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("group count row scan: %w", err)
	}
	return count, db.Close()
}

// Cronjob is used for system automation to generate dynamic HTML pages.
func Cronjob(force bool) error {
	for _, tag := range Wheres() {
		last, err := database.LastUpdate()
		if err != nil {
			return fmt.Errorf("group cronjob last update: %w", err)
		}
		f := tag + htm
		n := path.Join(viper.GetString("directory.html"), f)
		update := true
		if !force {
			update, err = database.FileUpdate(n, last)
		}
		switch {
		case err != nil:
			return fmt.Errorf("group cronjob file update: %w", err)
		case !update:
			logs.Printf("%s has nothing to update (%s)\n", tag, n)
		default:
			r := Request{tag, true, true, false}
			if err := r.HTML(f); err != nil {
				return fmt.Errorf("group cronjob html: %w", err)
			}
		}
	}
	return nil
}

// DataList prints an auto-complete list for HTML input elements.
func (r Request) DataList(filename string) error {
	// <option value="Bitchin ANSI Design" label="BAD (Bitchin ANSI Design)">
	tpl := `{{range .}}{{if .Initialism}}<option value="{{.Name}}" label="{{.Initialism}} ({{.Name}})">{{end}}`
	tpl += `<option value="{{.Name}}" label="{{.Name}}">{{end}}`
	if err := r.parse(filename, tpl); err != nil {
		return fmt.Errorf("datalist parse template: %w", err)
	}
	return nil
}

// HTML prints a snippet listing links to each group, with an optional file count.
func (r Request) HTML(filename string) error {
	// <h2><a href="/g/13-omens">13 OMENS</a> 13O</h2><hr>
	tpl := `{{range .}}{{if .Hr}}<hr>{{end}}<h2><a href="/g/{{.ID}}">{{.Name}}</a>`
	tpl += `{{if .Initialism}} ({{.Initialism}}){{end}}{{if .Count}} <small>({{.Count}})</small>{{end}}</h2>{{end}}`
	if err := r.parse(filename, tpl); err != nil {
		return fmt.Errorf("html parse template: %w", err)
	}
	return nil
}

func (r Request) files(group string) (total int, err error) {
	if r.Counts {
		total, err = Count(group)
		if err != nil {
			return 0, fmt.Errorf("request total files for %q: %w", group, err)
		}
		return total, nil
	}
	return 0, nil
}

func (r Request) initialism(group string) (string, error) {
	if r.Initialisms {
		name, err := acronym.Get(group)
		if err != nil {
			return "", fmt.Errorf("request initialism for %q: %w", group, err)
		}
		return name, nil
	}
	return "", nil
}

func (r Request) iterate(groups ...string) (*[]Result, error) {
	piped := str.Piped()
	total := len(groups)
	data := make([]Result, total)
	var (
		hr         bool
		lastLetter string
	)
	for i, grp := range groups {
		if !piped && !logs.Quiet && r.Progress {
			str.Progress(r.Filter, i+1, total)
		}
		lastLetter, hr = group.UseHr(lastLetter, grp)
		c, err := r.files(grp)
		if err != nil {
			return nil, fmt.Errorf("iterate group file %q: %w", grp, err)
		}
		init, err := r.initialism(grp)
		if err != nil {
			return nil, fmt.Errorf("iterate group initialism %q: %w", grp, err)
		}
		data[i] = Result{
			ID:         MakeSlug(grp),
			Name:       grp,
			Count:      c,
			Initialism: init,
			Hr:         hr,
		}
	}
	return &data, nil
}

func (r Request) parse(filename, tmpl string) error {
	groups, total, err := group.List(r.Filter)
	if err != nil {
		return fmt.Errorf("parse group: %w", err)
	}
	if !str.Piped() {
		if f := r.Filter; f == "" {
			logs.Println(total, "matching (all) records found")
		} else {
			p := path.Join(viper.GetString("directory.html"), filename)
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
	if filename == "" {
		var buf bytes.Buffer
		wr := bufio.NewWriter(&buf)
		if err = t.Execute(wr, &data); err != nil {
			return fmt.Errorf("parse template execute: %w", err)
		}
		if err := wr.Flush(); err != nil {
			return fmt.Errorf("parse writer flush: %w", err)
		}
		fmt.Println(buf.String())
		return nil
	}
	switch group.Get(r.Filter) {
	case group.BBS, group.FTP, group.Group, group.Magazine:
		f, err := os.Create(path.Join(viper.GetString("directory.html"), filename))
		if err != nil {
			return fmt.Errorf("parse create: %w", err)
		}
		defer f.Close()
		// prepend html
		s := "<div class=\"pagination-statistics\"><span class=\"label label-default\">"
		s += fmt.Sprintf("%d %s sites</span></div><div class=\"columns-list\" id=\"organisationDrillDown\">", total, r.Filter)
		if _, err = f.WriteString(s); err != nil {
			return fmt.Errorf("prepend html writestring: %w", err)
		}
		// html template
		if err = t.Execute(f, &data); err != nil {
			return fmt.Errorf("parse t execute: %w", err)
		}
		// append html
		if _, err := f.WriteString("</div>\n"); err != nil {
			return fmt.Errorf("append html writestring: %w", err)
		}
	case group.None:
		return fmt.Errorf("parse %q: %w", r.Filter, group.ErrFilter)
	}
	return nil
}

// MakeSlug takes a string and makes it into a URL friendly slug.
func MakeSlug(s string) string {
	n := database.TrimSP(s)
	n = acronym.Trim(n)
	n = strings.ReplaceAll(n, "-", "_")
	n = strings.ReplaceAll(n, ", ", "*")
	n = strings.ReplaceAll(n, " & ", " ampersand ")
	re := regexp.MustCompile(` (\d)`)
	n = re.ReplaceAllString(n, `-$1`)
	re = regexp.MustCompile(`[^A-Za-z0-9 \-\+.\_\*]`) // remove all chars except these
	n = re.ReplaceAllString(n, ``)
	n = strings.ToLower(n)
	re = regexp.MustCompile(` ([a-z])`)
	n = re.ReplaceAllString(n, `-$1`)
	return n
}

// Print list organisations or groups filtered by a name and summaries the results.
func Print(r Request) (total int, err error) {
	grp, total, err := group.List(r.Filter)
	if err != nil {
		return 0, fmt.Errorf("print list groups: %w", err)
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
			c, err := Count(g)
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

// Variations creates format variations for a named group.
func Variations(name string) ([]string, error) {
	if name == "" {
		return []string{}, nil
	}
	name = strings.ToLower(name)
	vars := []string{name}
	s := strings.Split(name, " ")
	if a := strings.Join(s, ""); name != a {
		vars = append(vars, a)
	}
	if b := strings.Join(s, "-"); name != b {
		vars = append(vars, b)
	}
	if c := strings.Join(s, "_"); name != c {
		vars = append(vars, c)
	}
	if d := strings.Join(s, "."); name != d {
		vars = append(vars, d)
	}
	if init, err := Initialism(name); err == nil && init != "" {
		vars = append(vars, strings.ToLower(init))
	} else if err != nil {
		return nil, fmt.Errorf("variations %q: %w", name, err)
	}
	return vars, nil
}

// Wheres are group categories.
func Wheres() []string {
	return []string{
		group.BBS.String(),
		group.FTP.String(),
		group.Group.String(),
		group.Magazine.String()}
}

// Initialism returns a named group initialism or acronym.
func Initialism(name string) (string, error) {
	g := acronym.Group{Name: name}
	if err := g.Get(); err != nil {
		return "", fmt.Errorf("initialism get %q: %w", name, err)
	}
	return g.Initialism, nil
}

// Fix any malformed group names found in the database.
func Fix(simulate bool) error {
	names, _, err := group.List("")
	if err != nil {
		return err
	}
	c, start := 0, time.Now()
	for _, name := range names {
		if r := group.Clean(name, simulate); r {
			c++
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
		logs.Printcr("no group fixes needed")
	}
	elapsed := time.Since(start).Seconds()
	logs.Print(fmt.Sprintf(", time taken %.1f seconds\n", elapsed))
	return nil
}
