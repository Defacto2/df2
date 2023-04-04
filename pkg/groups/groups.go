// Package groups deals with group names and their initialisms.
package groups

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups/internal/acronym"
	"github.com/Defacto2/df2/pkg/groups/internal/filter"
	"github.com/Defacto2/df2/pkg/groups/internal/rename"
	"github.com/Defacto2/df2/pkg/groups/internal/request"
	"github.com/Defacto2/df2/pkg/logger"
)

var (
	ErrCronDir = errors.New("cronjob directory does not exist")
	ErrHTMLDir = errors.New("the directory.html setting is empty")
	ErrTag     = errors.New("cronjob tag cannot be an empty string")
)

// Request flags for group functions.
type Request request.Flags

// DataList prints an auto-complete list for HTML input elements.
func (r Request) DataList(db *sql.DB, w, dest io.Writer) error {
	return request.Flags(r).DataList(db, w, dest)
}

// HTML prints a snippet listing links to each group, with an optional file count.
func (r Request) HTML(db *sql.DB, w, dest io.Writer) error {
	return request.Flags(r).HTML(db, w, dest)
}

// Print a list of organisations or groups filtered by a name and summarizes the results.
func (r Request) Print(db *sql.DB, w io.Writer) (int, error) {
	return request.Print(db, w, request.Flags(r))
}

// Count returns the number of file entries associated with a named group.
func Count(db *sql.DB, name string) (int, error) {
	return filter.Count(db, name)
}

// Cronjob is used for system automation to generate dynamic HTML pages.
func Cronjob(db *sql.DB, w, dest io.Writer, tag string, force bool) error {
	if db == nil {
		return database.ErrDB
	}
	if tag == "" {
		return ErrTag
	}
	if w == nil {
		w = io.Discard
	}
	if dest == nil {
		dest = io.Discard
	}
	r := request.Flags{
		Filter:      tag,
		Counts:      true,
		Initialisms: true,
		Progress:    false,
	}
	if force {
		r.Progress = true
	}
	return r.HTML(db, w, dest)
}

// Exact returns the number of file entries that match an exact named filter.
// The casing is ignored, but comma separated multi-groups are not matched to their parents.
// The name "tristar" will match "Tristar" but will not match records using
// "Tristar, Red Sector Inc".
func Exact(db *sql.DB, name string) (int, error) {
	if db == nil {
		return 0, database.ErrDB
	}
	if name == "" {
		return 0, nil
	}
	n, count := name, 0
	row := db.QueryRow("SELECT COUNT(*) FROM files WHERE group_brand_for=? OR "+
		"group_brand_by=?", n, n)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}

// Fix any malformed group names found in the database.
func Fix(db *sql.DB, w io.Writer) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	// fix group names stored in the files table
	names, _, err := filter.List(db, w, "")
	if err != nil {
		return err
	}
	c, start := 0, time.Now()
	for _, name := range names {
		r, err := rename.Clean(db, w, name)
		if err != nil {
			return err
		}
		if r {
			c++
		}
	}
	switch {
	case c == 1:
		logger.PrintCR(w, "1 fix applied")
	case c > 0:
		logger.PrintfCR(w, "%d fixes applied", c)
	default:
		logger.PrintCR(w, "no group fixes needed")
	}
	// fix initialisms stored in the groupnames table
	fmt.Fprint(w, " and...\n")
	i, err := acronym.Fix(db)
	if err != nil {
		return err
	}
	switch i {
	case 1:
		logger.PrintCR(w, "removed a broken initialism entry")
	case 0:
		logger.PrintCR(w, "no initialism fixes needed")
	default:
		logger.PrintfCR(w, "%d broken initialism entries removed", i)
	}
	// report time taken
	elapsed := time.Since(start).Seconds()
	fmt.Fprintf(w, ", time taken %.1f seconds\n", elapsed)
	return nil
}

// Format returns a copy of name with custom formatting.
func Format(name string) string {
	return rename.Format(name)
}

// Initialism returns a named group initialism or acronym.
func Initialism(db *sql.DB, name string) (string, error) {
	if db == nil {
		return "", database.ErrDB
	}
	g := acronym.Group{Name: name}
	if err := g.Get(db); err != nil {
		return "", fmt.Errorf("initialism %q: %w", name, err)
	}
	return g.Initialism, nil
}

// List returns all the distinct groups.
func List(db *sql.DB, w io.Writer) ([]string, int, error) {
	return filter.List(db, w, "")
}

// Slug takes a string and makes it into a URL friendly slug.
func Slug(s string) string {
	return filter.Slug(s)
}

// Update replaces all instances of the group name with the new group name.
func Update(db *sql.DB, newName, group string) (int64, error) {
	return rename.Update(db, newName, group)
}

// Variations creates format variations for a named filter.
func Variations(db *sql.DB, name string) ([]string, error) {
	if db == nil {
		return nil, database.ErrDB
	}
	if name == "" {
		return nil, nil
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
	if init, err := Initialism(db, name); err == nil && init != "" {
		vars = append(vars, strings.ToLower(init))
	} else if err != nil {
		return nil, fmt.Errorf("variations %q: %w", name, err)
	}
	return vars, nil
}

// Tags are the group categories.
func Tags() []string {
	return []string{
		filter.BBS.String(),
		filter.FTP.String(),
		filter.Group.String(),
		filter.Magazine.String(),
	}
}
