package zipcmmt

import (
	"fmt"
	"log"
	"time"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/zipcmmt/internal/cmmt"
)

const (
	errPrefix = "zipcmmt"
	fixStmt   = `SELECT id, uuid, filename, filesize, file_magic_type FROM files WHERE filename LIKE "%.zip"`
)

func Fix(ascii, unicode, overwrite, summary bool) error {
	start := time.Now()
	dir, db := directories.Init(false), database.Connect()
	defer db.Close()
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
		if err := z.Save(dir.UUID); err != nil {
			log.Println(err)
		}
	}
	elapsed := time.Since(start).Seconds()
	if ascii || unicode {
		logs.Println()
	}
	if summary {
		logs.Print(fmt.Sprintf("%d zip archives scanned for comments", i))
		logs.Print(fmt.Sprintf(", time taken %.3f seconds\n", elapsed))
	}
	return nil
}
