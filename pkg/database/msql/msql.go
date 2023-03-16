package msql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/go-sql-driver/mysql"
	"github.com/gookit/color"
)

const (
	// Protocol of the database driver.
	Protocol = "tcp"
	// DriverName of the database.
	DriverName = "mysql"

	hider = "***"
)

var (
	ErrConfig   = errors.New("no connection configuration was provided")
	ErrConnect  = errors.New("no connection to the mysql database server")
	ErrNoConn   = errors.New("no pointer to an open database connection")
	ErrDatabase = errors.New("name of the database to connect is missing")
	ErrHost     = errors.New("host name of the database server is missing")
	ErrUser     = errors.New("user for database login is missing")
)

// Connection details of the MySQL database connection.
type Connection struct {
	Host     string // Host is the host name of the server.
	Port     uint   // Port is the port number the server is listening on.
	Database string // Database is the database name.
	User     string // User is the database user used to connect to the database.
	Pass     string // Pass is the password for the database user.
	// NoSSLMode connects to the database using an insecure,
	// plain text connecction using the sslmode=disable param.
	NoSSLMode bool
}

func (c *Connection) Check() error {
	var errs error
	if c.Host == "" {
		errs = errors.Join(errs, ErrHost)
	}
	if c.Database == "" {
		errs = errors.Join(errs, ErrDatabase)
	}
	if c.User == "" {
		errs = errors.Join(errs, ErrUser)
	}
	if errs != nil {
		return errs
	}
	return nil
}

// Open opens a MySQL database connection.
func (c Connection) Open() (*sql.DB, error) {
	conn, err := sql.Open(DriverName, c.String())
	if err != nil {
		return nil, c.HidePass(err)
	}
	return conn, nil
}

// HidePass returns a MySQL database connection error with the user password removed.
func (c Connection) HidePass(err error) error {
	if err == nil {
		return nil
	}
	s := strings.Replace(err.Error(), c.Pass, hider, 1)
	return fmt.Errorf("mysql connection: %s", s)
}

// String returns a MySQL database connection.
func (c Connection) String() string {
	login := fmt.Sprintf("%s:%s",
		c.User,
		c.Pass)
	address := fmt.Sprintf("%s(%s:%d)",
		Protocol,
		c.Host,
		c.Port)
	v := url.Values{}
	v.Add("allowCleartextPasswords", fmt.Sprint(!c.NoSSLMode))
	v.Add("timeout", "5s")
	v.Add("parseTime", "true") // parseTime is required by the SQL boiler pkg.

	// example connector: "user:password@tcp(localhost:5432)/database?sslmode=false"
	return fmt.Sprintf("%s@%s/%s?%s", login, address, c.Database, v.Encode())
}

func (c Connection) Ping(db *sql.DB) error {
	if db == nil {
		return ErrNoConn
	}
	// ping the server to make sure the connection works
	if err := db.Ping(); err != nil {
		//fmt.Fprintln(w, color.Secondary.Sprint(strings.Replace(c.String(), c.Pass, hide, 1)))
		// filter the password and then print the datasource connection info
		// to discover more errors, use log.Printf("%T", err)
		me := &mysql.MySQLError{}
		nop := &net.OpError{}
		switch {
		case errors.As(err, &me):
			return fmt.Errorf("mysql connection: %w", err)
		case errors.As(err, &nop):
			switch nop.Op {
			case "dial":
				log.Fatal(fmt.Errorf("database server %v is either down or the %v %v port is blocked: %w",
					c.Host, Protocol, c.Port, ErrConnect))
			case "read":
				log.Fatal(fmt.Errorf("mysql database read: %w", err))
			case "write":
				log.Fatal(fmt.Errorf("mysql database write: %w", err))
			default:
				log.Fatal(fmt.Errorf("mysql database op: %w", err))
			}
		}
		return fmt.Errorf("mysql database other: %w", err)
	}
	return nil
}

// Connect to the MySQL database.
func Connect(cfg configger.Config) (*sql.DB, error) {
	if cfg == (configger.Config{}) {
		return nil, ErrConfig
	}
	// TODO: cfg checks, copied from connection pkg?
	// TODO: handle nil cfg value to use default
	dsn := Connection{
		User:      cfg.DBUser,
		Pass:      cfg.DBPass,
		Database:  cfg.DBName,
		Host:      cfg.DBHost,
		Port:      cfg.DBPort,
		NoSSLMode: true, // use IsProd...
	}
	if err := dsn.Check(); err != nil {
		return nil, err
	}
	conn, err := dsn.Open()
	if err != nil {
		return nil, err
	}
	if err := dsn.Ping(conn); err != nil {
		return nil, err
	}
	return conn, nil
}

func Info(db *sql.DB, cfg configger.Config) string {
	err := db.Ping()
	if err == nil {
		return ""
	}
	me := &mysql.MySQLError{}
	if ok := errors.As(err, &me); ok {
		e := strings.Replace(err.Error(), cfg.DBUser, color.Primary.Sprint(cfg.DBUser), 1)
		return fmt.Sprintf("%s %v", color.Info.Sprint("MySQL"), color.Danger.Sprint(e))
	}
	nop := &net.OpError{}
	if ok := errors.As(err, &nop); ok {
		if strings.Contains(err.Error(), "connect: connection refused") {
			return fmt.Sprintf("%s '%v' %s",
				color.Danger.Sprint("database server"),
				color.Primary.Sprintf("tcp(%s:%d)", cfg.DBHost, cfg.DBPort),
				color.Danger.Sprint("is either down or the port is blocked"))
		}
		return color.Danger.Sprint(err)
	}
	return ""
}
