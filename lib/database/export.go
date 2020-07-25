package database

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
	"github.com/mholt/archiver/v3"
	"github.com/spf13/viper"
)

type Table int

const (
	Files Table = iota
	Groups
	Netresources
	Users
)

func (t Table) String() string {
	switch t {
	case Files:
		return "files"
	case Groups:
		return "groups"
	case Netresources:
		return "netresources"
	case Users:
		return "users"
	}
	return ""
}

const (
	timestamp string      = "2006-01-02 15:04:05"
	fsql      os.FileMode = 0664
	fo                    = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
)

// Tbls are the available tables in the database.
var tbls = []string{Files.String(), Groups.String(), Netresources.String(), Users.String()}
var Tbls = strings.Join(tbls, ", ")

type colNames []string

func (c colNames) String() string {
	return fmt.Sprintf("`%s`", strings.Join(c, "`,`"))
}

type colValues []string

func (v colValues) String() string {
	return strings.Join(v, ",\n")
}

type dupeKeys []string

func (dk dupeKeys) String() string {
	for i, n := range dk {
		dk[i] = fmt.Sprintf("`%s` = VALUES(`%s`)", n, n)
	}
	return fmt.Sprintf("\nON DUPLICATE KEY UPDATE %s", strings.Join(dk, ", "))
}

type row []string

func (r row) String() string {
	s := strings.Join(r, ",\t")
	s = strings.ReplaceAll(s, `\\'`, `\'`)
	s = strings.ReplaceAll(s, `"'`, `'`)
	s = strings.ReplaceAll(s, `'"`, `'`)
	return fmt.Sprintf("(%s)", s)
}

// Flags are command line arguments.
type Flags struct {
	Compress bool   // Compress and save the output
	CronJob  bool   //
	Parallel bool   // Run --table=all queries in parallel
	Save     bool   // Save the output uncompressed
	Table    string // Table of the database to use
	Type     string // Type of export (create|update)
	Version  string // df2 app version pass-through
	Limit    uint   // Limit the number of records
}

