// Package zipcmmt processes the text adverts that are sometimes embedded
// into Zip file archives.
package zipcmmt

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/Defacto2/df2/pkg/zipcmmt/internal/cmmt"
)

const (
	errPrefix = "zipcmmt"
	fixStmt   = `SELECT id, uuid, filename, filesize, file_magic_type FROM` +
		` files WHERE filename LIKE "%.zip"`
)

func Fix(db *sql.DB, w io.Writer, cfg conf.Config, unicode, overwrite, stdout bool) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	start := time.Now()
	dir, err := directories.Init(cfg, false)
	if err != nil {
		return err
	}
	rows, err := db.Query(fixStmt)
	if err != nil {
		return fmt.Errorf("%s, db query: %w", errPrefix, err)
	} else if rows.Err() != nil {
		return fmt.Errorf("%s, db rows: %w", errPrefix, rows.Err())
	}
	defer rows.Close()
	// create a writer specifically for the zip comment reader
	swr := io.Discard
	if stdout {
		swr = os.Stdout
	}
	i := 0
	for rows.Next() {
		z := cmmt.Zipfile{
			CP437:     unicode,
			Overwrite: overwrite,
		}
		if err := rows.Scan(&z.ID, &z.UUID, &z.Name, &z.Size, &z.Magic); err != nil {
			return fmt.Errorf("%s rows scan: %w", errPrefix, err)
		}
		i++
		if ok, err := z.Exist(dir.UUID); err != nil {
			return err
		} else if !ok {
			continue
		}
		if _, err := z.Save(swr, dir.UUID); err != nil {
			fmt.Fprintln(w, err)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%d zip archives scanned for comments", i)
	str.TimeTaken(w, time.Since(start).Seconds())
	return nil
}
