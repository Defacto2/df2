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

const (
	timestamp string      = "2006-01-02 15:04:05"
	bz2                   = ".bz2"
	fsql      os.FileMode = 0664
	fo                    = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
)

// Flags are command line arguments
type Flags struct {
	Compress bool   // Compress and save the output
	CronJob  bool   //
	Limit    uint   // Limit the number of records
	Parallel bool   // Run --table=all queries in parallel
	Save     bool   // Save the output uncompressed
	Table    string // Table of the database to use
	Type     string // Type of export (create|update)
	Version  string // df2 app version pass-through
}

// Tbls are the available tables in the database
const Tbls = "files, groups, netresources, users"

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
	//example output: `id` = VALUES(`id`)
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

// columns returns the column names of table.
func columns(table string) (columns []string, err error) {
	db := Connect()
	defer db.Close()
	// LIMIT 0 quickly returns an empty set
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM `%s` LIMIT 0", table))
	if err != nil {
		return nil, err
	} else if err := rows.Err(); err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err = rows.Columns()
	if err != nil {
		return nil, err
	}
	return columns, nil
}

// rows returns the values of table.
func rows(table string, limit int) (values []string, err error) {
	stmt := fmt.Sprintf("SELECT * FROM `%s` LIMIT %d", table, limit)
	if limit < 0 {
		stmt = fmt.Sprintf("SELECT * FROM `%s`", table)
	}
	db := Connect()
	defer db.Close()
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	} else if err := rows.Err(); err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	vals := make([]sql.RawBytes, len(columns))
	dest := make([]interface{}, len(vals))
	for i := range vals {
		dest[i] = &vals[i]
	}
	for rows.Next() {
		result := make([]string, len(columns))
		err = rows.Scan(dest...)
		logs.Check(err)
		for i := range vals {
			result[i], err = format(vals[i],
				strings.ToLower(types[i].DatabaseTypeName()))
			logs.Check(err)
		}
		var r row = result
		values = append(values, fmt.Sprint(r))
	}
	return values, db.Close()
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
		t, err := time.Parse(time.RFC3339, string(b))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("'%s'", t.Format(timestamp)), nil
	default:
		return "", fmt.Errorf("db export rows: unsupported mysql column type %q with value %s", colType, string(b))
	}
}

// write the buffer to stdout, an SQL file or a compressed SQL file.
func (f Flags) write(buf *bytes.Buffer) error {
	name := path.Join(viper.GetString("directory.sql"), f.fileName())
	switch {
	case f.Compress:
		name += bz2
		file, err := os.OpenFile(name, fo, fsql)
		if err != nil {
			return err
		}
		defer file.Close()
		if err = archiver.NewBz2().Compress(buf, file); err != nil {
			return err
		}
		stat, err := file.Stat()
		if err != nil {
			return err
		}
		logs.Printf("Saved %s to %s\n", humanize.Bytes(uint64(stat.Size())), name)
		return file.Close()
	case f.Save:
		file, err := os.OpenFile(name, fo, fsql)
		if err != nil {
			return err
		}
		defer file.Close()
		n, err := io.Copy(file, buf)
		if err != nil {
			return err
		}
		logs.Printf("Saved %s to %s\n", humanize.Bytes(uint64(n)), name)
		return file.Close()
	default:
		_, err := io.WriteString(os.Stdout, buf.String())
		if err != nil {
			return err
		}
		return nil
	}
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
			return e1
		} else if e2 != nil {
			return e2
		}
	default:
		f.Type = "create"
		if err := f.ExportTable(); err != nil {
			return err
		}
		f.Type = "update"
		if err := f.ExportTable(); err != nil {
			return err
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("cronjob export took %s\n", elapsed)
	return nil
}

// ExportTable saves or prints a MySQL 5.7 compatible SQL import table statement.
func (f Flags) ExportTable() error {
	buf, err := f.queryTable()
	if err != nil {
		return err
	}
	if err := f.write(buf); err != nil {
		return err
	}
	return nil
}

// checkTable validates table against the list of Tbls
func checkTable(table string) error {
	tables := strings.Split(Tbls, ",")
	for _, t := range tables {
		if strings.ToLower(table) == t {
			return nil
		}
	}
	return fmt.Errorf("export query: unsupported database table '%s', please choose either %v", table, Tbls)
}

