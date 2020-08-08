package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gookit/color"
	"github.com/spf13/viper"
)

// UpdateID is a user id to use with the updatedby column.
const UpdateID string = "b66dc282-a029-4e99-85db-2cf2892fffcc"

// Datetime MySQL 5.7 format.
const Datetime string = "2006-01-02T15:04:05Z"

const (
	changeme = "changeme"
	z7       = ".7z"
	arj      = ".arj"
	bz2      = ".bz2"
	png      = ".png"
	rar      = ".rar"
	zip      = ".zip"
)

// Connection information for a MySQL database.
type Connection struct {
	Name     string // database name
	User     string // access username
	Pass     string // access password
	Server   string // host server protocol, address and port
	Protocol string
	Address  string
	Port     string
}

func (c *Connection) String() string {
	return fmt.Sprintf("%v:%v@%v/%v?timeout=5s&parseTime=true", c.User, c.Pass, c.Server, c.Name)
}

var (
	ErrColType  = errors.New("the value type is not usable with the mysql column")
	ErrConnect  = errors.New("could not connect to the mysql database server")
	ErrNoID     = errors.New("unique id is does not exist in the database table")
	ErrSynID    = errors.New("id is not a valid id or uuid value")
	ErrSynUUID  = errors.New("id is not a valid uuid")
	ErrNoTable  = errors.New("unknown database table")
	ErrNoMethod = errors.New("unknown database export type")
)

// Empty is used as a blank value for search maps.
// See: https://dave.cheney.net/2014/03/25/the-empty-struct
type Empty struct{}

// IDs are unique UUID values used by the database and filenames.
type IDs map[string]struct{}

// Init initializes the database connection using stored settings.
func Init() Connection {
	// load config from file or use defaults
	if viper.GetString("connection.name") == "" {
		if err := viper.ReadInConfig(); err != nil {
			defaults()
		}
	}
	return Connection{
		Name:     viper.GetString("connection.name"),
		User:     viper.GetString("connection.user"),
		Pass:     viper.GetString("connection.password"),
		Protocol: viper.GetString("connection.server.protocol"),
		Address:  viper.GetString("connection.server.host"),
		Port:     viper.GetString("connection.server.port"),
		Server: fmt.Sprintf("%v(%v:%v)", // example: tcp(localhost:3306)
			viper.GetString("connection.server.protocol"),
			viper.GetString("connection.server.host"),
			viper.GetString("connection.server.port")),
	}
}

// Connect will connect to the database and handle any errors.
func Connect() *sql.DB {
	c := Init()
	db, err := sql.Open("mysql", c.String())
	if err != nil {
		e := strings.Replace(err.Error(), c.Pass, "****", 1)
		log.Fatal(fmt.Errorf("connect database open: %s: %w", e, ErrConnect))
	}
	// ping the server to make sure the connection works
	if err = db.Ping(); err != nil {
		logs.Println(color.Secondary.Sprint(strings.Replace(c.String(), c.Pass, "****", 1)))
		// filter the password and then print the datasource connection info
		// to discover more errors fmt.Printf("%T", err)
		if err, _ := err.(*mysql.MySQLError); err != nil {
			log.Fatal(fmt.Errorf("connect database mysql: %w", err))
		}
		if err, _ := err.(*net.OpError); err != nil {
			if strings.Contains(err.Error(), "connect: connection refused") {
				log.Fatal(fmt.Errorf("database server %v is either down or the %v %v port is blocked: %w",
					c.Address, c.Protocol, c.Port, ErrConnect))
			}
		}
		log.Fatal(fmt.Errorf("connect database: %w", err))
	}
	return db
}

// ConnectErr will connect to the database or return any errors.
func ConnectErr() (db *sql.DB, err error) {
	c := Init()
	db, err = sql.Open("mysql", c.String())
	if err != nil {
		e := strings.Replace(err.Error(), c.Pass, "****", 1)
		return nil, fmt.Errorf("mysql open error: %s: %w", e, ErrConnect)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql ping error: %w", err)
	}
	return db, nil
}

