package zipcmmt

import (
	"fmt"
	"time"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/zipcmmt/internal/cmmt"
)

const (
	fixStmt = `SELECT id, uuid, filename, filesize, file_magic_type FROM files WHERE filename LIKE "%.zip"`
)

func Fix(ascii, unicode, overwrite bool) error {
	start := time.Now()
	dir, db := directories.Init(false), database.Connect()
	defer db.Close()
	rows, err := db.Query(fixStmt)
	if err != nil {
		return fmt.Errorf("fix db query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("fix rows: %w", rows.Err())
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
			return fmt.Errorf("fix rows scan: %w", err)
		}
		i++
		if ok := z.CheckDownload(dir.UUID); !ok {
			continue
		}
		if ok := z.CheckCmmtFile(dir.UUID); !ok {
			continue
		}
		if ascii || unicode {
			z.Save(dir.UUID)
			continue
		}
		z.Save(dir.UUID)
	}
	elapsed := time.Since(start).Seconds()
	if ascii || unicode {
		logs.Println()
	}
	logs.Print(fmt.Sprintf("%d zip archives scanned for comments", i))
	logs.Print(fmt.Sprintf(", time taken %.3f seconds\n", elapsed))
	return nil
}
