// Package recent is a work in progress JSON generator to display the most recent files on the file.
// It is intended to replace https://defacto2.net/welcome/recentfiles.
package recent

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/recent/internal/file"
)

var ErrJSON = errors.New("data fails json validation")

// List recent files as a JSON document.
func List(limit uint, compress bool) error {
	db := database.Connect()
	defer db.Close()
	query := sqlRecent(limit, false)
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("list query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("list rows: %w", rows.Err())
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("list columns: %w", err)
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	f := file.Files{Cols: [...]string{"uuid", "urlid", "title"}}
	for rows.Next() {
		if err = rows.Scan(scanArgs...); err != nil {
			return fmt.Errorf("list rows next: %w", err)
		} else if values == nil {
			continue
		}
		var v file.Thumb
		v.Parse(values)
		f.Data = append(f.Data, [...]string{v.UUID, v.URLID, v.Title})
	}
	jsonData, err := json.Marshal(f)
	if err != nil {
		return fmt.Errorf("list json marshal: %w", err)
	}
	jsonData = append(jsonData, []byte("\n")...)
	var out bytes.Buffer
	if !compress {
		if err = json.Indent(&out, jsonData, "", "    "); err != nil {
			return fmt.Errorf("list json indent: %w", err)
		}
	} else if err = json.Compact(&out, jsonData); err != nil {
		return fmt.Errorf("list json compact: %w", err)
	}
	if _, err = out.WriteTo(os.Stdout); err != nil {
		return fmt.Errorf("list write to: %w", err)
	}
	if ok := json.Valid(jsonData); !ok {
		return fmt.Errorf("list json validate: %w", ErrJSON)
	}
	return nil
}

func sqlRecent(limit uint, includeSoftDeletes bool) string {
	const (
		sel = "SELECT id,uuid,record_title,group_brand_for,group_brand_by,filename," +
			"date_issued_year,createdat,updatedat FROM files"
		where = " WHERE deletedat IS NULL"
		order = " ORDER BY createdat DESC"
	)
	stmt := sel
	if includeSoftDeletes {
		stmt += where
	}
	return stmt + order + " LIMIT " + strconv.Itoa(int(limit))
}
