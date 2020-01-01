package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

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

var (
	d      = Connection{} // connection details
	pwPath string         // path to a secured text file containing the d.User login password
)

// Connect to the database.
func Connect() *sql.DB {
	connectInit()
	pw := readPassword()
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@%v/%v?timeout=5s&parseTime=true", d.User, pw, d.Server, d.Name))
	logs.Check(err)
	err = db.Ping() // ping the server to make sure the connection works
	logs.Check(err)
	return db
}

// IsID xxx
func IsID(id string) bool {
	if _, err := strconv.Atoi(id); err != nil {
		return false
	}
	return true
}

// IsUUID xx
func IsUUID(id string) bool {
	if _, err := uuid.Parse(id); err != nil {
		return false
	}
	return true
}

// Update is a temp SQL update func.
func Update(id string, content string) {
	db := Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET file_zip_content=?,updatedat=NOW(),updatedby=?,platform=?,deletedat=NULL,deletedby=NULL WHERE id=?")
	logs.Check(err)
	r, err := update.Exec(content, UpdateID, "image", id)
	logs.Check(err)
	fmt.Println("Updated file_zip_content", r)
}

func connectInit() {
	if d != (Connection{}) { // check for empty struct
		return
	}
	d = Connection{
		Name: viper.GetString("connection.name"),
		User: viper.GetString("connection.user"),
		Pass: viper.GetString("connection.password"),
		Server: fmt.Sprintf("%v(%v:%v)", // example: tcp(localhost:3306)
			viper.GetString("connection.server.protocol"),
			viper.GetString("connection.server.host"),
			viper.GetString("connection.server.port")),
	}
	d.Pass = "password"
}

// readPassword attempts to read and return the Defacto2 database user password when stored in a local text file.
func readPassword() string {
	// fetch database password
	pwFile, err := os.Open(pwPath)
	// return an empty password if path fails
	if err != nil {
		//log.Print("WARNING: ", err)
		return d.Pass
	}
	defer pwFile.Close()
	pw, err := ioutil.ReadAll(pwFile)
	logs.Check(err)
	return strings.TrimSpace(fmt.Sprintf("%s", pw))
}
