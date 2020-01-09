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
	"time"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/viper"

	_ "github.com/go-sql-driver/mysql" // MySQL database driver
	"github.com/google/uuid"
)

// UpdateID is a user id to use with the updatedby column.
const UpdateID string = "b66dc282-a029-4e99-85db-2cf2892fffcc"

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

// Connect to the database.
func Connect() *sql.DB {
	config()
	//cfg := fmt.Sprint(c)
	db, err := sql.Open("mysql", fmt.Sprint(&c))
	logs.Check(err) // TODO this could log the password, so search and filter it out
	err = db.Ping() // ping the server to make sure the connection works
	if err != nil {
		println(strings.Replace(fmt.Sprint(&c), c.Pass, "****", 1)) // filter the password and then print the datasource connection info
		// to discover more errors fmt.Printf("%T", err)
		if err, ok := err.(*net.OpError); ok {
			if strings.Contains(err.Error(), "connect: connection refused") {
				logs.Check(fmt.Errorf("the database server is either down or the port is blocked"))
			} else {
				logs.Check(err)
			}
		}
	}
	return db
}

// ID reports whether id is a autogenerated record id.
func ID(id string) bool {
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

// UUID reports whether id is a universal unique record id.
func UUID(id string) bool {
	if _, err := uuid.Parse(id); err != nil {
		return false
	}
	return true
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
