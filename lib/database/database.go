// Package database interacts with the MySQL 5.7 datastore of Defacto2.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Defacto2/df2/lib/database/internal/my57"
	"github.com/Defacto2/df2/lib/database/internal/recd"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gookit/color"
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
	// Datetime MySQL 5.7 format.
	Datetime = "2006-01-02T15:04:05Z"

	// UpdateID is a user id to use with the updatedby column.
	UpdateID = "b66dc282-a029-4e99-85db-2cf2892fffcc"

	hide = "****"
	null = "NULL"
)

// Empty is used as a blank value for search maps.
// See: https://dave.cheney.net/2014/03/25/the-empty-struct
type Empty struct{}

// IDs are unique UUID values used by the database and filenames.
type IDs map[string]struct{}

// Init initialises the database connection using stored settings.
func Init() my57.Connection {
	return my57.Init()
}

// Connect will connect to the database and handle any errors.
func Connect() *sql.DB {
	return my57.Connect()
}

// ConnErr will connect to the database or return any errors.
func ConnErr() (*sql.DB, error) {
	c := my57.Init()
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
	c := my57.Init()
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
func Approve(verbose bool) error {
	return recd.Queries(verbose)
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
func ColTypes(table string) error {
	db := my57.Connect()
	defer db.Close()
	// LIMIT 0 quickly returns an empty set
	rows, err := db.Query("SELECT * FROM ? LIMIT 0", table)
	if err != nil {
		return fmt.Errorf("column types query: %w", err)
	}
	if rows.Err() != nil {
		return fmt.Errorf("column types rows: %w", rows.Err())
	}
	defer rows.Close()
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf("column types: %w", err)
	}
	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Column name\tType\tNullable\tLength\t")
	for _, s := range colTypes {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t\n",
			s.Name(), s.DatabaseTypeName(), nullable(s), recd.ColLen(s))
	}
	if err = w.Flush(); err != nil {
		return fmt.Errorf("column types flush tab writer: %w", err)
	}
	return nil
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
func DateTime(raw sql.RawBytes) string {
	v := string(raw)
	if v == "" {
		return ""
	}
	t, err := time.Parse(Datetime, v)
	if err != nil {
		logs.Log(err)
		return "?"
	}
	if t.UTC().Format("01 2006") != time.Now().Format("01 2006") {
		return fmt.Sprintf("%v ", color.Info.Sprint(t.UTC().Format("02 Jan 2006 ")))
	}
	return fmt.Sprintf("%v ", color.Info.Sprint(t.UTC().Format("02 Jan 15:04")))
}

// FileUpdate reports if the file is newer than the database time.
func FileUpdate(name string, database time.Time) (bool, error) {
	f, err := os.Stat(name)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("file update stat %q: %w", name, err)
	}
	return !f.ModTime().UTC().After(database.UTC()), nil
}

// GetID returns a numeric Id from a UUID or database id s value.
func GetID(s string) (uint, error) {
	db := my57.Connect()
	defer db.Close()
	// auto increment numeric ids
	var id uint
	if v, err := strconv.Atoi(s); err == nil {
		// https://stackoverflow.com/questions/1676551/best-way-to-test-if-a-row-exists-in-a-mysql-table
		if err = db.QueryRow("SELECT EXISTS(SELECT * FROM files WHERE id=?)", v).Scan(&id); err != nil {
			return 0, fmt.Errorf("lookupid query row: %w", err)
		}
		if id == 0 {
			return 0, fmt.Errorf("lookupid %q: %w", v, ErrNoID)
		}
		return uint(v), nil
	}
	// uuid ids
	s = strings.ToLower(s)
	if err := db.QueryRow("SELECT id FROM files WHERE uuid=?", s).Scan(&id); err != nil {
		return 0, fmt.Errorf("lookupid %q: %w", s, ErrNoID)
	}
	return id, db.Close()
}

// GetFile returns the filename from a supplied UUID or database ID value.
func GetFile(s string) (string, error) {
	db := my57.Connect()
	defer db.Close()
	var n sql.NullString
	if v, err := strconv.Atoi(s); err == nil {
		err = db.QueryRow("SELECT filename FROM files WHERE id=?", v).Scan(&n)
		if err != nil {
			return "", fmt.Errorf("lookup file by id queryrow %q: %w", s, err)
		}
		return n.String, nil
	}
	s = strings.ToLower(s)
	err := db.QueryRow("SELECT filename FROM files WHERE uuid=?", s).Scan(&n)
	if err != nil {
		return "", fmt.Errorf("lookup file by uuid queryrow %q: %w", s, err)
	}
	return n.String, nil
}

// IsDemozoo reports if a fetched demozoo file record is set to unapproved.
func IsDemozoo(b []sql.RawBytes) bool {
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
func IsProof(b []sql.RawBytes) bool {
	// SQL column names can be found in the sqlSelect() func in proof.go
	const deletedat, updatedat = 2, 6
	if len(b) < updatedat {
		return false
	}
	if b[deletedat] == nil {
		return false
	}
	n, err := recd.Valid(b[deletedat], b[updatedat])
	if err != nil {
		logs.Log(err)
	}
	return n
}

// IsUUID reports whether string is a universal unique record id.
func IsUUID(s string) bool {
	if _, err := uuid.Parse(s); err != nil {
		return false
	}
	return true
}

// LastUpdate reports the time when the files database was last modified.
func LastUpdate() (time.Time, error) {
	db := my57.Connect()
	defer db.Close()
	row := db.QueryRow("SELECT `updatedat` FROM `files` WHERE `deletedat` <> `updatedat`" +
		" ORDER BY `updatedat` DESC LIMIT 1")
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

// Total reports the number of records fetched by the supplied SQL query.
func Total(s *string) (int, error) {
	db := my57.Connect()
	defer db.Close()
	rows, err := db.Query(*s)
	switch {
	case err != nil && strings.Contains(err.Error(), "SQL syntax"):
		logs.Println(s)
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
	return sum, db.Close()
}

// Val returns the column value as either a string or "NULL".
func Val(col sql.RawBytes) string {
	if col == nil {
		return null
	}
	return string(col)
}

// Waiting returns the number of files requiring approval for public display.
func Waiting() (uint, error) {
	const countWaiting = "SELECT COUNT(*)\nFROM `files`\n" +
		"WHERE `deletedby` IS NULL AND `deletedat` IS NOT NULL"
	var count uint
	db := my57.Connect()
	defer db.Close()
	if err := db.QueryRow(countWaiting).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
