// Package groups deals with group names and their initialisms.
package groups

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/groups/internal/acronym"
	"github.com/Defacto2/df2/pkg/groups/internal/group"
	"github.com/Defacto2/df2/pkg/groups/internal/rename"
	"github.com/Defacto2/df2/pkg/groups/internal/request"
	"github.com/Defacto2/df2/pkg/logger"
)

var (
	ErrCJDir  = errors.New("cronjob directory does not exist")
	ErrCJFile = errors.New("cronjob file to save html does not exist")
	ErrCfg    = errors.New("the directory.html setting is empty")
)

const htm = ".htm"

// Request flags for group functions.
type Request request.Flags

// DataList prints an auto-complete list for HTML input elements.
func (r Request) DataList(db *sql.DB, w io.Writer, name, directory string) error {
	return request.Flags(r).DataList(db, w, name, directory)
}

// HTML prints a snippet listing links to each group, with an optional file count.
func (r Request) HTML(db *sql.DB, w io.Writer, name, directory string) error {
	return request.Flags(r).HTML(db, w, name, directory)
}

// Print a list of organisations or groups filtered by a name and summarizes the results.
func (r Request) Print(db *sql.DB, w io.Writer) (int, error) {
	return request.Print(db, w, request.Flags(r))
}

// Count returns the number of file entries associated with a named group.
func Count(db *sql.DB, name string) (int, error) {
	return group.Count(db, name)
}

// Cronjob is used for system automation to generate dynamic HTML pages.
func Cronjob(db *sql.DB, w io.Writer, directory string, force bool) error {
	// as the jobs take time, check the locations before querying the database
	for _, tag := range Wheres() {
		if err := croncheck(tag, htm, directory); err != nil {
			return err
		}
	}
	for _, tag := range Wheres() {
		if err := cronjob(db, w, tag, htm, directory, force); err != nil {
			return err
		}
	}
	return nil
}

func croncheck(tag, htm, directory string) error {
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
	return nil
}

// TODO: make args into a struct!
func cronjob(db *sql.DB, w io.Writer, tag, htm, directory string, force bool) error {
	f := tag + htm
	d := directory
	n := path.Join((d), f)
	last, err := database.LastUpdate(db)
	if err != nil {
		return fmt.Errorf("cronjob lastupdate: %w", err)
	}
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
		r := request.Flags{
			Filter:      tag,
			Counts:      true,
			Initialisms: true,
			Progress:    false,
		}
		if force {
			r.Progress = true
		}
		if err := r.HTML(db, w, f, d); err != nil {
			return fmt.Errorf("group cronjob html: %w", err)
		}
	}
	return nil
}

// Exact returns the number of file entries that match an exact named group.
// The casing is ignored, but comma separated multi-groups are not matched to their parents.
// The name "tristar" will match "Tristar" but will not match records using
// "Tristar, Red Sector Inc".
func Exact(db *sql.DB, name string) (int, error) {
	if name == "" {
		return 0, nil
	}
	n, count := name, 0
	row := db.QueryRow("SELECT COUNT(*) FROM files WHERE group_brand_for=? OR "+
		"group_brand_by=?", n, n)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, db.Close()
}

// Fix any malformed group names found in the database.
func Fix(db *sql.DB, w io.Writer) error {
	// fix group names stored in the files table
	names, _, err := group.List(db, w, "")
	if err != nil {
		return err
	}
	c, start := 0, time.Now()
	for _, name := range names {
		if r := rename.Clean(db, w, name); r {
			c++
		}
	}
	switch {
	case c == 1:
		logger.Printcr(w, "1 fix applied")
	case c > 0:
		logger.Printcrf(w, "%d fixes applied", c)
	default:
		logger.Printcr(w, "no group fixes needed")
	}
	// fix initialisms stored in the groupnames table
	fmt.Fprint(w, " and...\n")
	i, err := acronym.Fix(db)
	if err != nil {
		return err
	}
	switch i {
	case 1:
		logger.Printcr(w, "removed a broken initialism entry")
	case 0:
		logger.Printcr(w, "no initialism fixes needed")
	default:
		logger.Printcrf(w, "%d broken initialism entries removed", i)
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
	g := acronym.Group{Name: name}
	if err := g.Get(db); err != nil {
		return "", fmt.Errorf("initialism %q: %w", name, err)
	}
	return g.Initialism, nil
}

// List returns all the distinct groups.
func List(db *sql.DB, w io.Writer) ([]string, int, error) {
	return group.List(db, w, "")
}

// Slug takes a string and makes it into a URL friendly slug.
func Slug(s string) string {
	return group.Slug(s)
}

// Update replaces all instances of the group name with the new group name.
func Update(db *sql.DB, newName, group string) (int64, error) {
	return rename.Update(db, newName, group)
}

// Variations creates format variations for a named group.
func Variations(db *sql.DB, name string) ([]string, error) {
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
	if init, err := Initialism(db, name); err == nil && init != "" {
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
		group.Magazine.String(),
	}
}
