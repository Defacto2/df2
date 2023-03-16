// Package database interacts with the MySQL datastore of Defacto2.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database/connect"
	"github.com/Defacto2/df2/pkg/database/internal/export"
	"github.com/Defacto2/df2/pkg/database/internal/recd"
	"github.com/Defacto2/df2/pkg/database/internal/update"
	"github.com/Defacto2/df2/pkg/database/msql"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gookit/color"
	"go.uber.org/zap"
)

var (
	ErrColType  = errors.New("the value type is not usable with the mysql column")
	ErrConnect  = errors.New("could not connect to the mysql database server")
	ErrNoID     = errors.New("unique id is does not exist in the database table")
	ErrSynID    = errors.New("id is not a valid id or uuid value")
	ErrSynUUID  = errors.New("id is not a valid uuid")
	ErrNoTable  = errors.New("unknown database table")
	ErrNoMethod = errors.New("unknown database export type")
)

const (
	// Datetime MySQL format.
	Datetime = "2006-01-02T15:04:05Z"

	// UpdateID is a user id to use with the updatedby column.
	UpdateID = "b66dc282-a029-4e99-85db-2cf2892fffcc"

	hide = "****"
	Null = "NULL"

	CountFiles   = "SELECT COUNT(*) FROM `files`"
	CountWaiting = CountFiles + " WHERE `deletedby` IS NULL AND `deletedat` IS NOT NULL"

	SelKeys   = "SELECT `id` FROM `files`"
	SelNames  = "SELECT `filename` FROM `files`"
	SelUpdate = "SELECT `updatedat` FROM `files`" +
		" WHERE `createdat` <> `updatedat` AND `deletedby` IS NULL" +
		" ORDER BY `updatedat` DESC LIMIT 1"

	WhereDownloadBlock = "WHERE `file_security_alert_url` IS NOT NULL AND `file_security_alert_url` != ''"
	WhereAvailable     = "WHERE `deletedat` IS NULL"
	WhereHidden        = "WHERE `deletedat` IS NOT NULL"
)

// Empty is used as a blank value for search maps.
// See: https://dave.cheney.net/2014/03/25/the-empty-struct
type Empty struct{}

// IDs are unique UUID values used by the database and filenames.
type IDs map[string]struct{}

// Flags are command line arguments.
type Flags = export.Flags

// A database table.
type Table int

const (
	Files        Table = iota // Files records.
	Groups                    // Groups names.
	Netresources              // Netresources for online websites.
	Users                     // Users are site logins.
)

func (t Table) String() string {
	return [...]string{"files", "groupnames", "netresources", "users"}[t]
}

// Init initialises the database connection using stored settings.
func Init() connect.Connection {
	return connect.Init()
}

// Connect will connect to the database and handle any errors.
func Connect(cfg configger.Config) (*sql.DB, error) {
	// In the future this could use either psql or msql.
	return msql.Connect(cfg)
}

// ConnErr will connect to the database or return any errors.
func ConnErr() (*sql.DB, error) {
	c := connect.Init()
	db, err := sql.Open("mysql", c.String())
	if err != nil {
		e := strings.Replace(err.Error(), c.Pass, hide, 1)
		return nil, fmt.Errorf("mysql open error: %s: %w", e, ErrConnect)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql ping error: %w", err)
	}
	return db, nil
}

// ConnInfo will connect to the database and return any errors.
func ConnInfo() string {
	c := connect.Init()
	db, err := sql.Open("mysql", c.String())
	defer func() {
		db.Close()
	}()
	if err != nil {
		return strings.Replace(err.Error(), c.Pass, hide, 1)
	}
	err = db.Ping()
	if err == nil {
		return ""
	}
	me := &mysql.MySQLError{}
	if ok := errors.As(err, &me); ok {
		e := strings.Replace(err.Error(), c.User, color.Primary.Sprint(c.User), 1)
		return fmt.Sprintf("%s %v", color.Info.Sprint("MySQL"), color.Danger.Sprint(e))
	}
	nop := &net.OpError{}
	if ok := errors.As(err, &nop); ok {
		if strings.Contains(err.Error(), "connect: connection refused") {
			return fmt.Sprintf("%s '%v' %s",
				color.Danger.Sprint("database server"),
				color.Primary.Sprint(c.Server),
				color.Danger.Sprint("is either down or the port is blocked"))
		}
		return color.Danger.Sprint(err)
	}
	return ""
}

