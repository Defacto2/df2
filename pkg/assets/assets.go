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
	ErrDB = errors.New("database handle pointer cannot be nil")
)

type Clean struct {
	Name   string // Named section to clean.
	Remove bool   // Remove any orphaned files from the directories.
	Human  bool   // Use humanized, binary size values.
	Config configger.Config
}

// Walk through and scans directories containing UUID files
// and erases any orphans that cannot be matched to the database.
func (c Clean) Walk(db *sql.DB, w io.Writer) error {
	if db == nil {
		return ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	d, err := directories.Init(c.Config, false)
	if err != nil {
		return err
	}
	return c.Walker(db, w, targetfy(c.Name), &d)
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

func (c Clean) Walker(db *sql.DB, w io.Writer, t Target, d *directories.Dir) error {
	if db == nil {
		return ErrDB
	}
	if d == nil {
		return directories.ErrNil
	}
	if w == nil {
		w = io.Discard
	}
	paths, err := Targets(c.Config, t, d)
	if err != nil {
		return err
	}
	if paths == nil {
		return fmt.Errorf("assets walker target check %q: %w", t, scan.ErrTarget)
	}
	// connect to the database
	rows, ids, err := CreateUUIDMap(db)
	if err != nil {
		return fmt.Errorf("assets walkter uuid map: %w", err)
	}
	fmt.Fprintln(w, "The following files do not match any UUIDs in the database")
	// parse directories
	sum := scan.Results{}
	for p := range paths {
		s := scan.Scan{
			Path:   paths[p],
			Delete: c.Remove,
			Human:  c.Human,
			IDs:    ids,
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
		s := fmt.Sprintf("%v B", sum.Bytes)
		if c.Human {
			s = humanize.Bytes(uint64(sum.Bytes))
		}
		fmt.Fprintf(w, "%v drive space consumed\n", s)
	}
	return nil
}

// CreateUUIDMap builds a map of all the unique UUID values stored in the Defacto2 database.
// Returns the total number of UUID and a collection of UUIDs.
func CreateUUIDMap(db *sql.DB) (int, database.IDs, error) {
	if db == nil {
		return 0, nil, ErrDB
	}
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
	return total, uuids, nil
}

func Targets(cfg configger.Config, t Target, d *directories.Dir) ([]string, error) {
	if d == nil {
		return nil, directories.ErrNil
	}
	if d.Base == "" {
		reset, err := directories.Init(cfg, false)
		if err != nil {
			return nil, err
		}
		d = &reset
	}
	paths := []string{}
	switch t {
	case All:
		paths = append(paths, d.UUID, d.Emu, d.Backup, d.Img000, d.Img400)
	case Download:
		paths = append(paths, d.UUID, d.Backup)
	case Emulation:
		paths = append(paths, d.Emu)
	case Image:
		paths = append(paths, d.Img000, d.Img400)
	default:
		return nil, scan.ErrTarget
	}
	return paths, nil
}