// ConnectInfo will connect to the database and return any errors.
func ConnectInfo() string {
	c := Init()
	db, err := sql.Open("mysql", c.String())
	defer func() {
		db.Close()
	}()
	if err != nil {
		return strings.Replace(err.Error(), c.Pass, "****", 1)
	}
	if err = db.Ping(); err != nil {
		if err, ok := err.(*mysql.MySQLError); ok {
			e := strings.Replace(err.Error(), c.User, color.Primary.Sprint(c.User), 1)
			return fmt.Sprintf("%s %v", color.Info.Sprint("MySQL"), color.Danger.Sprint(e))
		}
		if err, ok := err.(*net.OpError); ok {
			if strings.Contains(err.Error(), "connect: connection refused") {
				return fmt.Sprintf("%s '%v' %s",
					color.Danger.Sprint("database server"),
					color.Primary.Sprint(c.Server),
					color.Danger.Sprint("is either down or the port is blocked"))
			}
			return color.Danger.Sprint(err)
		}
	}
	return ""
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

// ColumnTypes details the columns used by the table.
func ColumnTypes(table string) error {
	db := Connect()
	defer db.Close()
	// LIMIT 0 quickly returns an empty set
	rows, err := db.Query("SELECT * FROM ? LIMIT 0", table)
	if err != nil {
		return fmt.Errorf("column types query: %w", err)
	} else if rows.Err() != nil {
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
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t\n", s.Name(), s.DatabaseTypeName(), nullable(s), collen(s))
	}
	if err = w.Flush(); err != nil {
		return fmt.Errorf("column types flush tab writer: %w", err)
	}
	return nil
}

// DateTime colours and formats a date and time string.
func DateTime(b sql.RawBytes) string {
	v := string(b)
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
	} else if err != nil {
		return false, fmt.Errorf("file update stat %q: %w", name, err)
	}
	return !f.ModTime().UTC().After(database.UTC()), nil
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

// IsNew reports if a file record is set to unapproved.
func IsNew(b []sql.RawBytes) bool {
	if b[2] == nil {
		return false
	}
	n, err := valid(b[2], b[3])
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
func LastUpdate() (t time.Time, err error) {
	db := Connect()
	defer db.Close()
	row := db.QueryRow("SELECT `updatedat` FROM `files` WHERE `deletedat` <> `updatedat` ORDER BY `updatedat` DESC LIMIT 1")
	if err := row.Scan(&t); err != nil {
		return t, fmt.Errorf("last update: %w", err)
	}
	return t, nil
}

// LookupID returns the ID from a supplied UUID or database ID value.
func LookupID(s string) (id uint, err error) {
	db := Connect()
	defer db.Close()
	// auto increment numeric ids
	var v int
	if v, err = strconv.Atoi(s); err == nil {
		// https://stackoverflow.com/questions/1676551/best-way-to-test-if-a-row-exists-in-a-mysql-table
		if err = db.QueryRow("SELECT EXISTS(SELECT * FROM files WHERE id=?)", v).Scan(&id); err != nil {
			return 0, fmt.Errorf("lookupid query row: %w", err)
		} else if id == 0 {
			return 0, fmt.Errorf("lookupid %q: %w", v, ErrNoID)
		}
		return uint(v), nil
	}
	// uuid ids
	s = strings.ToLower(s)
	if err = db.QueryRow("SELECT id FROM files WHERE uuid=?", s).Scan(&id); err != nil {
		return 0, fmt.Errorf("lookupid %q: %w", s, ErrNoID)
	}
	return id, db.Close()
}

// LookupFile returns the filename from a supplied UUID or database ID value.
func LookupFile(s string) (name string, err error) {
	db := Connect()
	defer db.Close()
	var n sql.NullString
	var v int
	if v, err = strconv.Atoi(s); err == nil {
		err = db.QueryRow("SELECT filename FROM files WHERE id=?", v).Scan(&n)
		if err != nil {
			return "", fmt.Errorf("lookup file by id queryrow %q: %w", s, err)
		}
	} else {
		s = strings.ToLower(s)
		err = db.QueryRow("SELECT filename FROM files WHERE uuid=?", s).Scan(&n)
		if err != nil {
			return "", fmt.Errorf("lookup file by uuid queryrow %q: %w", s, err)
		}
	}
	return n.String, nil
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
	r, err := reverseInt(uint(pint))
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
	a ^= 461
	b += 154
	return strconv.FormatInt(int64(b), 16) + strconv.FormatInt(int64(a), 16)
}

// Rename replaces all instances of the group name with a new group name.
func Rename(replacement, group string) (count int64, err error) {
	db := Connect()
	defer db.Close()
	stmt, err := db.Prepare("UPDATE `files` SET group_brand_for=?, group_brand_by=? WHERE (group_brand_for=? OR group_brand_by=?)")
	if err != nil {
		return 0, fmt.Errorf("rename group statement: %w", err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(replacement, replacement, group, group)
	if err != nil {
		return 0, fmt.Errorf("rename group exec: %w", err)
	}
	count, err = res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rename group rows affected: %w", err)
	}
	return count, db.Close()
}

// Total reports the number of records fetched by the supplied SQL query.
func Total(s *string) (sum int, err error) {
	db := Connect()
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
	for rows.Next() {
		sum++
	}
	return sum, db.Close()
}

func collen(s *sql.ColumnType) string {
	l, ok := s.Length()
	if !ok {
		return ""
	}
	if l > 0 {
		return strconv.Itoa(int(l))
	}
	return ""
}

// defaults initializes default connection settings.
func defaults() {
	viper.SetDefault("connection.name", "defacto2-inno")
	viper.SetDefault("connection.user", "root")
	viper.SetDefault("connection.password", "password")
	viper.SetDefault("connection.server.protocol", "tcp")
	viper.SetDefault("connection.server.host", "localhost")
	viper.SetDefault("connection.server.port", "3306")
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

// reverseInt swaps the direction of the value, 12345 would return 54321.
func reverseInt(value uint) (reversed uint, err error) {
	v, s, n := strconv.Itoa(int(value)), "", 0
	for x := len(v); x > 0; x-- {
		s += string(v[x-1])
	}
	if n, err = strconv.Atoi(s); err != nil {
		return value, fmt.Errorf("reverse int %q: %w", s, err)
	}
	return uint(n), nil
}

func valid(deleted, updated sql.RawBytes) (bool, error) {
	const (
		min = -5
		max = 5
	)
	// normalise the date values as sometimes updatedat & deletedat can be off by a second.
	del, err := time.Parse(time.RFC3339, string(deleted))
	if err != nil {
		return false, fmt.Errorf("valid deleted time: %w", err)
	}
	upd, err := time.Parse(time.RFC3339, string(updated))
	if err != nil {
		return false, fmt.Errorf("valid updated time: %w", err)
	}
	if diff := upd.Sub(del); diff.Seconds() > max || diff.Seconds() < min {
		return false, nil
	}
	return true, nil
}