// Approve automatically checks and clears file records for live.
func Approve(db *sql.DB, w io.Writer, l *zap.SugaredLogger, incoming string, verbose bool) error {
	return recd.Queries(db, w, l, incoming, verbose)
}

// CheckID reports an error message for an incorrect universal unique record id or MySQL auto-generated id.
func CheckID(s string) error {
	if !IsUUID(s) && !IsID(s) {
		return fmt.Errorf("invalid id, it needs to be an auto-generated MySQL id or an uuid: %w", ErrSynID)
	}
	return nil
}

// CheckUUID reports an error message for an incorrect universal unique record id.
func CheckUUID(s string) error {
	if !IsUUID(s) {
		const example = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
		return fmt.Errorf("invalid uuid %q, it requires RFC 4122 syntax %s: %w", s, example, ErrSynUUID)
	}
	return nil
}

// ColTypes details the columns used by the table.
func ColTypes(db *sql.DB, w io.Writer, t Table) (string, error) {
	// LIMIT 0 quickly returns an empty set
	var query string
	switch t {
	case Files:
		query = fmt.Sprintf("SELECT * FROM %s LIMIT 0", Files)
	case Groups:
		query = fmt.Sprintf("SELECT * FROM %s LIMIT 0", Groups)
	case Netresources:
		query = fmt.Sprintf("SELECT * FROM %s LIMIT 0", Netresources)
	case Users:
		query = fmt.Sprintf("SELECT * FROM %s LIMIT 0", Users)
	}
	rows, err := db.Query(query)
	if err != nil {
		return "", fmt.Errorf("column types query: %w", err)
	}
	if rows.Err() != nil {
		return "", fmt.Errorf("column types rows: %w", rows.Err())
	}
	defer rows.Close()
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return "", fmt.Errorf("column types: %w", err)
	}
	const padding = 3
	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 0, 0, padding, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "Column name\tType\tNullable\tLength\t")
	for _, s := range colTypes {
		fmt.Fprintf(tw, "%v\t%v\t%v\t%v\t\n",
			s.Name(), s.DatabaseTypeName(), nullable(s), recd.ColLen(s))
	}
	if err = tw.Flush(); err != nil {
		return "", fmt.Errorf("column types flush tab writer: %w", err)
	}
	return buf.String(), nil
}

func nullable(s *sql.ColumnType) string {
	n, ok := s.Nullable()
	if !ok {
		return ""
	}
	if n {
		return "✓"
	}
	return "✗"
}

// DateTime colours and formats a date and time string.
func DateTime(l *zap.SugaredLogger, raw sql.RawBytes) string {
	v := string(raw)
	if v == "" {
		return ""
	}
	t, err := time.Parse(Datetime, v)
	if err != nil {
		l.Errorln(err)
		return "?"
	}
	if t.UTC().Format("01 2006") != time.Now().Format("01 2006") {
		return fmt.Sprintf("%v ", color.Info.Sprint(t.UTC().Format("02 Jan 2006 ")))
	}
	return fmt.Sprintf("%v ", color.Info.Sprint(t.UTC().Format("02 Jan 15:04")))
}

