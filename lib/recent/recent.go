package recent

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/database"
	"github.com/hako/durafmt"
)

// This will eventually replace https://defacto2.net/welcome/recentfiles

// File data for new thumbnails
type File struct {
	UUID    string `json:"uuid"`
	URLID   string `json:"urlid"`
	Title   string `json:"title"`
	timeAgo string
	title   string
	group   string
	year    int
}

type data [3]string

type files struct {
	Cols [3]string `json:"COLUMNS"`
	Data []data    `json:"DATA"`
}

func (f *File) parse(values []sql.RawBytes) {
	if id := string(values[0]); id != "" {
		f.URLID = database.ObfuscateParam(id)
	}
	f.UUID = strings.ToLower(string(values[1]))
	if t, err := time.Parse(time.RFC3339, string(values[7])); err != nil {
		f.timeAgo = "Sometime"
	} else {
		f.timeAgo = fmt.Sprint(durafmt.Parse(time.Since(t)).LimitFirstN(1))
	}
	if rt := string(values[2]); rt != "" {
		f.title = fmt.Sprintf("%s (%s)", values[2], values[5])
	} else {
		f.title = string(values[5])
	}
	if g := string(values[3]); g != "" {
		f.group = g
	} else if g := string(values[4]); g != "" {
		f.group = g
	} else {
		f.group = "an unknown group"
	}
	if y := string(values[6]); y != "" {
		i, err := strconv.Atoi(y)
		if err == nil {
			f.year = i
		}
	}
	f.Title = fmt.Sprintf("%s ago, %s for %s", f.timeAgo, f.title, f.group)
	const min = 1980
	if f.year >= min {
		f.Title += fmt.Sprintf(" in %d", f.year)
	}
}

// List recent files as a JSON document.
func List(limit uint, compress bool) error {
	db := database.Connect()
	defer db.Close()
	query := sqlRecent(limit, false)
	rows, err := db.Query(query)
	if err != nil {
		return err
	} else if rows.Err() != nil {
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
	f := files{
		Cols: [3]string{"uuid", "urlid", "title"},
	}
	for rows.Next() {
		if err = rows.Scan(scanArgs...); err != nil {
			return err
		} else if values == nil {
			continue
		}
		var v File
		v.parse(values)
		f.Data = append(f.Data, [3]string{v.UUID, v.URLID, v.Title})
	}
	jsonData, err := json.Marshal(f)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	if !compress {
		if err := json.Indent(&out, jsonData, "", "    "); err != nil {
			return err
		}
	} else if err := json.Compact(&out, jsonData); err != nil {
		return err
	}
	if _, err = out.WriteTo(os.Stdout); err != nil {
		return err
	}
	if ok := json.Valid(jsonData); !ok {
		err := errors.New("recent list: jsonData fails JSON encoding validation")
		return err
	}
	return nil
}

func sqlRecent(limit uint, includeSoftDeletes bool) (sql string) {
	const (
		sel   = "SELECT id,uuid,record_title,group_brand_for,group_brand_by,filename,date_issued_year,createdat,updatedat FROM files"
		where = " WHERE deletedat IS NULL"
		order = " ORDER BY createdat DESC"
	)
	sql = sel
	if includeSoftDeletes {
		sql += where
	}
	return sql + order + " LIMIT " + strconv.Itoa(int(limit))
}
