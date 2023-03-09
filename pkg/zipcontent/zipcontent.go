// Package zipcontent scans archives for file and directory content.
package zipcontent

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/zipcontent/internal/record"
	"github.com/Defacto2/df2/pkg/zipcontent/internal/scan"
)

// Fix the content of zip archives within in the database.
func Fix(summary bool) error { //nolint:cyclop
	s := scan.Init()
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(where())
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]any, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		s.Total++
	}
	rows, err = db.Query(where())
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return err
		}
		s.Count++
		r, err := record.New(values, s.BasePath)
		if err != nil {
			return err
		}
		s.Columns = columns
		s.Values = &values
		if err := r.Iterate(&s); err != nil {
			log.Printf("\n%s\n", err)
			continue
		}
		logs.Println()
	}
	if summary {
		s.Summary()
	}
	return nil
}

func where() string {
	const s = "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`updatedat`,`retrotxt_readme`"
	const w = " WHERE file_zip_content IS NULL AND (`filename` LIKE '%.zip' OR `filename`" +
		" LIKE '%.rar' OR `filename` LIKE '%.7z')"
	return fmt.Sprintf("%s FROM `files` %s", s, w)
}
