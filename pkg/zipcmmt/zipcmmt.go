package zipcmmt

import (
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/zipcmmt/internal/cmmt"
)

const (
	errPrefix = "zipcmmt"
	fixStmt   = `SELECT id, uuid, filename, filesize, file_magic_type FROM files WHERE filename LIKE "%.zip"`
)

func Fix(db *sql.DB, w io.Writer, ascii, unicode, overwrite, summary bool) error {
	start := time.Now()
	dir := directories.Init(false)
	rows, err := db.Query(fixStmt)
	if err != nil {
		return fmt.Errorf("%s, db query: %w", errPrefix, err)
	} else if rows.Err() != nil {
		return fmt.Errorf("%s, db rows: %w", errPrefix, rows.Err())
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		z := cmmt.Zipfile{
			ASCII:     ascii,
			Unicode:   unicode,
			Overwrite: overwrite,
		}
		if err := rows.Scan(&z.ID, &z.UUID, &z.Name, &z.Size, &z.Magic); err != nil {
			return fmt.Errorf("%s rows scan: %w", errPrefix, err)
		}
		i++
		if ok := z.CheckDownload(dir.UUID); !ok {
			continue
		}
		if ok := z.CheckCmmtFile(dir.UUID); !ok {
			continue
		}
		if err := z.Save(w, dir.UUID); err != nil {
			fmt.Fprintln(w, err)
		}
	}
	elapsed := time.Since(start).Seconds()
	if ascii || unicode {
		fmt.Fprintln(w)
	}
	if summary {
		fmt.Fprintf(w, "%d zip archives scanned for comments", i)
		fmt.Fprintf(w, ", time taken %.3f seconds\n", elapsed)
	}
	return nil
}
