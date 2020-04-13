package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/Defacto2/df2/lib/logs"
)

const timestamp string = "2006-01-02 15:04:05"

const templ = `
-- df2 v{{.VER}}     Defacto2 MySQL file dump
-- source:        https://defacto2.net/sql
-- documentation: https://github.com/Defacto2/database

SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

INSERT INTO files ({{.INSERT}}) VALUES
{{.SQL}}
ON DUPLICATE KEY UPDATE {{.DUPE}};

-- {{now}}
`

// Data container
type Data struct {
	VER    string
	INSERT string
	SQL    string
	DUPE   string
}

type colNames []string

func (c colNames) String() string {
	return fmt.Sprintf("`%s`", strings.Join(c, "`,`"))
}

type colValues []string

func (v colValues) String() string {
	return fmt.Sprintf(`%s`, strings.Join(v, ",\n"))
}

type dupeKeys []string

func (dk dupeKeys) String() string {
	//`id` = VALUES(`id`)
	for i, n := range dk {
		dk[i] = fmt.Sprintf("`%s` = VALUES(`%s`)", n, n)
	}
	return strings.Join(dk, ",")
}

type row []string

func (r row) String() string {
	s := strings.Join(r, ",\t")
	s = strings.ReplaceAll(s, `\\'`, `\'`)
	s = strings.ReplaceAll(s, `"'`, `'`)
	s = strings.ReplaceAll(s, `'"`, `'`)
	return fmt.Sprintf("(%s)", s)
}

// columns returns the column names of table.
func columns(table string) ([]string, error) {
	db := Connect()
	defer db.Close()
	// LIMIT 0 quickly returns an empty set
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM `%s` LIMIT 0", table))
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	return columns, nil
}

func rows(table string, limit int) ([]string, error) {
	stmt := fmt.Sprintf("SELECT * FROM `%s` LIMIT %d", table, limit)
	if limit < 0 {
		stmt = fmt.Sprintf("SELECT * FROM `%s`", table)
	}
	db := Connect()
	defer db.Close()
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	result := make([]string, len(columns))
	var sql []string
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		logs.Check(err)
		for i := range values {
			t := strings.ToLower(types[i].DatabaseTypeName())
			v := values[i]
			switch {
			case string(v) == "":
				result[i] = "NULL"
			case strings.Contains(t, "char"):
				result[i] = fmt.Sprintf(`'%s'`, strings.ReplaceAll(fmt.Sprintf(`%s`, v), `'`, `\'`))
			case strings.Contains(t, "int"):
				result[i] = fmt.Sprintf("%s", v)
			case t == "text":
				result[i] = fmt.Sprintf(`'%q'`, strings.ReplaceAll(fmt.Sprintf(`%s`, v), `'`, `\'`))
			case t == "datetime":
				t, err := time.Parse(time.RFC3339, string(v))
				if err != nil {
					return nil, err
				}
				result[i] = fmt.Sprintf("'%s'", t.Format(timestamp))
			default:
				return nil, fmt.Errorf("db export rows: unsupported mysql column type %q with value %s", t, string(v))
			}
		}
		var r row = result
		sql = append(sql, fmt.Sprint(r))
	}
	return sql, nil
}

// ExportFiles outputs a MySQL 5.7 compatible SQL import files table statement.
func ExportFiles(limit uint, ver string) {
	col, err := columns("files")
	logs.Check(err)
	var names colNames = col
	var dupes dupeKeys = col
	l := int(limit)
	if limit == 0 {
		l = -1 // list all
	}
	vals, err := rows("files", l)
	logs.Check(err)
	var values colValues = vals
	dat := Data{
		ver,
		fmt.Sprint(names),
		fmt.Sprint(values),
		fmt.Sprint(dupes)}
	// template functions
	fm := make(template.FuncMap)
	fm["now"] = now // now()
	t := template.Must(template.New("statement").Funcs(fm).Parse(fmt.Sprintf(`%s`, templ)))
	err = t.Execute(os.Stdout, dat)
	logs.Check(err)
}

func now() string {
	var l, _ = time.LoadLocation("UTC")
	return time.Now().In(l).Format(timestamp)
}
