package database

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
	"github.com/mholt/archiver/v3"
	"github.com/spf13/viper"
)

const timestamp string = "2006-01-02 15:04:05"

// TODO implement export type
const templ = `
-- df2 v{{.VER}} Defacto2 MySQL {{.TABLE}} dump
-- source:        https://defacto2.net/sql
-- documentation: https://github.com/Defacto2/database

SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

INSERT INTO {{.TABLE}} ({{.INSERT}}) VALUES
{{.SQL}}
ON DUPLICATE KEY UPDATE {{.DUPE}};

-- {{now}}
`

// Data container
type Data struct {
	VER    string
	TABLE  string
	INSERT string
	SQL    string
	DUPE   string
}

// Flags are command line arguments
type Flags struct {
	Compress bool   // Compress and save the output
	Limit    uint   // Limit the number of records
	Save     bool   // Save the output uncompressed
	Table    string // Table of the database to use
	Type     string // Type of export (create|update)
	Version  string // df2 app version pass-through
}

// Tables available in the database
var Tables = []string{"files", "groups", "netresources", "users"}

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

// Export saves or prints a MySQL 5.7 compatible SQL import table statement.
func (f Flags) Export() {
	var (
		buf  *bytes.Buffer
		err  error
		file *os.File
		name string
	)
	buf, err = f.query()
	logs.Check(err)
	name = path.Join(viper.GetString("directory.sql"), f.fileName())
	switch {
	case f.Compress:
		name += ".bz2"
		file, err = os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
		logs.Check(err)
		defer file.Close()
		var bz2 = archiver.NewBz2()
		err = bz2.Compress(buf, file)
		logs.Check(err)
		stat, err := file.Stat()
		logs.Check(err)
		logs.Printf("Saved %s to %s\n", humanize.Bytes(uint64(stat.Size())), name)
	case f.Save:
		file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
		logs.Check(err)
		defer file.Close()
		wrote, err := io.Copy(file, buf)
		logs.Check(err)
		logs.Printf("Saved %s to %s\n", humanize.Bytes(uint64(wrote)), name)
	default:
		io.WriteString(os.Stdout, buf.String())
	}
}

func (f Flags) fileName() string {
	l, y, t := "", "export", "table"
	if f.Type != "" {
		y = f.Type
	}
	if f.Limit > 0 {
		l = fmt.Sprintf("%d_", f.Limit)
	}
	if f.Table != "" {
		t = f.Table
	}
	return fmt.Sprintf("d2-%s_%s%s.sql", y, l, t)
}

func checkTable(table string) error {
	for _, t := range Tables {
		if strings.ToLower(table) == t {
			return nil
		}
	}
	return fmt.Errorf("export query: unsupported database table '%s', please choose either %v", table, strings.Join(Tables, ", "))
}

// query generates the SQL import table statement.
func (f Flags) query() (*bytes.Buffer, error) {
	if err := checkTable(f.Table); err != nil {
		return nil, err
	}
	col, err := columns(f.Table)
	if err != nil {
		return nil, err
	}
	var names colNames = col
	var dupes dupeKeys = col
	l := int(f.Limit)
	if f.Limit == 0 {
		l = -1 // list all
	}
	vals, err := rows(f.Table, l)
	if err != nil {
		return nil, err
	}
	var values colValues = vals
	dat := Data{
		f.ver(),
		f.Table,
		fmt.Sprint(names),
		fmt.Sprint(values),
		fmt.Sprint(dupes)}
	// template functions
	fm := make(template.FuncMap)
	fm["now"] = now // now()
	t := template.Must(template.New("statement").Funcs(fm).Parse(fmt.Sprintf(`%s`, templ)))
	var b bytes.Buffer
	err = t.Execute(&b, dat)
	logs.Check(err)
	return &b, err
}

func (f Flags) ver() string {
	pad := 9 - len(f.Version) // 9 is the maximum number of characters
	if pad < 0 {
		return f.Version
	}
	return fmt.Sprintf("%s%s", f.Version, strings.Repeat(" ", pad))
}

func now() string {
	var l, _ = time.LoadLocation("UTC")
	return time.Now().In(l).Format(timestamp)
}
