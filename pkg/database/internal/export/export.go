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

	"github.com/Defacto2/df2/pkg/database/internal/templ"
	"github.com/dustin/go-humanize"
	"github.com/mholt/archiver"
)

var (
	ErrColType  = errors.New("the value type is not usable with the mysql column")
	ErrDB       = errors.New("database handle pointer cannot be nil")
	ErrNoTable  = errors.New("unknown database table")
	ErrMethod   = errors.New("unknown database export type")
	ErrNoMethod = errors.New("no database export type provided")
	ErrPointer  = errors.New("pointer value cannot be nil")
)

// A database table.
type Table int

const (
	Files        Table = iota // Files records.
	Groups                    // Groups names.
	Netresources              // Netresources for online websites.
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
	return [...]string{"files", "groupnames", "netresources"}[t]
}

// Tbls are the available tables in the database.
func Tbls() string {
	s := []string{
		Files.String(),
		Groups.String(),
		Netresources.String(),
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
	SQLDumps string // SQLDumps should be the value of config.SQLDumps
}

// Run is intended for an operating system time-based job scheduler.
// It creates both create and update types exports for the files table.
func (f *Flags) Run(db *sql.DB, w io.Writer) error {
	if db == nil {
		return ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	f.Compress, f.Limit, f.Table = true, 0, Files
	start := time.Now()
	const delta = 2
	mu := sync.Mutex{}
	switch f.Parallel {
	case true:
		wg := sync.WaitGroup{}
		var e1, e2 error
		wg.Add(delta)
		go func(f *Flags) {
			defer wg.Done()
			mu.Lock()
			f.Method = Create
			e1 = f.ExportTable(db, w)
			mu.Unlock()
		}(f)
		go func(f *Flags) {
			defer wg.Done()
			mu.Lock()
			f.Method = Insert
			e2 = f.ExportTable(db, w)
			mu.Unlock()
		}(f)
		wg.Wait()
		if e1 != nil {
			return fmt.Errorf("run e1: %w", e1)
		}
		if e2 != nil {
			return fmt.Errorf("run e2: %w", e2)
		}
	default:
		f.Method = Create
		if err := f.ExportTable(db, w); err != nil {
			return fmt.Errorf("run create: %w", err)
		}
		f.Method = Insert
		if err := f.ExportTable(db, w); err != nil {
			return fmt.Errorf("run update: %w", err)
		}
	}
	elapsed := time.Since(start)
	fmt.Fprintf(w, "cronjob export took %s\n", elapsed)
	return nil
}

// DB saves or prints a MySQL compatible SQL import database statement.
func (f *Flags) DB(db *sql.DB, w io.Writer) error {
	if db == nil {
		return ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	start := time.Now()
	if err := f.method(); err != nil {
		return fmt.Errorf("db: %w", err)
	}
	buf, err := f.queryTables(db)
	if err != nil {
		return fmt.Errorf("db query: %w", err)
	}
	if err = f.write(w, buf); err != nil {
		return fmt.Errorf("db write: %w", err)
	}
	elapsed := time.Since(start)
	fmt.Fprintf(w, "sql exports took %s\n", elapsed)
	return nil
}

// ExportTable saves or prints a MySQL compatible SQL import table statement.
func (f *Flags) ExportTable(db *sql.DB, w io.Writer) error {
	if db == nil {
		return ErrDB
	}
	if w == nil {
		w = io.Discard
	}
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
	default:
		return fmt.Errorf("invalid table: %w", ErrNoTable)
	}
	buf, err := f.queryTable(db)
	if err != nil {
		return fmt.Errorf("table query: %w", err)
	}
	if err := f.write(w, buf); err != nil {
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
	if f.Table < Netresources {
		t = f.Table.String()
	}
	return fmt.Sprintf("d2-%s_%s%s.sql", y, l, t)
}

func (f *Flags) method() error {
	if f.Type == "" {
		return fmt.Errorf("method %w", ErrNoMethod)
	}
	switch strings.ToLower(f.Type) {
	case cr, "c":
		f.Method = Create
	case up, "u":
		f.Method = Insert
	default:
		return fmt.Errorf("method %w: %s", ErrMethod, f.Type)
	}
	return nil
}

// queryDB requests columns and values of f.Table to create an INSERT INTO ? VALUES ? SQL statement.
func (f *Flags) queryDB(db *sql.DB) (*bytes.Buffer, error) {
	if db == nil {
		return nil, ErrDB
	}
	if err := f.Table.check(); err != nil {
		return nil, fmt.Errorf("query db table: %w", err)
	}
	col, err := columns(db, f.Table)
	if err != nil {
		return nil, fmt.Errorf("query db columns: %w", err)
	}
	names := col
	l := int(f.Limit)
	if f.Limit == 0 {
		const listAll = -1
		l = listAll
	}
	vals, err := rows(db, f.Table, l)
	if err != nil {
		return nil, fmt.Errorf("query db rows: %w", err)
	}
	values := vals
	data := templ.TablesData{
		Table:   f.Table.String(),
		Columns: fmt.Sprint(names),
		Rows:    fmt.Sprint(values),
	}
	tmpl, err := template.New("insert").Parse("INSERT INTO {{.Table}} ({{.Columns}}) VALUES\n{{.Rows}};")
	if err != nil {
		return nil, fmt.Errorf("query db template: %w", err)
	}
	b := bytes.Buffer{}
	if err = tmpl.Execute(&b, data); err != nil {
		return nil, fmt.Errorf("query db template execute: %w", err)
	}
	return &b, nil
}

// query generates the SQL import table statement.
func (f *Flags) queryTable(db *sql.DB) (*bytes.Buffer, error) {
	if db == nil {
		return nil, ErrDB
	}
	if err := f.Table.check(); err != nil {
		return nil, fmt.Errorf("query table check: %w", err)
	}
	col, err := columns(db, f.Table)
	if err != nil {
		return nil, fmt.Errorf("query table columns: %w", err)
	}
	names := col
	l := int(f.Limit)
	if f.Limit == 0 {
		l = -1 // list all
	}
	vals, err := rows(db, f.Table, l)
	if err != nil {
		return nil, fmt.Errorf("query table rows: %w", err)
	}
	values := vals
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
	b := bytes.Buffer{}
	if err = t.Execute(&b, dat); err != nil {
		return nil, fmt.Errorf("query table template execute: %w", err)
	}
	return &b, nil
}

// queryTables generates the SQL import database and tables statement.
func (f *Flags) queryTables(db *sql.DB) (*bytes.Buffer, error) {
	if db == nil {
		return nil, ErrDB
	}
	var buf1, buf2, buf3 *bytes.Buffer
	var err error
	switch f.Parallel {
	case true:
		buf1, buf2, buf3, err = f.queryTablesWG(db)
		if err != nil {
			return nil, err
		}
	default:
		buf1, buf2, buf3, err = f.queryTablesSeq(db)
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
		},
	}
	tmpl, err := template.New("test").Funcs(tmplFunc()).Parse(templ.Tables)
	if err != nil {
		return nil, fmt.Errorf("query tables template: %w", err)
	}
	b := bytes.Buffer{}
	if err = tmpl.Execute(&b, &data); err != nil {
		return nil, fmt.Errorf("query tables template execute: %w", err)
	}
	return &b, nil
}

func (f *Flags) queryTablesWG(db *sql.DB) (
	*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error,
) {
	if db == nil {
		return nil, nil, nil, ErrDB
	}
	const delta = 3
	wg := sync.WaitGroup{}
	var e1, e2, e3 error
	var buf1, buf2, buf3 *bytes.Buffer
	wg.Add(delta)
	go func(f *Flags) {
		defer wg.Done()
		buf1, e1 = f.reqDB(db, Files)
	}(f)
	go func(f *Flags) {
		defer wg.Done()
		buf2, e2 = f.reqDB(db, Groups)
	}(f)
	go func(f *Flags) {
		defer wg.Done()
		buf3, e3 = f.reqDB(db, Netresources)
	}(f)
	wg.Wait()
	for _, err := range []error{e1, e2, e3} {
		if err != nil {
			return nil, nil, nil, fmt.Errorf("query tables: %w", err)
		}
	}
	return buf1, buf2, buf3, nil
}

func (f *Flags) queryTablesSeq(db *sql.DB) (
	*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error,
) {
	if db == nil {
		return nil, nil, nil, ErrDB
	}
	buf1, err := f.reqDB(db, Files)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("query file table: %w", err)
	}
	buf2, err := f.reqDB(db, Groups)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("query groups table: %w", err)
	}
	buf3, err := f.reqDB(db, Netresources)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("query netresources table: %w", err)
	}
	return buf1, buf2, buf3, nil
}

