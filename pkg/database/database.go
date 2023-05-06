// Package database interacts with the remote datastore for Defacto2. Currently
// MySQL is implemented with Postgres to be added later.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database/internal/export"
	"github.com/Defacto2/df2/pkg/database/internal/recd"
	"github.com/Defacto2/df2/pkg/database/internal/templ"
	"github.com/Defacto2/df2/pkg/database/internal/update"
	"github.com/Defacto2/df2/pkg/database/msql"
	"github.com/google/uuid"
	"github.com/gookit/color"
)

var (
	ErrDB      = errors.New("database handle pointer cannot be nil")
	ErrNoID    = errors.New("unique id is does not exist in the database table")
	ErrPointer = errors.New("pointer value cannot be nil")
	ErrSynID   = errors.New("id value is not a valid")
	ErrValue   = errors.New("argument cannot be an empty value")
)

const (
	Datetime = "2006-01-02T15:04:05Z" // Datetime MySQL format.

	// ExampleID is an invalid placeholder UUID, where x represents a digit.
	ExampleID = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	// TestID is a generic UUID that can be used for unit tests.
	TestID = "00000000-0000-0000-0000-000000000000"
	// UpdateID is a user id to use with the updatedby column.
	UpdateID = "b66dc282-a029-4e99-85db-2cf2892fffcc"

	WhereAvailable     = templ.WhereAvailable
	WhereDownloadBlock = templ.WhereDownloadBlock
	WhereHidden        = templ.WhereHidden
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
)

func (t Table) String() string {
	if t > Netresources {
		return ""
	}
	return [...]string{"files", "groupnames", "netresources"}[t]
}

type Update update.Update

// Connect the database and handle any errors.
// The DB connection must be closed after use.
func Connect(cfg conf.Config) (*sql.DB, error) {
	// In the future this could use either psql or msql.
	return msql.Connect(cfg)
}

// ConnDebug will connect to the database and return any errors.
func ConnDebug(cfg conf.Config) (string, error) {
	return msql.ConnDebug(cfg)
}

// Approve automatically checks and clears file records for live.
func Approve(db *sql.DB, w io.Writer, cfg conf.Config, verbose bool) error {
	if db == nil {
		return ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	return recd.Queries(db, w, cfg, verbose)
}

// CheckID checks the syntax of the universal unique record id or MySQL auto-generated id.
func CheckID(id string) error {
	if !IsUUID(id) && !IsID(id) {
		return fmt.Errorf("invalid id, it needs to be an auto-generated MySQL id or an uuid: %w",
			ErrSynID)
	}
	return nil
}

// CheckUUID checks the syntax of the universal unique record id.
func CheckUUID(uuid string) error {
	if !IsUUID(uuid) {
		return fmt.Errorf("invalid uuid %q, it requires RFC 4122 syntax %s: %w",
			uuid, ExampleID, ErrSynID)
	}
	return nil
}

// Columns details the columns used by the table.
func Columns(db *sql.DB, w io.Writer, t Table) error {
	if db == nil {
		return ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	// LIMIT 0 quickly returns an empty set
	query := ""
	switch t {
	case Files:
		query = fmt.Sprintf("SELECT * FROM %s LIMIT 0", Files)
	case Groups:
		query = fmt.Sprintf("SELECT * FROM %s LIMIT 0", Groups)
	case Netresources:
		query = fmt.Sprintf("SELECT * FROM %s LIMIT 0", Netresources)
	}
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("column types query: %w", err)
	}
	if rows.Err() != nil {
		return fmt.Errorf("column types rows: %w", rows.Err())
	}
	defer rows.Close()
	types, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf("column types: %w", err)
	}
	const padding = 3
	buf := strings.Builder{}
	tw := tabwriter.NewWriter(&buf, 0, 0, padding, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "Column name\tType\tNullable\tLength\t")
	for _, s := range types {
		fmt.Fprintf(tw, "%v\t%v\t%v\t%v\t\n",
			s.Name(), s.DatabaseTypeName(), nullable(s), recd.ColLen(s))
	}
	if err = tw.Flush(); err != nil {
		return fmt.Errorf("column types flush tab writer: %w", err)
	}
	fmt.Fprintln(w, buf)
	return nil
}

