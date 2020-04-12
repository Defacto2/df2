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
	Name   string // database name
	User   string // access username
	Pass   string // access password
	Server string // host server protocol, address and port
}

// Empty is used as a blank value for search maps.
// See: https://dave.cheney.net/2014/03/25/the-empty-struct
type Empty struct{}

// IDs are unique UUID values used by the database and filenames.
type IDs map[string]struct{}

var c = Connection{} // connection details

func (c *Connection) String() string {
	return fmt.Sprintf("%v:%v@%v/%v?timeout=5s&parseTime=true", c.User, c.Pass, c.Server, c.Name)
}

// Connect will connect to the database and handle any errors.
func Connect() *sql.DB {
	config()
	db, err := sql.Open("mysql", fmt.Sprint(&c))
	if err != nil {
		e := strings.Replace(err.Error(), c.Pass, "****", 1)
		logs.Check(fmt.Errorf(e))
	}
	err = db.Ping() // ping the server to make sure the connection works
	if err != nil {
		println(color.Secondary.Sprint(strings.Replace(fmt.Sprint(&c), c.Pass, "****", 1)))
		// filter the password and then print the datasource connection info
		// to discover more errors fmt.Printf("%T", err)
		if err, ok := err.(*mysql.MySQLError); ok {
			e := strings.Replace(err.Error(), c.User, color.Primary.Sprint(c.User), 1)
			logs.Check(fmt.Errorf("%s %v", color.Info.Sprint("MySQL"), e))
		}
		if err, ok := err.(*net.OpError); ok {
			if strings.Contains(err.Error(), "connect: connection refused") {
				logs.Check(fmt.Errorf("database server '%v' is either down or the port is blocked", color.Primary.Sprint(c.Server)))
			} else {
				logs.Check(err)
			}
		}
	}
	return db
}

// ConnectErr will connect to the database or return any errors.
func ConnectErr() (*sql.DB, error) {
	config()
	db, err := sql.Open("mysql", fmt.Sprint(&c))
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
	config()
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
		n, _ := s.Nullable()
		var null string
		switch n {
		case true:
			null = "✓"
		default:
			null = "✗"
		}
		l, _ := s.Length()
		var len string
		if l > 0 {
			len = strconv.Itoa(int(l))
		}
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t\n", s.Name(), s.DatabaseTypeName(), null, len)
	}
	w.Flush()
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
func LastUpdate() time.Time {
	db := Connect()
	defer db.Close()
	var updatedat time.Time
	row := db.QueryRow("SELECT `updatedat` FROM `files` WHERE `deletedat` <> `updatedat` ORDER BY `updatedat` DESC LIMIT 1")
	err := row.Scan(&updatedat)
	logs.Check(err)
	return updatedat
}

// RenGroup replaces all instances of the old group name with the new group name.
func RenGroup(new, old string) (int64, error) {
	db := Connect()
	defer db.Close()
	var sql = [2]string{
		"UPDATE files SET group_brand_for=? WHERE group_brand_for=?",
		"UPDATE files SET group_brand_by=? WHERE group_brand_by=?",
	}
	var ra int64
	for i := range sql {
		update, err := db.Prepare(sql[i])
		if err != nil {
			return 0, err
		}
		res, err := update.Exec(new, old)
		if err != nil {
			return 0, err
		}
		ra, err = res.RowsAffected()
		if err != nil {
			return 0, err
		}
	}
	return ra, nil
}

// Total reports the number of records fetched by the supplied SQL query.
func Total(s *string) int {
	db := Connect()
	rows, err := db.Query(*s)
	if err != nil && strings.Contains(err.Error(), "SQL syntax") {
		println(s)
	}
	logs.Check(err)
	defer db.Close()
	total := 0
	for rows.Next() {
		total++
	}
	return total
}

// config initializes the database connection using stored settings.
func config() {
	if c != (Connection{}) { // check for empty struct
		return
	}
	c = Connection{
		Name: viper.GetString("connection.name"),
		User: viper.GetString("connection.user"),
		Pass: viper.GetString("connection.password"),
		Server: fmt.Sprintf("%v(%v:%v)", // example: tcp(localhost:3306)
			viper.GetString("connection.server.protocol"),
			viper.GetString("connection.server.host"),
			viper.GetString("connection.server.port")),
	}
}
