// Package zipcontent scans archives for file and directory content.
package zipcontent

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/zipcontent/internal/record"
	"github.com/Defacto2/df2/lib/zipcontent/internal/stat"
)

func Fix() error {
	s := stat.Init()
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
	scanArgs := make([]interface{}, len(values))
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
		r := newRec(values, s.BasePath)
		s.Columns = columns
		s.Values = &values
		if err := r.Iterate(&s); err != nil {
			fmt.Println()
			log.Printf("%s\n", err)
			continue
		}
		logs.Println()
	}
	s.Summary()
	return nil
}

func newRec(values []sql.RawBytes, path string) record.Record {
	const id, uuid, filename, readme = 0, 1, 4, 6
	return record.Record{
		ID:   string(values[id]),
		UUID: string(values[uuid]),
		Name: string(values[filename]),
		NFO:  string(values[readme]),
		File: filepath.Join(path, string(values[uuid])),
	}
}

func where() string {
	const s = "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`updatedat`,`retrotxt_readme`"
	const w = " WHERE file_zip_content IS NULL AND (`filename` LIKE '%.zip' OR `filename`" +
		" LIKE '%.rar' OR `filename` LIKE '%.7z')"
	return fmt.Sprintf("%s FROM `files` %s", s, w)
}
