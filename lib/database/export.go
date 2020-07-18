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

const timestamp string = "2006-01-02 15:04:05"

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

// TblNames are the available table in the database
var TblNames = []string{"files", "groups", "netresources", "users"}

// Format the values of these slices when used as a string

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
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
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
	}
	if err := rows.Err(); err != nil {
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
	vals := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(vals))
	for i := range vals {
		scanArgs[i] = &vals[i]
	}
	for rows.Next() {
		result := make([]string, len(columns))
		err = rows.Scan(scanArgs...)
		logs.Check(err)
		for i := range vals {
			result[i], err = format(vals[i],
				strings.ToLower(types[i].DatabaseTypeName()))
			logs.Check(err)
		}
		var r row = result
		values = append(values, fmt.Sprint(r))
	}
	return values, nil
}

// format the value based on the database type name column type.
func format(value sql.RawBytes, dtn string) (string, error) {
	switch {
	case string(value) == "":
		return "NULL", nil
	case strings.Contains(dtn, "char"):
		return fmt.Sprintf(`'%s'`, strings.ReplaceAll(fmt.Sprintf(`%s`, value), `'`, `\'`)), nil
	case strings.Contains(dtn, "int"):
		return fmt.Sprintf("%s", value), nil
	case dtn == "text":
		return fmt.Sprintf(`'%q'`, strings.ReplaceAll(fmt.Sprintf(`%s`, value), `'`, `\'`)), nil
	case dtn == "datetime":
		t, err := time.Parse(time.RFC3339, string(value))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("'%s'", t.Format(timestamp)), nil
	default:
		return "", fmt.Errorf("db export rows: unsupported mysql column type %q with value %s", dtn, string(value))
	}
}

// write the buffer to stdout, an SQL file or a compressed SQL file.
func (f Flags) write(buf *bytes.Buffer) {
	var (
		err  error
		file *os.File
		name string
	)
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
		_, err := io.WriteString(os.Stdout, buf.String())
		logs.Check(err)
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
func (f Flags) ExportCronJob() {
	f.Compress = true
	f.Limit = 0
	f.Table = "files"
	start := time.Now()
	switch f.Parallel {
	case true:
		var wg sync.WaitGroup
		wg.Add(2)
		go func(f Flags) {
			defer wg.Done()
			f.Type = "create"
			f.ExportTable()
		}(f)
		go func(f Flags) {
			defer wg.Done()
			f.Type = "update"
			f.ExportTable()
		}(f)
		wg.Wait()
	default:
		f.Type = "create"
		f.ExportTable()
		f.Type = "update"
		f.ExportTable()
	}
	elapsed := time.Since(start)
	fmt.Printf("cronjob export took %s\n", elapsed)
}

// ExportTable saves or prints a MySQL 5.7 compatible SQL import table statement.
func (f Flags) ExportTable() {
	buf, err := f.queryTable()
	logs.Check(err)
	f.write(buf)
}

// checkTable validates table against the TblNames collection.
func checkTable(table string) error {
	for _, t := range TblNames {
		if strings.ToLower(table) == t {
			return nil
		}
	}
	return fmt.Errorf("export query: unsupported database table '%s', please choose either %v", table, strings.Join(TblNames, ", "))
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
	t := template.Must(template.New("statement").Funcs(tmplFunc()).Parse(tableTmpl))
	var b bytes.Buffer
	err = t.Execute(&b, dat)
	logs.Check(err)
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
func (f Flags) ExportDB() {
	start := time.Now()
	buf, err := f.queryTables()
	logs.Check(err)
	f.write(buf)
	elapsed := time.Since(start)
	fmt.Printf("sql exports took %s\n", elapsed)
}

// queryTables generates the SQL import database and tables statement.
func (f Flags) queryTables() (*bytes.Buffer, error) {
	var buf1, buf2, buf3, buf4 *bytes.Buffer
	var err error
	switch f.Parallel {
	case true:
		var wg sync.WaitGroup
		wg.Add(4)
		go func(f Flags) {
			defer wg.Done()
			buf1 = f.reqDB("files")
		}(f)
		go func(f Flags) {
			defer wg.Done()
			buf2 = f.reqDB("groups")
		}(f)
		go func(f Flags) {
			defer wg.Done()
			buf3 = f.reqDB("groups")
		}(f)
		go func(f Flags) {
			defer wg.Done()
			buf4 = f.reqDB("users")
		}(f)
		wg.Wait()
	default:
		buf1 = f.reqDB("files")
		buf2 = f.reqDB("groups")
		buf3 = f.reqDB("netresources")
		buf4 = f.reqDB("users")
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
		panic(err)
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, &data)
	logs.Check(err)
	return &b, err
}

// reqDB requests an INSERT INTO ? VALUES ? SQL statement for table.
func (f Flags) reqDB(table string) *bytes.Buffer {
	c1 := Flags{
		Table: table,
		Type:  "create",
		Limit: f.Limit,
	}
	buf, err := c1.queryDB()
	logs.Check(err)
	return buf
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
	fm["now"] = now // now()
	return fm
}

// now returns the current UTC date and time in a MySQL timestamp format.
func now() string {
	var l, _ = time.LoadLocation("UTC")
	return time.Now().In(l).Format(timestamp)
}

// ver pads the df2 tool version value for use in templates.
func (f Flags) ver() string {
	pad := 9 - len(f.Version) // 9 is the maximum number of characters
	if pad < 0 {
		return f.Version
	}
	return fmt.Sprintf("%s%s", f.Version, strings.Repeat(" ", pad))
}
