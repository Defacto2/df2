package export

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Defacto2/df2/pkg/database/internal/my57"
	"github.com/Defacto2/df2/pkg/database/internal/templ"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/dustin/go-humanize"
	"github.com/mholt/archiver"
	"github.com/spf13/viper"
)

var (
	ErrColType  = errors.New("the value type is not usable with the mysql column")
	ErrNoTable  = errors.New("unknown database table")
	ErrNoMethod = errors.New("unknown database export type")
)

// A database table.
type Table int

const (
	Files        Table = iota // Files records.
	Groups                    // Groups names.
	Netresources              // Netresources for online websites.
	Users                     // Users are site logins.
)

const (
	cr                     = "create"
	up                     = "update"
	null                   = "NULL"
	apostrophe             = "'"
	timestamp  string      = "2006-01-02 15:04:05"
	fsql       os.FileMode = 0o664
	fo                     = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
)

func (t Table) String() string {
	return [...]string{"files", "groups", "netresources", "users"}[t]
}

// Tbls are the available tables in the database.
func Tbls() string {
	s := []string{
		Files.String(),
		Groups.String(),
		Netresources.String(),
		Users.String(),
	}
	return strings.Join(s, ", ")
}

// Method to interact with the database.
type Method int

const (
	Create Method = iota // Create uses the CREATE SQL statement to make a new record.
	Insert               // Insert uses the UPDATE SQL statement to edit an existing record.
)

func (m Method) String() string {
	return [...]string{cr, up}[m]
}

// Flags are command line arguments.
type Flags struct {
	Compress bool   // Compress and save the output
	CronJob  bool   // Run in an automated mode
	Parallel bool   // Run --table=all queries in parallel
	Save     bool   // Save the output uncompressed
	Table    Table  // Table of the database to use
	Method   Method // Method to export
	Tables   string // --table flag result
	Type     string // Type of export (create|update)
	Version  string // df2 app version pass-through
	Limit    uint   // Limit the number of records
}