// query generates the SQL import table statement.
func (f Flags) queryTable() (*bytes.Buffer, error) {
	if err := checkTable(f.Table); err != nil {
		return nil, err
	}
	col, err := columns(f.Table)
	if err != nil {
		return nil, err
	}
	var names colNames = col
	l := int(f.Limit)
	if f.Limit == 0 {
		l = -1 // list all
	}
	vals, err := rows(f.Table, l)
	if err != nil {
		return nil, err
	}
	var values colValues = vals
	dat := TableData{
		VER:    f.ver(),
		CREATE: f.create(),
		TABLE:  f.Table,
		INSERT: fmt.Sprint(names),
		SQL:    fmt.Sprint(values),
		UPDATE: ""}
	if f.Type == "update" {
		var dupes dupeKeys = col
		dat.UPDATE = fmt.Sprint(dupes)
	}
	t := template.Must(template.New("stmt").Funcs(tmplFunc()).Parse(tableTmpl))
	var b bytes.Buffer
	if err = t.Execute(&b, dat); err != nil {
		return nil, err
	}
	return &b, err
}

// create makes the CREATE template variable value.
func (f Flags) create() (value string) {
	if f.Type != "create" {
		return ""
	}
	switch f.Table {
	case "files":
		value += newFilesTempl
	case "groups":
		value += newGroupsTempl
	case "netresources":
		value += newNetresourcesTempl
	case "users":
		value += newUsersTempl
	}
	return value
}

// ExportDB saves or prints a MySQL 5.7 compatible SQL import database statement.
func (f Flags) ExportDB() error {
	start := time.Now()
	buf, err := f.queryTables()
	if err != nil {
		return err
	}
	if err = f.write(buf); err != nil {
		return err
	}
	elapsed := time.Since(start)
	fmt.Printf("sql exports took %s\n", elapsed)
	return nil
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
			buf1, e1 = f.reqDB("files")
		}(f)
		go func(f Flags) {
			defer wg.Done()
			buf2, e2 = f.reqDB("groups")
		}(f)
		go func(f Flags) {
			defer wg.Done()
			buf3, e3 = f.reqDB("netresources")
		}(f)
		go func(f Flags) {
			defer wg.Done()
			buf4, e4 = f.reqDB("users")
		}(f)
		wg.Wait()
		for _, err := range []error{e1, e2, e3, e4} {
			if err != nil {
				return nil, err
			}
		}
	default:
		buf1, err = f.reqDB("files")
		if err != nil {
			return nil, err
		}
		buf2, err = f.reqDB("groups")
		if err != nil {
			return nil, err
		}
		buf3, err = f.reqDB("netresources")
		if err != nil {
			return nil, err
		}
		buf4, err = f.reqDB("users")
		if err != nil {
			return nil, err
		}
	}
	var data = Tables{
		VER: f.ver(),
		DB:  newDBTempl,
		CREATE: []TablesData{
			{newFilesTempl, buf1.String()},
			{newGroupsTempl, buf2.String()},
			{newNetresourcesTempl, buf3.String()},
			{newUsersTempl, buf4.String()},
		},
	}
	tmpl, err := template.New("test").Funcs(tmplFunc()).Parse(tablesTmpl)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err = tmpl.Execute(&b, &data); err != nil {
		return nil, err
	}
	return &b, err
}

// reqDB requests an INSERT INTO ? VALUES ? SQL statement for table.
func (f Flags) reqDB(table string) (*bytes.Buffer, error) {
	c := Flags{
		Table: table,
		Type:  "create",
		Limit: f.Limit,
	}
	buf, err := c.queryDB()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// queryDB requests columns and values of f.Table to create an INSERT INTO ? VALUES ? SQL statement.
func (f Flags) queryDB() (*bytes.Buffer, error) {
	if err := checkTable(f.Table); err != nil {
		return nil, err
	}
	col, err := columns(f.Table)
	if err != nil {
		return nil, err
	}
	var names colNames = col
	l := int(f.Limit)
	if f.Limit == 0 {
		l = -1 // list all
	}
	vals, err := rows(f.Table, l)
	if err != nil {
		return nil, err
	}
	var values colValues = vals
	data := TablesData{
		Columns: fmt.Sprint(names),
		Rows:    fmt.Sprint(values),
	}
	ins := fmt.Sprintf("INSERT INTO %s ({{.Columns}}) VALUES\n{{.Rows}};", f.Table)
	tmpl, err := template.New("insert").Parse(ins)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err = tmpl.Execute(&b, data); err != nil {
		return nil, err
	}
	return &b, nil
}

// template functions.
func tmplFunc() template.FuncMap {
	fm := make(template.FuncMap)
	fm["now"] = utc
	return fm
}

// utc returns the current UTC date and time in a MySQL timestamp format.
func utc() string {
	var l, _ = time.LoadLocation("UTC")
	return time.Now().In(l).Format(timestamp)
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
