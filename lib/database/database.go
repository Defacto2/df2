package database

import (
	"database/sql"
	"fmt"
	"math"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
	"github.com/spf13/viper"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql" // MySQL database driver
	"github.com/google/uuid"
)

// UpdateID is a user id to use with the updatedby column.
const UpdateID string = "b66dc282-a029-4e99-85db-2cf2892fffcc"

// Datetime MySQL 5.7 format.
const Datetime string = "2006-01-02T15:04:05Z"

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

// Empty is used as a blank value for search maps.
// See: https://dave.cheney.net/2014/03/25/the-empty-struct
type Empty struct{}

// IDs are unique UUID values used by the database and filenames.
type IDs map[string]struct{}

func (c *Connection) String() string {
	return fmt.Sprintf("%v:%v@%v/%v?timeout=5s&parseTime=true", c.User, c.Pass, c.Server, c.Name)
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
	db, err := sql.Open("mysql", fmt.Sprint(&c))
	if err != nil {
		e := strings.Replace(err.Error(), c.Pass, "****", 1)
		logs.Check(fmt.Errorf(e))
	}
	err = db.Ping() // ping the server to make sure the connection works
	if err != nil {
		logs.Println(color.Secondary.Sprint(strings.Replace(fmt.Sprint(&c), c.Pass, "****", 1)))
		// filter the password and then print the datasource connection info
		// to discover more errors fmt.Printf("%T", err)
		var p = color.Primary.Sprint
		if err, ok := err.(*mysql.MySQLError); ok {
			e := strings.Replace(err.Error(), c.User, p(c.User), 1)
			logs.Check(fmt.Errorf("%s %v", color.Info.Sprint("MySQL"), e))
		}
		if err, ok := err.(*net.OpError); ok {
			if strings.Contains(err.Error(), "connect: connection refused") {
				logs.Check(fmt.Errorf("database server %v is either down or the %v %v port is blocked", p(c.Address), p(c.Protocol), p(c.Port)))
			} else {
				logs.Check(err)
			}
		}
	}
	return db
}