// Run is intended for an operating system time-based job scheduler.
// It creates both create and update types exports for the files table.
func (f *Flags) Run() error {
	f.Compress, f.Limit, f.Table = true, 0, Files
	start := time.Now()
	const delta = 2
	switch f.Parallel {
	case true:
		var wg sync.WaitGroup
		var e1, e2 error
		wg.Add(delta)
		go func(f *Flags) {
			defer wg.Done()
			f.Method = Create
			e1 = f.ExportTable()
		}(f)
		go func(f *Flags) {
			defer wg.Done()
			f.Method = Insert
			e2 = f.ExportTable()
		}(f)
		wg.Wait()
		if e1 != nil {
			return fmt.Errorf("run e1: %w", e1)
		} else if e2 != nil {
			return fmt.Errorf("run e2: %w", e2)
		}
	default:
		f.Method = Create
		if err := f.ExportTable(); err != nil {
			return fmt.Errorf("run create: %w", err)
		}
		f.Method = Insert
		if err := f.ExportTable(); err != nil {
			return fmt.Errorf("run update: %w", err)
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("cronjob export took %s\n", elapsed)
	return nil
}

// DB saves or prints a MySQL 5.7 compatible SQL import database statement.
func (f *Flags) DB() error {
	start := time.Now()
	if err := f.method(); err != nil {
		return fmt.Errorf("db: %w", err)
	}
	buf, err := f.queryTables()
	if err != nil {
		return fmt.Errorf("db query: %w", err)
	}
	if err = f.write(buf); err != nil {
		return fmt.Errorf("db write: %w", err)
	}
	elapsed := time.Since(start)
	fmt.Printf("sql exports took %s\n", elapsed)
	return nil
}

// ExportTable saves or prints a MySQL 5.7 compatible SQL import table statement.
func (f *Flags) ExportTable() error {
	if err := f.method(); err != nil {
		return fmt.Errorf("table: %w", err)
	}
	switch strings.ToLower(f.Tables) {
	case Files.String(), "f":
		f.Table = Files
	case Groups.String(), "g":
		f.Table = Groups
	case Netresources.String(), "n":
		f.Table = Netresources
	case Users.String(), "u":
		f.Table = Users
	default:
		return fmt.Errorf("invalid table: %w", ErrNoTable)
	}
	buf, err := f.queryTable()
	if err != nil {
		return fmt.Errorf("table query: %w", err)
	}
	if err := f.write(buf); err != nil {
		return fmt.Errorf("table write: %w", err)
	}
	return nil
}

// create makes the CREATE template variable value.
func (f *Flags) create() string {
	if f.Method != Create {
		return ""
	}
	s := ""
	switch f.Table {
	case Files:
		s += templ.NewFiles
	case Groups:
		s += templ.NewGroups
	case Netresources:
		s += templ.NewNetresources
	case Users:
		s += templ.NewUsers
	}
	return s
}

// fileName to use when writing SQL to a file.
func (f *Flags) fileName() string {
	l, y, t := "", "export", "table"
	if f.Method == Create || f.Method == Insert {
		y = f.Method.String()
	}
	if f.Limit > 0 {
		l = fmt.Sprintf("%d_", f.Limit)
	}
	if f.Table < Users {
		t = f.Table.String()
	}
	return fmt.Sprintf("d2-%s_%s%s.sql", y, l, t)
}

func (f *Flags) method() error {
	switch strings.ToLower(f.Type) {
	case cr, "c":
		f.Method = Create
	case up, "u":
		f.Method = Insert
	default:
		return fmt.Errorf("method %w", ErrNoMethod)
	}
	return nil
}

// queryDB requests columns and values of f.Table to create an INSERT INTO ? VALUES ? SQL statement.
func (f *Flags) queryDB() (*bytes.Buffer, error) {
	if err := f.Table.check(); err != nil {
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
	data := templ.TablesData{
		Table:   f.Table.String(),
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
func (f *Flags) queryTable() (*bytes.Buffer, error) {
	if err := f.Table.check(); err != nil {
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
	dat := templ.TableData{
		VER:    f.ver(),
		CREATE: f.create(),
		TABLE:  f.Table.String(),
		INSERT: fmt.Sprint(names),
		SQL:    fmt.Sprint(values),
	}
	if f.Method == Insert {
		var dupes dupeKeys = col
		dat.UPDATE = dupes.String()
	}
	t := template.Must(template.New("stmt").Funcs(tmplFunc()).Parse(templ.Table))
	var b bytes.Buffer
	if err = t.Execute(&b, dat); err != nil {
		return nil, fmt.Errorf("query table template execute: %w", err)
	}
	return &b, nil
}

// queryTables generates the SQL import database and tables statement.
func (f *Flags) queryTables() (*bytes.Buffer, error) {
	var buf1, buf2, buf3, buf4 *bytes.Buffer
	var err error
	switch f.Parallel {
	case true:
		buf1, buf2, buf3, buf4, err = f.queryTablesWG()
		if err != nil {
			return nil, err
		}
	default:
		buf1, buf2, buf3, buf4, err = f.queryTablesSeq()
		if err != nil {
			return nil, err
		}
	}
	data := templ.TablesTmp{
		VER: f.ver(),
		DB:  templ.NewDB,
		CREATE: []templ.TablesData{
			{Columns: templ.NewFiles, Rows: buf1.String(), Table: ""},
			{Columns: templ.NewGroups, Rows: buf2.String(), Table: ""},
			{Columns: templ.NewNetresources, Rows: buf3.String(), Table: ""},
			{Columns: templ.NewUsers, Rows: buf4.String(), Table: ""},
		},
	}
	tmpl, err := template.New("test").Funcs(tmplFunc()).Parse(templ.Tables)
	if err != nil {
		return nil, fmt.Errorf("query tables template: %w", err)
	}
	var b bytes.Buffer
	if err = tmpl.Execute(&b, &data); err != nil {
		return nil, fmt.Errorf("query tables template execute: %w", err)
	}
	return &b, nil
}

func (f *Flags) queryTablesWG() (buf1, buf2, buf3, buf4 *bytes.Buffer, err error) {
	const delta = 4
	var wg sync.WaitGroup
	var e1, e2, e3, e4 error
	wg.Add(delta)
	go func(f *Flags) {
		defer wg.Done()
		buf1, e1 = f.reqDB(Files)
	}(f)
	go func(f *Flags) {
		defer wg.Done()
		buf2, e2 = f.reqDB(Groups)
	}(f)
	go func(f *Flags) {
		defer wg.Done()
		buf3, e3 = f.reqDB(Netresources)
	}(f)
	go func(f *Flags) {
		defer wg.Done()
		buf4, e4 = f.reqDB(Users)
	}(f)
	wg.Wait()
	for _, err := range []error{e1, e2, e3, e4} {
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("query tables: %w", err)
		}
	}
	return buf1, buf2, buf3, buf4, nil
}

func (f *Flags) queryTablesSeq() (buf1, buf2, buf3, buf4 *bytes.Buffer, err error) {
	buf1, err = f.reqDB(Files)
	if err != nil {
		return nil, nil, nil, nil, qttErr(Files.String(), err)
	}
	buf2, err = f.reqDB(Groups)
	if err != nil {
		return nil, nil, nil, nil, qttErr(Groups.String(), err)
	}
	buf3, err = f.reqDB(Netresources)
	if err != nil {
		return nil, nil, nil, nil, qttErr(Netresources.String(), err)
	}
	buf4, err = f.reqDB(Users)
	if err != nil {
		return nil, nil, nil, nil, qttErr(Users.String(), err)
	}
	return buf1, buf2, buf3, buf4, nil
}

func qttErr(s string, err error) error {
	return fmt.Errorf("query tables %s: %w", s, err)
}

// reqDB requests an INSERT INTO ? VALUES ? SQL statement for table.
func (f *Flags) reqDB(t Table) (*bytes.Buffer, error) {
	if f.Table.check() != nil {
		return nil, fmt.Errorf("reqdb table: %w", ErrNoTable)
	}
	c := Flags{
		Table:  t,
		Method: Create,
		Limit:  f.Limit,
	}
	buf, err := c.queryDB()
	if err != nil {
		return nil, fmt.Errorf("reqdb: %w", err)
	}
	return buf, nil
}

// ver pads the df2 tool version value for use in templates.
func (f *Flags) ver() string {
	const maxChars = 9
	pad := maxChars - len(f.Version)
	if pad < 0 {
		return f.Version
	}
	return fmt.Sprintf("%s%s", f.Version, strings.Repeat(" ", pad))
}

// write the buffer to stdout, an SQL file or a compressed SQL file.
func (f *Flags) write(buf *bytes.Buffer) error {
	const bz2 = ".bz2"
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
	stmt := fmt.Sprintf("\nON DUPLICATE KEY UPDATE %s", strings.Join(dk, ", "))
	return stmt
}

type row []string

func (r row) String() string {
	s := strings.Join(r, ",\t")
	s = strings.ReplaceAll(s, `\\'`, `\'`)
	s = strings.ReplaceAll(s, `"'`, apostrophe)
	s = strings.ReplaceAll(s, `'"`, apostrophe)
	return fmt.Sprintf("(%s)", s)
}

func (t Table) check() error {
	const outOfRange = 4
	if t < 0 || t >= outOfRange {
		return fmt.Errorf("check table failure %q != %s: %w", t, Tbls(), ErrNoTable)
	}
	return nil
}

// columns returns the column names of table.
func columns(t Table) ([]string, error) {
	var (
		columns []string
		err     error
		rows    *sql.Rows
		db      = my57.Connect()
	)
	defer db.Close()
	switch t {
	case Files:
		rows, err = db.Query("SELECT * FROM files LIMIT 0")
		if err != nil {
			return nil, fmt.Errorf("columns files query: %w", err)
		} else if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("columns files query rows: %w", rows.Err())
		}
		defer rows.Close()
	case Groups:
		rows, err = db.Query("SELECT * FROM groups LIMIT 0")
		if err != nil {
			return nil, fmt.Errorf("columns groups query: %w", err)
		} else if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("columns groups query rows: %w", rows.Err())
		}
		defer rows.Close()
	case Netresources:
		rows, err = db.Query("SELECT * FROM netresources LIMIT 0")
		if err != nil {
			return nil, fmt.Errorf("columns netresources query: %w", err)
		} else if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("columns netresources query rows: %w", rows.Err())
		}
		defer rows.Close()
	case Users:
		rows, err = db.Query("SELECT * FROM users LIMIT 0")
		if err != nil {
			return nil, fmt.Errorf("columns users query: %w", err)
		} else if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("columns users query rows: %w", rows.Err())
		}
		defer rows.Close()
	}
	columns, err = rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("columns rows: %w", err)
	}
	return columns, nil
}

