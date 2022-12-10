// Package groups deals with group names and their initialisms.
package groups

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups/internal/acronym"
	"github.com/Defacto2/df2/pkg/groups/internal/group"
	"github.com/Defacto2/df2/pkg/groups/internal/rename"
	"github.com/Defacto2/df2/pkg/groups/internal/request"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/spf13/viper"
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
func (r Request) DataList(name string) error {
	return request.Flags(r).DataList(name)
}

// HTML prints a snippet listing links to each group, with an optional file count.
func (r Request) HTML(name string) error {
	return request.Flags(r).HTML(name)
}

// Print a list of organisations or groups filtered by a name and summarizes the results.
func (r Request) Print() (total int, err error) {
	return request.Print(request.Flags(r))
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
			return fmt.Errorf("cronjob: %w: %s", ErrCJFile, n)
		}
	}
	for _, tag := range Wheres() {
		f := tag + htm
		d := viper.GetString("directory.html")
		n := path.Join((d), f)
		last, err := database.LastUpdate()
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
			logs.Printf("%s has nothing to update (%s)\n", tag, n)
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
			if err := r.HTML(f); err != nil {
				return fmt.Errorf("group cronjob html: %w", err)
			}
		}
	}
	return nil
}

// Exact returns the number of file entries that match an exact named group.
// The casing is ignored, but comma seperated multi-groups are not matched to their parents.
// The name "tristar" will match "Tristar" but will not match records using
// "Tristar, Red Sector Inc".
func Exact(name string) (int, error) {
	if name == "" {
		return 0, nil
	}
	db := database.Connect()
	defer db.Close()
	n, count := name, 0
	row := db.QueryRow("SELECT COUNT(*) FROM files WHERE group_brand_for=? OR "+
		"group_brand_by=?", n, n)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, db.Close()
}

// Fix any malformed group names found in the database.
func Fix() error {
	// fix group names stored in the files table
	names, _, err := group.List("")
	if err != nil {
		return err
	}
	c, start := 0, time.Now()
	for _, name := range names {
		if r := rename.Clean(name); r {
			c++
		}
	}
	switch {
	case c == 1:
		logs.Printcr("1 fix applied")
	case c > 0:
		logs.Printcrf("%d fixes applied", c)
	default:
		logs.Printcr("no group fixes needed")
	}
	// fix initialisms stored in the groupnames table
	logs.Print(" and...\n")
	i, err := acronym.Fix()
	if err != nil {
		return err
	}
	switch i {
	case 1:
		logs.Printcr("removed a broken initialism entry")
	case 0:
		logs.Printcr("no initialism fixes needed")
	default:
		logs.Printcrf("%d broken initialism entries removed", i)
	}
	// report time taken
	elapsed := time.Since(start).Seconds()
	logs.Print(fmt.Sprintf(", time taken %.1f seconds\n", elapsed))
	return nil
}

// Format returns a copy of name with custom formatting.
func Format(name string) string {
	return rename.Format(name)
}

// Initialism returns a named group initialism or acronym.
func Initialism(name string) (string, error) {
	g := acronym.Group{Name: name}
	if err := g.Get(); err != nil {
		return "", fmt.Errorf("initialism %q: %w", name, err)
	}
	return g.Initialism, nil
}

// Slug takes a string and makes it into a URL friendly slug.
func Slug(s string) string {
	return group.Slug(s)
}

// Update replaces all instances of the group name with the new group name.
func Update(newName, group string) (count int64, err error) {
	return rename.Update(newName, group)
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
		group.Magazine.String(),
	}
}