// Distinct returns a unique list of values from the table column.
func Distinct(db *sql.DB, w io.Writer, value string) ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT ? FROM `files`", value)
	if err != nil {
		return nil, fmt.Errorf("distinct query: %w", err)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("distinct rows: %w", rows.Err())
	}
	defer rows.Close()
	res := []string{}
	var dest sql.NullString
	for rows.Next() {
		if err = rows.Scan(&dest); err != nil {
			return nil, fmt.Errorf("distinct scan: %w", err)
		}
		res = append(res, strings.ToLower(dest.String))
	}
	return res, nil
}

// FileUpdate reports if the named file is newer than the database time.
// True is always returned when the named file does not exist.
func FileUpdate(name string, db time.Time) (bool, error) {
	f, err := os.Stat(name)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("file update stat %q: %w", name, err)
	}
	return !f.ModTime().UTC().After(db.UTC()), nil
}

type Update update.Update

// Execute Query and Args to update the database and returns the total number of changes.
func Execute(db *sql.DB, u Update) (int64, error) {
	return update.Update(u).Execute(db)
}

// Fix any malformed section and platforms found in the database.
func Fix(db *sql.DB, w io.Writer, l *zap.SugaredLogger) error {
	start := time.Now()
	if err := update.Filename.NamedTitles(db, w); err != nil {
		return fmt.Errorf("update filenames: %w", err)
	}
	if err := update.GroupFor.NamedTitles(db, w); err != nil {
		return fmt.Errorf("update groups for: %w", err)
	}
	if err := update.GroupBy.NamedTitles(db, w); err != nil {
		return fmt.Errorf("update groups by: %w", err)
	}
	elapsed := time.Since(start).Seconds()
	fmt.Fprintf(w, ", time taken %.1f seconds\n", elapsed)

	dist, err := update.Distinct(db, "section")
	if err != nil {
		return fmt.Errorf("fix distinct section: %w", err)
	}
	update.Sections(db, w, l, &dist)
	dist, err = update.Distinct(db, "platform")
	if err != nil {
		return fmt.Errorf("fix distinct platform: %w", err)
	}
	update.Platforms(db, w, l, &dist)
	return nil
}

// DemozooID finds a record ID by the saved Demozoo production ID. If no production exists a zero is returned.
func DemozooID(db *sql.DB, id uint) (uint, error) {
	var dz uint
	// https://stackoverflow.com/questions/1676551/best-way-to-test-if-a-row-exists-in-a-mysql-table
	if err := db.QueryRow("SELECT id FROM files WHERE web_id_demozoo=?", id).Scan(&dz); err != nil {
		return 0, fmt.Errorf("demozoo exist query row: %w", err)
	}
	return dz, nil
}

// GetID returns a numeric Id from a UUID or database id s value.
func GetID(db *sql.DB, s string) (uint, error) {
	// auto increment numeric ids
	var id uint
	if v, err := strconv.Atoi(s); err == nil {
		// https://stackoverflow.com/questions/1676551/best-way-to-test-if-a-row-exists-in-a-mysql-table
		if err = db.QueryRow("SELECT EXISTS(SELECT * FROM files WHERE id=?)", v).Scan(&id); err != nil {
			return 0, fmt.Errorf("lookupid query row: %w", err)
		}
		if id == 0 {
			return 0, fmt.Errorf("lookupid %q: %w", s, ErrNoID)
		}
		return uint(v), nil
	}
	// uuid ids
	s = strings.ToLower(s)
	if err := db.QueryRow("SELECT id FROM files WHERE uuid=?", s).Scan(&id); err != nil {
		return 0, fmt.Errorf("lookupid %q: %w", s, ErrNoID)
	}
	return id, nil
}

