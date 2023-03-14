package msql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // MySQL driver.
)

const (
	// Protocol of the database driver.
	Protocol = "tcp"
	// DriverName of the database.
	DriverName = "mysql"
)

// Connection details of the MySQL database connection.
type Connection struct {
	User     string // User is the database user used to connect to the database.
	Password string // Password is the password for the database user.
	HostName string // HostName is the host name of the server. Defaults to localhost.
	HostPort int    // HostPort is the port number the server is listening on. Defaults to 5432.
	Database string // Database is the database name.
	// NoSSLMode connects to the database using an insecure,
	// plain text connecction using the sslmode=disable param.
	NoSSLMode bool
}

// Open opens a MySQL database connection.
func (c Connection) Open() (*sql.DB, error) {
	conn, err := sql.Open(DriverName, c.String())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// String returns a MySQL database connection.
func (c Connection) String() string {
	// "user:password@tcp(localhost:5432)/database?sslmode=false"
	if c.User == "" {
		c.User = "user"
	}
	if c.Password == "" {
		c.Password = "password"
	}
	if c.HostName == "" {
		c.HostName = "localhost"
	}
	if c.HostPort < 1 {
		c.HostPort = 3306
	}
	// parseTime=true is required by SQL boiler.
	return fmt.Sprintf("%s:%s@%s(%s:%d)/%s?allowCleartextPasswords=%t&parseTime=true",
		c.User,
		c.Password,
		Protocol,
		c.HostName,
		c.HostPort,
		c.Database,
		!c.NoSSLMode,
	)
}

// ConnectDB connects to the MySQL database.
func ConnectDB() (*sql.DB, error) {
	dsn := Connection{
		User:      "root",
		Password:  "example",
		Database:  "defacto2-inno",
		NoSSLMode: true,
	}
	conn, err := dsn.Open()
	if err != nil {
		return nil, err
	}
	return conn, nil
}