// ExportCronJob is intended for an operating system time-based job scheduler.
// It creates both create and update types exports for the files table.
func (f Flags) ExportCronJob() error {
	f.Compress, f.Limit, f.Table = true, 0, "files"
	start := time.Now()
	const delta = 2
	switch f.Parallel {
	case true:
		var wg sync.WaitGroup
		var e1, e2 error
		wg.Add(delta)
		go func(f Flags) {
			defer wg.Done()
			f.Type = "create"
			e1 = f.ExportTable()
		}(f)
		go func(f Flags) {
			defer wg.Done()
			f.Type = "update"
			e2 = f.ExportTable()
		}(f)
		wg.Wait()
		if e1 != nil {
			return fmt.Errorf("export cronjob: %w", e1)
		} else if e2 != nil {
			return fmt.Errorf("export cronjob: %w", e2)
		}
	default:
		f.Type = "create"
		if err := f.ExportTable(); err != nil {
			return fmt.Errorf("export cronjob create: %w", err)
		}
		f.Type = "update"
		if err := f.ExportTable(); err != nil {
			return fmt.Errorf("export cronjob update: %w", err)
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("cronjob export took %s\n", elapsed)
	return nil
}

// ExportDB saves or prints a MySQL 5.7 compatible SQL import database statement.
func (f Flags) ExportDB() error {
	start := time.Now()
	buf, err := f.queryTables()
	if err != nil {
		return fmt.Errorf("exportdb query: %w", err)
	}
	if err = f.write(buf); err != nil {
		return fmt.Errorf("exportdb write buffer: %w", err)
	}
	elapsed := time.Since(start)
	fmt.Printf("sql exports took %s\n", elapsed)
	return nil
}

// ExportTable saves or prints a MySQL 5.7 compatible SQL import table statement.
func (f Flags) ExportTable() error {
	buf, err := f.queryTable()
	if err != nil {
		return fmt.Errorf("export table query: %w", err)
	}
	if err := f.write(buf); err != nil {
		return fmt.Errorf("export table write buffer: %w", err)
	}
	return nil
}

// create makes the CREATE template variable value.
func (f Flags) create() (value string) {
	if f.Type != "create" {
		return ""
	}
	switch f.Table {
	case Files.String():
		value += newFilesTempl
	case Groups.String():
		value += newGroupsTempl
	case Netresources.String():
		value += newNetresourcesTempl
	case Users.String():
		value += newUsersTempl
	}
	return value
}

// fileName to use when writing SQL to a file.
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

// queryDB requests columns and values of f.Table to create an INSERT INTO ? VALUES ? SQL statement.
func (f Flags) queryDB() (*bytes.Buffer, error) {
	if err := checkTable(f.Table); err != nil {
		return nil, fmt.Errorf("query db table: %w", err)
	}
	col, err := columns(f.Table)
	if err != nil {
		return nil, fmt.Errorf("query db columns: %w", err)
	}
	var names colNames = col
	l := int(f.Limit)
	if f.Limit == 0 {
		l = -1 // list all
	}
	vals, err := rows(f.Table, l)
	if err != nil {
		return nil, fmt.Errorf("query db rows: %w", err)
	}
	var values colValues = vals
	data := TablesData{
		Table:   f.Table,
		Columns: fmt.Sprint(names),
		Rows:    fmt.Sprint(values),
	}
	tmpl, err := template.New("insert").Parse("INSERT INTO {{.Table}} ({{.Columns}}) VALUES\n{{.Rows}};")
	if err != nil {
		return nil, fmt.Errorf("query db template: %w", err)
	}
	var b bytes.Buffer
	if err = tmpl.Execute(&b, data); err != nil {
		return nil, fmt.Errorf("query db template execute: %w", err)
	}
	return &b, nil
}

// query generates the SQL import table statement.
func (f Flags) queryTable() (*bytes.Buffer, error) {
	if err := checkTable(f.Table); err != nil {
		return nil, fmt.Errorf("query table check: %w", err)
	}
	col, err := columns(f.Table)
	if err != nil {
		return nil, fmt.Errorf("query table columns: %w", err)
	}
	var names colNames = col
	l := int(f.Limit)
	if f.Limit == 0 {
		l = -1 // list all
	}
	vals, err := rows(f.Table, l)
	if err != nil {
		return nil, fmt.Errorf("query table rows: %w", err)
	}
	var values colValues = vals
	dat := TableData{VER: f.ver(), CREATE: f.create(), TABLE: f.Table, INSERT: fmt.Sprint(names), SQL: fmt.Sprint(values)}
	if f.Type == "update" {
		var dupes dupeKeys = col
		dat.UPDATE = fmt.Sprint(dupes)
	}
	t := template.Must(template.New("stmt").Funcs(tmplFunc()).Parse(tableTmpl))
	var b bytes.Buffer
	if err = t.Execute(&b, dat); err != nil {
		return nil, fmt.Errorf("query table template execute: %w", err)
	}
	return &b, nil
}

// queryTables generates the SQL import database and tables statement.
func (f Flags) queryTables() (buf *bytes.Buffer, err error) {
	const delta = 4
	var buf1, buf2, buf3, buf4 *bytes.Buffer
	switch f.Parallel {
	case true:
		var wg sync.WaitGroup
		var e1, e2, e3, e4 error
		wg.Add(delta)
		go func(f Flags) {
			defer wg.Done()
			buf1, e1 = f.reqDB(Files)
		}(f)
		go func(f Flags) {
			defer wg.Done()
			buf2, e2 = f.reqDB(Groups)
		}(f)
		go func(f Flags) {
			defer wg.Done()
			buf3, e3 = f.reqDB(Netresources)
		}(f)
		go func(f Flags) {
			defer wg.Done()
			buf4, e4 = f.reqDB(Users)
		}(f)
		wg.Wait()
		for _, err := range []error{e1, e2, e3, e4} {
			if err != nil {
				return nil, fmt.Errorf("query tables: %w", err)
			}
		}
	default:
		buf1, err = f.reqDB(Files)
		if err != nil {
			return nil, fmt.Errorf("query tables %s: %w", Files.String(), err)
		}
		buf2, err = f.reqDB(Groups)
		if err != nil {
			return nil, fmt.Errorf("query tables %s: %w", Groups.String(), err)
		}
		buf3, err = f.reqDB(Netresources)
		if err != nil {
			return nil, fmt.Errorf("query tables %s: %w", Netresources.String(), err)
		}
		buf4, err = f.reqDB(Users)
		if err != nil {
			return nil, fmt.Errorf("query tables %s: %w", Users.String(), err)
		}
	}
	data := TablesTmp{
		VER: f.ver(),
		DB:  newDBTempl,
		CREATE: []TablesData{
			{newFilesTempl, buf1.String(), ""},
			{newGroupsTempl, buf2.String(), ""},
			{newNetresourcesTempl, buf3.String(), ""},
			{newUsersTempl, buf4.String(), ""},
		},
	}
	tmpl, err := template.New("test").Funcs(tmplFunc()).Parse(tablesTmpl)
	if err != nil {
		return nil, fmt.Errorf("query tables template: %w", err)
	}
	var b bytes.Buffer
	if err = tmpl.Execute(&b, &data); err != nil {
		return nil, fmt.Errorf("query tables template execute: %w", err)
	}
	return &b, nil
}

// reqDB requests an INSERT INTO ? VALUES ? SQL statement for table.
func (f Flags) reqDB(t Table) (*bytes.Buffer, error) {
	if t.String() == "" {
		return nil, fmt.Errorf("reqdb table: %w", ErrNoTable)
	}
	c := Flags{
		Table: t.String(),
		Type:  "create",
		Limit: f.Limit,
	}
	buf, err := c.queryDB()
	if err != nil {
		return nil, fmt.Errorf("reqdb: %w", err)
	}
	return buf, nil
}

// ver pads the df2 tool version value for use in templates.
func (f Flags) ver() string {
	const maxChars = 9
	pad := maxChars - len(f.Version)
	if pad < 0 {
		return f.Version
	}
	return fmt.Sprintf("%s%s", f.Version, strings.Repeat(" ", pad))
}

// write the buffer to stdout, an SQL file or a compressed SQL file.
func (f Flags) write(buf *bytes.Buffer) error {
	name := path.Join(viper.GetString("directory.sql"), f.fileName())
	switch {
	case f.Compress:
		name += bz2
		file, err := os.OpenFile(name, fo, fsql)
		if err != nil {
			return fmt.Errorf("flags write compress open %q: %w", name, err)
		}
		defer file.Close()
		if err = archiver.NewBz2().Compress(buf, file); err != nil {
			return fmt.Errorf("flags write bz2 compression %q: %w", file.Name(), err)
		}
		stat, err := file.Stat()
		if err != nil {
			return fmt.Errorf("flags write compress stat: %w", err)
		}
		logs.Printf("Saved %s to %s\n", humanize.Bytes(uint64(stat.Size())), name)
		return file.Close()
	case f.Save:
		file, err := os.OpenFile(name, fo, fsql)
		if err != nil {
			return fmt.Errorf("flags write save open: %w", err)
		}
		defer file.Close()
		n, err := io.Copy(file, buf)
		if err != nil {
			return fmt.Errorf("flags write save io copy: %w", err)
		}
		logs.Printf("Saved %s to %s\n", humanize.Bytes(uint64(n)), name)
		return file.Close()
	default:
		if _, err := io.WriteString(os.Stdout, buf.String()); err != nil {
			return fmt.Errorf("flags write io writer: %w", err)
		}
		return nil
	}
}

// checkTable validates table against the list of Tbls.
func checkTable(table string) error {
	for _, t := range tbls {
		if strings.ToLower(table) == t {
			return nil
		}
	}
	return fmt.Errorf("check table failure %q != %s: %w", table, Tbls, ErrNoTable)
}

// columns returns the column names of table.
func columns(table string) (columns []string, err error) {
	db := Connect()
	defer db.Close()
	// LIMIT 0 quickly returns an empty set
	rows, err := db.Query("SELECT * FROM ? LIMIT 0", table)
	if err != nil {
		return nil, fmt.Errorf("columns query: %w", err)
	} else if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("columns query rows: %w", rows.Err())
	}
	defer rows.Close()
	columns, err = rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("columns rows: %w", err)
	}
	return columns, nil
}