func nullable(s *sql.ColumnType) string {
	if s == nil {
		return "error"
	}
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
func DateTime(raw sql.RawBytes) (string, error) {
	v := string(raw)
	if v == "" {
		return "", nil
	}
	t, err := time.Parse(Datetime, v)
	if err != nil {
		return "?", err
	}
	if t.UTC().Format("01 2006") != time.Now().Format("01 2006") {
		return color.Info.Sprint(t.UTC().Format("02 Jan 2006")), nil
	}
	return color.Info.Sprint(t.UTC().Format("02 Jan 15:04")), nil
}

// Distinct returns a unique list of values from the table column.
func Distinct(db *sql.DB, value string) ([]string, error) {
	if db == nil {
		return nil, ErrDB
	}
	if value == "" {
		return nil, nil
	}
	rows, err := db.Query("SELECT DISTINCT ? FROM `files`", value)
	if err != nil {
		return nil, fmt.Errorf("distinct query: %w", err)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("distinct rows: %w", rows.Err())
	}
	defer rows.Close()
	res := []string{}
	dest := sql.NullString{}
	for rows.Next() {
		if err = rows.Scan(&dest); err != nil {
			return nil, fmt.Errorf("distinct scan: %w", err)
		}
		res = append(res, strings.ToLower(dest.String))
	}
	return res, nil
}

// FileUpdate returns true when named file is newer than the database time.
// True is always returned when the named file does not exist or
// whenever it is 0 bytes in size.
func FileUpdate(name string, db time.Time) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("name %w", ErrValue)
	}
	f, err := os.Stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("file update stat %q: %w", name, err)
	}
	if f.Size() == 0 {
		return true, nil
	}
	return f.ModTime().UTC().After(db.UTC()), nil
}

// Execute Query and Args to update the database and returns the total number of changes.
func Execute(db *sql.DB, u Update) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	return update.Update(u).Execute(db)
}

// Fix any malformed section and platforms found in the database.
func Fix(db *sql.DB, w io.Writer) error {
	if db == nil {
		return ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	if err := update.Filename.NamedTitles(db, w); err != nil {
		return fmt.Errorf("update filenames: %w", err)
	}
	if err := update.GroupFor.NamedTitles(db, w); err != nil {
		return fmt.Errorf("update groups for: %w", err)
	}
	if err := update.GroupBy.NamedTitles(db, w); err != nil {
		return fmt.Errorf("update groups by: %w", err)
	}
	dist, err := update.Distinct(db, "section")
	if err != nil {
		return fmt.Errorf("fix distinct section: %w", err)
	}
	if err := update.Sections(db, w, &dist); err != nil {
		return fmt.Errorf("update sections: %w", err)
	}
	dist, err = update.Distinct(db, "platform")
	if err != nil {
		return fmt.Errorf("fix distinct platform: %w", err)
	}
	if err = update.Platforms(db, w, &dist); err != nil {
		return fmt.Errorf("update platforms: %w", err)
	}
	return nil
}

// DemozooID looks up a Demozoo productions ID in the files table,
// and returns the ID of the first matched Defacto2 file record.
// If no match is found then a zero is returned.
func DemozooID(db *sql.DB, demozoo uint) (int, error) {
	if db == nil {
		return 0, ErrDB
	}
	dz := int(0)
	if err := db.QueryRow("SELECT id FROM files WHERE web_id_demozoo=?", demozoo).Scan(&dz); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("demozooid %w", err)
	}
	return dz, nil
}

