// Package recent is a work in progress JSON generator to display the most recent files on the file.
// It is intended to replace https://defacto2.net/welcome/recentfiles.
package recent

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/recent/internal/file"
)

var ErrJSON = errors.New("data fails json validation")

// List recent files as a JSON document.
func List(db *sql.DB, w io.Writer, limit uint, compress bool) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	query := sqlRecent(limit, false)
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("list query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("list rows: %w", rows.Err())
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("list columns: %w", err)
	}
	values := make([]sql.RawBytes, len(cols))
	args := make([]any, len(values))
	for i := range values {
		args[i] = &values[i]
	}
	f := file.Files{Cols: [...]string{"uuid", "urlid", "title"}}
	for rows.Next() {
		if err = rows.Scan(args...); err != nil {
			return fmt.Errorf("list rows next: %w", err)
		}
		th := file.Thumb{}
		th.Scan(values)
		f.Data = append(f.Data, [...]string{th.UUID, th.URLID, th.Title})
	}
	return list(w, f, compress)
}

func list(w io.Writer, f file.Files, compress bool) error {
	if w == nil {
		w = io.Discard
	}
	b, err := json.Marshal(f)
	if err != nil {
		return fmt.Errorf("list json marshal: %w", err)
	}
	b = append(b, []byte("\n")...)
	dst := bytes.Buffer{}
	switch compress {
	case true:
		if err := json.Compact(&dst, b); err != nil {
			return fmt.Errorf("list json compact: %w", err)
		}
	case false:
		if err := json.Indent(&dst, b, "", "    "); err != nil {
			return fmt.Errorf("list json indent: %w", err)
		}
	}
	if _, err := dst.WriteTo(w); err != nil {
		return fmt.Errorf("list write to: %w", err)
	}
	if ok := json.Valid(b); !ok {
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