// GetKeys returns all the primary keys used by the files table.
// The integer keys are sorted incrementally.
// An SQL WHERE statement can be provided to filter the results.
func GetKeys(db *sql.DB, whereStmt string) ([]int, error) {
	whereStmt = strings.TrimSpace(whereStmt)
	cs := CountFiles
	if whereStmt != "" {
		cs = fmt.Sprintf("%s %s", cs, whereStmt)
	}
	var count int
	if err := db.QueryRow(cs).Scan(&count); err != nil {
		return nil, err
	}

	ss := SelKeys
	if whereStmt != "" {
		ss = fmt.Sprintf("%s %s", ss, whereStmt)
	}
	rows, err := db.Query(ss)
	if err != nil {
		return nil, fmt.Errorf("getkeys query: %w", err)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("getkeys rows: %w", rows.Err())
	}
	defer rows.Close()

	id, i := "", -1
	keys := make([]int, 0, count)
	for rows.Next() {
		i++
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("getkeys scan: %w", err)
		}
		val, err := strconv.Atoi(id)
		if err != nil {
			continue
		}
		keys = append(keys, val)
	}
	return keys, nil
}

// GetFile returns the filename from a supplied UUID or database ID value.
func GetFile(db *sql.DB, s string) (string, error) {
	var n sql.NullString
	if v, err := strconv.Atoi(s); err == nil {
		err = db.QueryRow(SelNames+" WHERE id=?", v).Scan(&n)
		if err != nil {
			return "", fmt.Errorf("lookup file by id queryrow %q: %w", s, err)
		}
		return n.String, nil
	}
	s = strings.ToLower(s)
	err := db.QueryRow(SelNames+" WHERE uuid=?", s).Scan(&n)
	if err != nil {
		return "", fmt.Errorf("lookup file by uuid queryrow %q: %w", s, err)
	}
	return n.String, nil
}

// IsDemozoo reports if a fetched demozoo file record is set to unapproved.
func IsDemozoo(b []sql.RawBytes) (bool, error) {
	return recd.IsDemozoo(b)
}

// IsID reports whether string is a autogenerated record id.
func IsID(s string) bool {
	r := regexp.MustCompile(`^0+`)
	if x := r.ReplaceAllString(s, ""); x != s {
		return false // leading zeros found
	}
	if i, err := strconv.Atoi(s); err != nil {
		return false // not a number
	} else if f := float64(i); f != math.Abs(f) {
		return false // not an absolute value
	}
	return true
}

// IsProof reports if a fetched proof file record is set to unapproved.
func IsProof(l *zap.SugaredLogger, b []sql.RawBytes) (bool, error) {
	// SQL column names can be found in the sqlSelect() func in proof.go
	const deletedat, updatedat = 2, 6
	if len(b) < updatedat {
		return false, nil
	}
	if b[deletedat] == nil {
		return false, nil
	}
	n, err := recd.Valid(b[deletedat], b[updatedat])
	if err != nil {
		return false, err
	}
	return n, nil
}

// IsUUID reports whether string is a universal unique record id.
func IsUUID(s string) bool {
	if _, err := uuid.Parse(s); err != nil {
		return false
	}
	return true
}

// LastUpdate reports the time when the files database was last modified.
func LastUpdate(db *sql.DB) (time.Time, error) {
	row := db.QueryRow(SelUpdate)
	t := time.Time{}
	if err := row.Scan(&t); err != nil {
		return t, fmt.Errorf("last update: %w", err)
	}
	return t, nil
}

// ObfuscateParam hides the param value using the method implemented in CFWheels obfuscateParam() helper.
func ObfuscateParam(param string) string {
	if param == "" {
		return ""
	}
	// check to make sure param doesn't begin with a 0 digit
	if param[0] == '0' {
		return param
	}
	pint, err := strconv.Atoi(param)
	if err != nil {
		return param
	}
	l := len(param)
	r, err := recd.ReverseInt(uint(pint))
	if err != nil {
		return param
	}
	afloat64 := math.Pow10(l) + float64(r)
	// keep a and b as int type
	a, b := int(afloat64), 0
	for i := 1; i <= l; i++ {
		// slice individual digits from param and sum them
		s, err := strconv.Atoi(string(param[l-i]))
		if err != nil {
			return param
		}
		b += s
	}
	// base 64 conversion
	const hex, xor, sum = 16, 461, 154
	a ^= xor
	b += sum
	return strconv.FormatInt(int64(b), hex) + strconv.FormatInt(int64(a), hex)
}