// format the value based on the database type name column type.
func format(b sql.RawBytes, colType string) (string, error) {
	switch {
	case string(b) == "":
		return "NULL", nil
	case strings.Contains(colType, "char"):
		return fmt.Sprintf(`'%s'`, strings.ReplaceAll(fmt.Sprintf(`%s`, b), `'`, `\'`)), nil
	case strings.Contains(colType, "int"):
		return fmt.Sprintf("%s", b), nil
	case colType == "text":
		return fmt.Sprintf(`'%q'`, strings.ReplaceAll(fmt.Sprintf(`%s`, b), `'`, `\'`)), nil
	case colType == "datetime":
		s := string(b)
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return "", fmt.Errorf("format parse datetime %q: %w", s, err)
		}
		return fmt.Sprintf("'%s'", t.Format(timestamp)), nil
	default:
		return "", fmt.Errorf("format invalid value %v: %w", b, ErrColType)
	}
}

// rows returns the values of table.
func rows(table string, limit int) (values []string, err error) {
	db := Connect()
	defer db.Close()
	var rows *sql.Rows
	if limit < 0 {
		rows, err = db.Query("SELECT * FROM ?", table)
	} else {
		rows, err = db.Query("SELECT * FROM ? LIMIT ?", table, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("rows query: %w", err)
	} else if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows query rows: %w", rows.Err())
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("rows columns: %w", err)
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("rows column types: %w", err)
	}
	vals := make([]sql.RawBytes, len(columns))
	dest := make([]interface{}, len(vals))
	for i := range vals {
		dest[i] = &vals[i]
	}
	for rows.Next() {
		result := make([]string, len(columns))
		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("rows next: %w", err)
		}
		for i := range vals {
			result[i], err = format(vals[i], strings.ToLower(types[i].DatabaseTypeName()))
			return nil, fmt.Errorf("rows %q: %w", i, err)
		}
		var r row = result
		values = append(values, fmt.Sprint(r))
	}
	return values, db.Close()
}

// template functions.
func tmplFunc() template.FuncMap {
	fm := make(template.FuncMap)
	fm["now"] = utc
	return fm
}

// utc returns the current UTC date and time in a MySQL timestamp format.
func utc() string {
	l, _ := time.LoadLocation("UTC")
	return time.Now().In(l).Format(timestamp)
}
