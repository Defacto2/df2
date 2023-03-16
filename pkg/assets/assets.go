// Package assets handles the site resources such as file downloads, thumbnails and backups.
package assets

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Defacto2/df2/pkg/assets/internal/scan"
	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/dustin/go-humanize"
	_ "github.com/go-sql-driver/mysql" // MySQL database driver.
	"github.com/gookit/color"
)

// Target filters the file assets.
type Target int

const (
	All       Target = iota // All files.
	Download                // Download are files for download.
	Emulation               // Emulation are files for the DOSee emulation.
	Image                   // Image and thumbnail files.
)

var (
	ErrStructNil = errors.New("structure cannot be nil")
	ErrPathEmpty = errors.New("path cannot be empty")
	ErrTarget    = errors.New("unknown target")
)

// Clean walks through and scans directories containing UUID files
// and erases any orphans that cannot be matched to the database.
func Clean(db *sql.DB, w io.Writer, cfg configger.Config, dir string, remove, human bool) error {
	d := directories.Init(cfg, false)
	return Cleaner(db, w, cfg, targetfy(dir), &d, remove, human)
}

func targetfy(s string) Target {
	switch strings.ToLower(s) {
	case "all":
		return All
	case "download":
		return Download
	case "emulation":
		return Emulation
	case "image":
		return Image
	}
	return -1
}

func Cleaner(db *sql.DB, w io.Writer, cfg configger.Config, t Target, d *directories.Dir, remove, human bool) error {
	paths := Targets(cfg, t, d)
	if paths == nil {
		return fmt.Errorf("check target %q: %w", t, ErrTarget)
	}
	// connect to the database
	rows, m, err := CreateUUIDMap(db)
	if err != nil {
		return fmt.Errorf("clean uuid map: %w", err)
	}
	fmt.Fprintln(w, "The following files do not match any UUIDs in the database")
	// parse directories
	var sum scan.Results
	for p := range paths {
		s := scan.Scan{
			Path:   paths[p],
			Delete: remove,
			Human:  human,
			M:      m,
		}
		if err := sum.Calculate(w, s, d); err != nil {
			return fmt.Errorf("clean sum calculate: %w", err)
		}
	}
	// output a summary of the Results
	fmt.Fprintln(w, color.Notice.Sprintf("\nTotal orphaned files discovered %v out of %v",
		humanize.Comma(int64(sum.Count)), humanize.Comma(int64(rows))))
	if sum.Fails > 0 {
		fmt.Fprintf(w, "assets clean: due to errors %v files could not be deleted\n",
			sum.Fails)
	}
	if len(paths) > 1 && sum.Bytes > 0 {
		pts := fmt.Sprintf("%v B", sum.Bytes)
		if human {
			pts = humanize.Bytes(uint64(sum.Bytes))
		}
		fmt.Fprintf(w, "%v drive space consumed\n", pts)
	}
	return nil
}

// CreateUUIDMap builds a map of all the unique UUID values stored in the Defacto2 database.
// Returns the total number of UUID and a collection of UUIDs.
func CreateUUIDMap(db *sql.DB) (int, database.IDs, error) {
	// count rows
	count := 0
	if err := db.QueryRow("SELECT COUNT(*) FROM `files`").Scan(&count); err != nil {
		return 0, nil, fmt.Errorf("create uuid map query row: %w", err)
	}
	// query database
	var id, uuid string
	rows, err := db.Query("SELECT `id`,`uuid` FROM `files`")
	if err != nil {
		return 0, nil, fmt.Errorf("create uuid map query: %w", err)
	}
	defer rows.Close()
	if rows.Err() != nil {
		return 0, nil, rows.Err()
	}
	total := 0
	uuids := make(database.IDs, count)
	for rows.Next() {
		if err = rows.Scan(&id, &uuid); err != nil {
			return 0, nil, fmt.Errorf("create uuid map row: %w", err)
		}
		// store record `uuid` value as a key name in the map `m` with an empty value
		uuids[uuid] = database.Empty{}
		total++
	}
	return total, uuids, db.Close()
}

func Targets(cfg configger.Config, t Target, d *directories.Dir) []string {
	if d.Base == "" {
		reset := directories.Init(cfg, false)
		d = &reset
	}
	var paths []string
	switch t {
	case All:
		paths = append(paths, d.UUID, d.Emu, d.Backup, d.Img000, d.Img400)
	case Download:
		paths = append(paths, d.UUID, d.Backup)
	case Emulation:
		paths = append(paths, d.Emu)
	case Image:
		paths = append(paths, d.Img000, d.Img400)
	}
	return paths
}