// StripChars removes incompatible characters used for groups and author names.
func StripChars(s string) string {
	r := regexp.MustCompile(`[^A-Za-zÀ-ÖØ-öø-ÿ0-9\-,& ]`)
	return r.ReplaceAllString(s, "")
}

// StripStart removes non-alphanumeric characters from the start of the string.
func StripStart(s string) string {
	r := regexp.MustCompile(`[A-Za-z0-9À-ÖØ-öø-ÿ]`)
	f := r.FindStringIndex(s)
	if f == nil {
		return ""
	}
	if f[0] != 0 {
		return s[f[0]:]
	}
	return s
}

// Tbls are the available tables in the database.
func Tbls() string {
	return export.Tbls()
}

// Total reports the number of records fetched by the supplied SQL query.
func Total(db *sql.DB, w io.Writer, s *string) (int, error) {
	rows, err := db.Query(*s)
	switch {
	case err != nil && strings.Contains(err.Error(), "SQL syntax"):
		fmt.Fprintln(w, *s)
		return -1, err
	case err != nil:
		return -1, fmt.Errorf("total query: %w", err)
	case rows.Err() != nil:
		return -1, fmt.Errorf("total query rows: %w", rows.Err())
	}
	defer rows.Close()
	sum := 0
	for rows.Next() {
		sum++
	}
	return sum, nil
}

// TrimSP removes duplicate spaces from a string.
func TrimSP(s string) string {
	r := regexp.MustCompile(`\s+`)
	return r.ReplaceAllString(s, " ")
}

// Val returns the column value as either a string or "NULL".
func Val(col sql.RawBytes) string {
	if col == nil {
		return Null
	}
	return string(col)
}

// Waiting returns the number of files requiring approval for public display.
func Waiting(db *sql.DB) (uint, error) {
	var count uint
	if err := db.QueryRow(CountWaiting).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// DeObfuscate a public facing, obfuscated file ID or file URL.
// A URL can point to a Defacto2 file download or detail page.
func DeObfuscate(s string) int {
	p := strings.Split(s, "?")
	d := deObfuscate(path.Base(p[0]))
	id, _ := strconv.Atoi(d)
	return id
}

// deObfuscate de-obfuscates a CFWheels obfuscateParam or Obfuscate() obfuscated string.
func deObfuscate(s string) string {
	const twoChrs, decimal, hexadecimal = 2, 10, 16
	// CFML source:
	// https://github.com/cfwheels/cfwheels/blob/cf8e6da4b9a216b642862e7205345dd5fca34b54/wheels/global/misc.cfm
	if _, err := strconv.Atoi(s); err == nil || len(s) < twoChrs {
		return s
	}
	// De-obfuscate string.
	tail := s[twoChrs:]
	n, err := strconv.ParseInt(tail, hexadecimal, 0)
	if err != nil {
		return s
	}

	n ^= 461 // bitxor
	ns := strconv.Itoa(int(n))
	l := len(ns) - 1
	tail = ""

	for i := 0; i < l; i++ {
		f := ns[l-i:][:1]
		tail += f
	}
	// Create checks.
	ct := 0
	l = len(tail)

	for i := 0; i < l; i++ {
		chr := tail[i : i+1]
		n, err1 := strconv.Atoi(chr)

		if err1 != nil {
			return s
		}

		ct += n
	}
	// Run checks.
	ci, err := strconv.ParseInt(s[:2], hexadecimal, 0)
	if err != nil {
		return s
	}

	c2 := strconv.FormatInt(ci, decimal)

	const unknown = 154

	if strconv.FormatInt(int64(ct+unknown), decimal) != c2 {
		return s
	}

	return tail
}