// rows returns the values of table.
func rows(t Table, limit int) ([]string, error) {
	db := my57.Connect()
	defer db.Close()
	var rows *sql.Rows
	const listAll = 0
	if limit < listAll {
		switch t {
		case Files:
			return allFiles(db)
		case Groups:
			return allGroups(db)
		case Netresources:
			return allNetresources(db)
		case Users:
			return allUsers(db)
		}
		return values(rows)
	}
	switch t {
	case Files:
		return limitFiles(limit, db)
	case Groups:
		return limitGroups(limit, db)
	case Netresources:
		return limitNetresources(limit, db)
	case Users:
		return limitUsers(limit, db)
	}
	return values(rows)
}

func allFiles(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT * FROM files")
	if err != nil {
		return nil, fmt.Errorf("rows files query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows files query rows: %w", rows.Err())
	}
	defer rows.Close()
	return values(rows)
}

func allGroups(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT * FROM groups")
	if err != nil {
		return nil, fmt.Errorf("rows groups query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows groups query rows: %w", rows.Err())
	}
	defer rows.Close()
	return values(rows)
}

func allNetresources(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return nil, fmt.Errorf("rows users query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows users query rows: %w", rows.Err())
	}
	defer rows.Close()
	return values(rows)
}

func allUsers(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return nil, fmt.Errorf("rows users query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows users query rows: %w", rows.Err())
	}
	defer rows.Close()
	return values(rows)
}