// GetID looks up the val and returns a matching auto-increment ID of the file record.
// The val string must be either a UUID of the record or an increment ID.
func GetID(db *sql.DB, val string) (int, error) {
	if db == nil {
		return 0, ErrDB
	}
	id := 0
	// numeric ids
	if v, err := strconv.Atoi(val); err == nil {
		// https://stackoverflow.com/questions/1676551/best-way-to-test-if-a-row-exists-in-a-mysql-table
		if err = db.QueryRow("SELECT EXISTS(SELECT * FROM files WHERE id=?)", v).Scan(&id); err != nil {
			return 0, fmt.Errorf("get id query row: %w", err)
		}
		if id == 0 {
			return 0, fmt.Errorf("get id %d: %w", v, ErrNoID)
		}
		return v, nil
	}
	// uuid ids
	v := strings.ToLower(val)
	if err := db.QueryRow("SELECT id FROM files WHERE uuid=?", v).Scan(&id); err != nil {
		return 0, fmt.Errorf("get id %s: %w", v, ErrNoID)
	}
	return id, nil
}

// GetKeys returns all the primary keys used by the files table.
// The integer keys are sorted incrementally.
// An optional, statement can be provided to filter the results.
func GetKeys(db *sql.DB, stmt string) ([]int, error) {
	if db == nil {
		return nil, ErrDB
	}
	stmt = strings.TrimSpace(stmt)
	query := templ.CountFiles
	if stmt != "" {
		query = fmt.Sprintf("%s %s", query, stmt)
	}
	count := 0
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return nil, err
	}
	queryKeys := templ.SelKeys
	if stmt != "" {
		queryKeys = fmt.Sprintf("%s %s", queryKeys, stmt)
	}
	rows, err := db.Query(queryKeys)
	if err != nil {
		return nil, fmt.Errorf("get keys query: %w", err)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("get keys rows: %w", rows.Err())
	}
	defer rows.Close()

	id := ""
	keys := []int{}
	for rows.Next() {
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("get keys scan: %w", err)
		}
		val, err := strconv.Atoi(id)
		if err != nil {
			continue
		}
		keys = append(keys, val)
	}
	return keys, nil
}

// GetFile looks up val and returns the filename of the file record.
// The val string must be either a UUID of the record or an increment ID.
func GetFile(db *sql.DB, val string) (string, error) {
	if db == nil {
		return "", ErrDB
	}
	n := sql.NullString{}
	if v, err := strconv.Atoi(val); err == nil {
		err = db.QueryRow(templ.SelNames+" WHERE id=?", v).Scan(&n) //nolint:execinquery
		if err != nil {
			return "", fmt.Errorf("lookup file by id queryrow %q: %w", val, err)
		}
		return n.String, nil
	}
	val = strings.ToLower(val)
	err := db.QueryRow(templ.SelNames+" WHERE uuid=?", val).Scan(&n) //nolint:execinquery
	if err != nil {
		return "", fmt.Errorf("lookup file by uuid queryrow %q: %w", val, err)
	}
	return n.String, nil
}

// IsDemozoo reports if a fetched demozoo file record is set to unapproved.
func IsDemozoo(b []sql.RawBytes) (bool, error) {
	return recd.IsDemozoo(b)
}

// IsID reports whether string is an auto-generated record id.
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

// IsUnApproved reports if a fetched file record is set to unapproved.
func IsUnApproved(b []sql.RawBytes) (bool, error) {
	// SQL column names can be found in the sqlSelect() func in proof.go
	const deletedat, updatedat = 2, 6
	if len(b) <= updatedat {
		return false, nil
	}
	if b[deletedat] == nil || b[updatedat] == nil {
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
	if db == nil {
		return time.Time{}, ErrDB
	}
	row := db.QueryRow(templ.SelUpdate)
	t := time.Time{}
	if err := row.Scan(&t); err != nil {
		return t, fmt.Errorf("last update: %w", err)
	}
	return t, nil
}

// ObfuscateParam hides the param value using the method implemented in
// CFWheels obfuscateParam() helper.
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
	r, err := recd.ReverseInt(pint)
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
	if db == nil {
		return -1, ErrDB
	}
	if s == nil {
		return -1, fmt.Errorf("s %w", ErrPointer)
	}
	if w == nil {
		w = io.Discard
	}
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
		return "NULL"
	}
	return string(col)
}

// Waiting returns the number of files requiring approval for public display.
func Waiting(db *sql.DB) (int, error) {
	if db == nil {
		return -1, ErrDB
	}
	count := -1
	if err := db.QueryRow(templ.CountWaiting).Scan(&count); err != nil {
		return -1, err
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