// ConnectErr will connect to the database or return any errors.
func ConnectErr() (db *sql.DB, err error) {
	c := Init()
	db, err = sql.Open("mysql", fmt.Sprint(&c))
	if err != nil {
		e := strings.Replace(err.Error(), c.Pass, "****", 1)
		return nil, fmt.Errorf(e)
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// ConnectInfo will connect to the database and return any errors.
func ConnectInfo() string {
	c := Init()
	db, err := sql.Open("mysql", fmt.Sprint(&c))
	defer db.Close()
	if err != nil {
		e := strings.Replace(err.Error(), c.Pass, "****", 1)
		return fmt.Sprint(fmt.Errorf(e))
	}
	err = db.Ping()
	if err != nil {
		if err, ok := err.(*mysql.MySQLError); ok {
			e := strings.Replace(err.Error(), c.User, color.Primary.Sprint(c.User), 1)
			return fmt.Sprint(fmt.Errorf("%s %v", color.Info.Sprint("MySQL"), color.Danger.Sprint(e)))
		}
		if err, ok := err.(*net.OpError); ok {
			if strings.Contains(err.Error(), "connect: connection refused") {
				return fmt.Sprint(fmt.Errorf("%s '%v' %s", color.Danger.Sprint("database server"), color.Primary.Sprint(c.Server), color.Danger.Sprint("is either down or the port is blocked")))
			}
			return color.Danger.Sprint(err)
		}
	}
	return ""
}

// DateTime colours and formats a date and time string.
func DateTime(value sql.RawBytes) string {
	v := string(value)
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

// CheckID reports an error message for an incorrect universal unique record id or MySQL auto-generated id.
func CheckID(id string) error {
	if !IsUUID(id) && !IsID(id) {
		return fmt.Errorf("invalid id given %q it needs to be an auto-generated MySQL id or an uuid", id)
	}
	return nil
}

// CheckUUID reports an error message for an incorrect universal unique record id.
func CheckUUID(id string) error {
	if !IsUUID(id) {
		return fmt.Errorf("%q is an invalid uuid it requires RFC 4122 syntax: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", id)
	}
	return nil
}

// ColumnTypes details the columns used by the table.
func ColumnTypes(table string) {
	db := Connect()
	defer db.Close()
	// LIMIT 0 quickly returns an empty set
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM `%s` LIMIT 0", table))
	logs.Check(err)
	colTypes, _ := rows.ColumnTypes()
	logs.Check(err)
	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Column name\tType\tNullable\tLength\t")
	for _, s := range colTypes {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t\n", s.Name(), s.DatabaseTypeName(), printNull(s), printLen(s))
	}
	w.Flush()
}

func printNull(s *sql.ColumnType) (null string) {
	n, ok := s.Nullable()
	if !ok {
		return null
	}
	switch n {
	case true:
		null = "✓"
	default:
		null = "✗"
	}
	return null
}

func printLen(s *sql.ColumnType) (length string) {
	l, ok := s.Length()
	if !ok {
		return length
	}
	if l > 0 {
		length = strconv.Itoa(int(l))
	}
	return length
}

// IsID reports whether id is a autogenerated record id.
func IsID(id string) bool {
	r := regexp.MustCompile(`^0+`)
	if x := r.ReplaceAllString(id, ""); x != id {
		return false // leading zeros found
	}
	if i, err := strconv.Atoi(id); err != nil {
		return false // not a number
	} else if f := float64(i); f != math.Abs(f) {
		return false // not an absolute value
	}
	return true
}

// IsUUID reports whether id is a universal unique record id.
func IsUUID(id string) bool {
	if _, err := uuid.Parse(id); err != nil {
		return false
	}
	return true
}

// IsNew reports if a file record is set to unapproved.
func IsNew(values []sql.RawBytes) bool {
	if values[2] == nil {
		return false
	}
	new, err := isNew(values[2], values[3])
	logs.Log(err)
	return new
}

func isNew(deleted, updated sql.RawBytes) (bool, error) {
	// normalise the date values as sometimes updatedat & deletedat can be off by a second.
	del, err := time.Parse(time.RFC3339, string(deleted))
	if err != nil {
		return false, err
	}
	upd, err := time.Parse(time.RFC3339, string(updated))
	if err != nil {
		return false, err
	}
	if diff := upd.Sub(del); diff.Seconds() > 5 || diff.Seconds() < -5 {
		return false, nil
	}
	return true, nil
}

// FileUpdate reports if the file is newer than the database time.
func FileUpdate(name string, database time.Time) bool {
	f, err := os.Stat(name)
	if os.IsNotExist(err) {
		return true
	}
	logs.Check(err)
	mod := f.ModTime()
	return !mod.UTC().After(database.UTC())
}

// LastUpdate reports the time when the files database was last modified.
func LastUpdate() (updatedat time.Time) {
	db := Connect()
	defer db.Close()
	row := db.QueryRow("SELECT `updatedat` FROM `files` WHERE `deletedat` <> `updatedat` ORDER BY `updatedat` DESC LIMIT 1")
	err := row.Scan(&updatedat)
	logs.Check(err)
	return updatedat
}

// LookupID returns the number of file entries associated with a group.
func LookupID(value string) (id uint, err error) {
	db := Connect()
	defer db.Close()
	var s string
	if v, err := strconv.Atoi(value); err == nil {
		// https://stackoverflow.com/questions/1676551/best-way-to-test-if-a-row-exists-in-a-mysql-table
		s = fmt.Sprintf("SELECT EXISTS(SELECT * FROM files WHERE id='%d')", v)
		err = db.QueryRow(s).Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("lookupid: %s", err)
		} else if id == 0 {
			return 0, fmt.Errorf("lookupid: unique id '%v' is does not exist in the database", v)
		}
		return uint(v), nil
	}
	value = strings.ToLower(value)
	s = fmt.Sprintf("SELECT id FROM files WHERE uuid='%s'", value)
	err = db.QueryRow(s).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("lookupid: uuid '%v' is does not exist in the database", value)
	}
	return id, nil
}

// LookupFile returns the filename from a supplied UUID or database ID value.
func LookupFile(value string) (filename string, err error) {
	db := Connect()
	defer db.Close()
	var s string
	if v, err := strconv.Atoi(value); err == nil {
		s = fmt.Sprintf("SELECT filename FROM files WHERE id='%d'", v)
		err = db.QueryRow(s).Scan(&filename)
		if err != nil {
			return "", fmt.Errorf("lookupfile: %s", err)
		} else if filename == "" {
			return "", fmt.Errorf("lookupfile: unique id '%v' is does not exist in the database", v)
		}
		return filename, err
	}
	value = strings.ToLower(value)
	s = fmt.Sprintf("SELECT filename FROM files WHERE uuid='%s'", value)
	err = db.QueryRow(s).Scan(&filename)
	if err != nil {
		return "", fmt.Errorf("lookupfile: uuid '%v' is does not exist in the database", value)
	}
	return filename, err
}

// ObfuscateParam hides the param value using the method implemented in CFWheels obfuscateParam() helper.
func ObfuscateParam(param string) string {
	rv := param // return value
	// check to make sure param doesn't begin with a 0 digit
	if rv0 := rv[0]; rv0 == '0' {
		return rv
	}
	paramInt, err := strconv.Atoi(param) // convert param to an int type
	if err != nil {
		return rv
	}
	iEnd := len(rv) // count the number of digits in param
	afloat64 := math.Pow10(iEnd) + float64(reverseInt(paramInt))
	// keep a and b as int type
	a := int(afloat64)
	b := 0
	for i := 1; i <= iEnd; i++ {
		// slice individual digits from param and sum them
		paramSlice, err := strconv.Atoi(string(param[iEnd-i]))
		if err != nil {
			return rv
		}
		b += paramSlice
	}
	// base 64 conversion
	a ^= 461
	b += 154
	return strconv.FormatInt(int64(b), 16) + strconv.FormatInt(int64(a), 16)
}

// reverseInt swaps the direction of the value, 12345 would return 54321.
func reverseInt(value int) (reversed int) {
	int := strconv.Itoa(value)
	new := ""
	for x := len(int); x > 0; x-- {
		new += string(int[x-1])
	}
	reversed, err := strconv.Atoi(new)
	logs.Check(err)
	return reversed
}

// RenGroup replaces all instances of the old group name with the new group name.
func RenGroup(new, old string) (count int64, err error) {
	db := Connect()
	defer db.Close()
	var sql = [2]string{
		"UPDATE files SET group_brand_for=? WHERE group_brand_for=?",
		"UPDATE files SET group_brand_by=? WHERE group_brand_by=?",
	}
	for i := range sql {
		update, err := db.Prepare(sql[i])
		if err != nil {
			return 0, err
		}
		res, err := update.Exec(new, old)
		if err != nil {
			return 0, err
		}
		count, err = res.RowsAffected()
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

// Total reports the number of records fetched by the supplied SQL query.
func Total(s *string) (total int) {
	db := Connect()
	rows, err := db.Query(*s)
	if err != nil && strings.Contains(err.Error(), "SQL syntax") {
		logs.Println(s)
	}
	logs.Check(err)
	defer db.Close()
	total = 0
	for rows.Next() {
		total++
	}
	return total
}