// reqDB requests an INSERT INTO ? VALUES ? SQL statement for table.
func (f *Flags) reqDB(db *sql.DB, t Table) (*bytes.Buffer, error) {
	if db == nil {
		return nil, ErrDB
	}
	if f.Table.check() != nil {
		return nil, fmt.Errorf("reqdb table: %w", ErrNoTable)
	}
	c := Flags{
		Table:  t,
		Method: Create,
		Limit:  f.Limit,
	}
	buf, err := c.queryDB(db)
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
func (f *Flags) write(w io.Writer, buf *bytes.Buffer) error {
	if buf == nil {
		return fmt.Errorf("buf %w", ErrPointer)
	}
	const bz2 = ".bz2"
	name := path.Join(f.SQLDumps, f.fileName())
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
		fmt.Fprintf(w, "Saved %s to %s\n", humanize.Bytes(uint64(stat.Size())), name)
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
		fmt.Fprintf(w, "Saved %s to %s\n", humanize.Bytes(uint64(n)), name)
		return file.Close()
	default:
		if _, err := io.WriteString(os.Stdout, buf.String()); err != nil {
			return fmt.Errorf("flags write io writer: %w", err)
		}
		return nil
	}
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
func columns(db *sql.DB, t Table) ([]string, error) {
	if db == nil {
		return nil, ErrDB
	}
	query, info := "", ""
	switch t {
	case Files:
		query = "SELECT * FROM files LIMIT 0"
		info = "files"
	case Groups:
		query = "SELECT * FROM groupnames LIMIT 0"
		info = "groupnames"
	case Netresources:
		query = "SELECT * FROM netresources LIMIT 0"
		info = "netresources"
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("columns %s query: %w", info, err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("columns %s query rows: %w", info, rows.Err())
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("columns rows: %w", err)
	}
	return cols, nil
}

// rows returns the values of table.
func rows(db *sql.DB, t Table, limit int) ([]string, error) {
	if db == nil {
		return nil, ErrDB
	}
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
	}
	return values(rows)
}

func allFiles(db *sql.DB) ([]string, error) {
	if db == nil {
		return nil, ErrDB
	}
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
	if db == nil {
		return nil, ErrDB
	}
	rows, err := db.Query("SELECT * FROM groupnames")
	if err != nil {
		return nil, fmt.Errorf("rows groupnames query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows groupnames query rows: %w", rows.Err())
	}
	defer rows.Close()
	return values(rows)
}

func allNetresources(db *sql.DB) ([]string, error) {
	if db == nil {
		return nil, ErrDB
	}
	rows, err := db.Query("SELECT * FROM netresources")
	if err != nil {
		return nil, fmt.Errorf("rows netresources query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows netresources query rows: %w", rows.Err())
	}
	defer rows.Close()
	return values(rows)
}

func limitFiles(limit int, db *sql.DB) ([]string, error) {
	if db == nil {
		return nil, ErrDB
	}
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
	if db == nil {
		return nil, ErrDB
	}
	rows, err := db.Query("SELECT * FROM groupnames LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("rows limit groupnames query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows limit groupnames query rows: %w", rows.Err())
	}
	defer rows.Close()
	return values(rows)
}

func limitNetresources(limit int, db *sql.DB) ([]string, error) {
	if db == nil {
		return nil, ErrDB
	}
	rows, err := db.Query("SELECT * FROM netresources LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("rows limit netresources query: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows limit netresources query rows: %w", rows.Err())
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
	if rows == nil {
		return nil, fmt.Errorf("rows %w", ErrPointer)
	}
	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("rows columns: %w", err)
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("rows column types: %w", err)
	}
	vals := make([]sql.RawBytes, len(cols))
	dest := make([]any, len(vals))
	for i := range vals {
		dest[i] = &vals[i]
	}
	v := []string{}
	for rows.Next() {
		result := make([]string, len(cols))
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