func limitFiles(limit int, db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT * FROM files LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("rows limit files query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows limit files query rows: %w", rows.Err())
	}
	defer rows.Close()
	return values(rows)
}

func limitGroups(limit int, db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT * FROM groups LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("rows limit groups query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows limit groups query rows: %w", rows.Err())
	}
	defer rows.Close()
	return values(rows)
}

func limitNetresources(limit int, db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT * FROM netresources LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("rows limit netresources query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows limit netresources query rows: %w", rows.Err())
	}
	defer rows.Close()
	return values(rows)
}

func limitUsers(limit int, db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT * FROM users LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("rows limit users query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows limit users query rows: %w", rows.Err())
	}
	defer rows.Close()
	return values(rows)
}

// format the value based on the database type name column type.
func format(b sql.RawBytes, colType string) (string, error) {
	switch {
	case string(b) == "":
		return null, nil
	case strings.Contains(colType, "char"):
		return fmt.Sprintf(`'%s'`, strings.ReplaceAll(string(b), apostrophe, `\'`)), nil
	case strings.Contains(colType, "int"):
		return string(b), nil
	case colType == "text":
		return fmt.Sprintf(`'%q'`, strings.ReplaceAll(string(b), apostrophe, `\'`)), nil
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

// template functions.
func tmplFunc() template.FuncMap {
	fm := make(template.FuncMap)
	fm["now"] = utc
	return fm
}

func values(rows *sql.Rows) ([]string, error) {
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
	v := []string{}
	for rows.Next() {
		result := make([]string, len(columns))
		if err = rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("rows next: %w", err)
		}
		for i := range vals {
			result[i], err = format(vals[i], strings.ToLower(types[i].DatabaseTypeName()))
			if err != nil {
				return nil, fmt.Errorf("rows %q: %w", i, err)
			}
		}
		var r row = result
		v = append(v, fmt.Sprint(r))
	}
	rows.Close()
	return v, nil
}

// utc returns the current UTC date and time in a MySQL timestamp format.
func utc() string {
	l, err := time.LoadLocation("UTC")
	if err != nil {
		return fmt.Sprintf("utc error: %s", err)
	}
	return time.Now().In(l).Format(timestamp)
}
