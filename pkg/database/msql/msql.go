package msql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/go-sql-driver/mysql"
	"github.com/gookit/color"
)

const (
	// Protocol of the database driver.
	Protocol = "tcp"
	// DriverName of the database.
	DriverName = "mysql"
	// Timeout default in seconds for a database connection.
	Timeout = 30

	mask = "***"
)

var (
	ErrConfig   = errors.New("no connection configuration was provided")
	ErrConnect  = errors.New("no connection to the mysql database server")
	ErrDB       = errors.New("database handle pointer cannot be nil")
	ErrDatabase = errors.New("name of the database to connect is missing")
	ErrTimeout  = errors.New("connection timeout is too short")
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
	Timeout   time.Duration // Timeout context in seconds.
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
	if c.Timeout < 1*time.Second {
		errs = errors.Join(errs, fmt.Errorf("%w: %v", ErrTimeout, c.Timeout))
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
		return nil, c.MaskPass(err)
	}
	return conn, nil
}

// MaskPass returns a MySQL database connection error with the user password removed.
func (c Connection) MaskPass(err error) error {
	if err == nil {
		return nil
	}
	s := strings.Replace(err.Error(), c.Pass, mask, 1)
	return fmt.Errorf("mysql connection: %s", s) //nolint:goerr113
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
	if c.Timeout == 0 {
		c.Timeout = Timeout
	}
	v.Add("allowCleartextPasswords", fmt.Sprint(!c.NoSSLMode))
	v.Add("timeout", fmt.Sprintf("%v", c.Timeout))
	v.Add("parseTime", "true") // parseTime is required by the SQL boiler pkg.
	// example connector: "user:password@tcp(localhost:5432)/database?sslmode=false"
	return fmt.Sprintf("%s@%s/%s?%s", login, address, c.Database, v.Encode())
}

func (c Connection) Ping(db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if c.Timeout == 0 {
		c.Timeout = Timeout
	}
	// ping the server to make sure the connection works
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()
	err := db.PingContext(ctx)
	if err == nil {
		return nil
	}
	// filter the password and then print the datasource connection info
	// to discover more errors, use log.Printf("%T", err)
	my := &mysql.MySQLError{}
	op := &net.OpError{}
	switch {
	case errors.As(err, &my):
		return fmt.Errorf("mysql connection: %w", err)
	case errors.As(err, &op):
		switch op.Op {
		case "dial":
			return fmt.Errorf("database server %v is either down or the %v %v port is blocked: %w",
				c.Host, Protocol, c.Port, ErrConnect)
		case "read":
			return fmt.Errorf("mysql read: %w", err)
		case "write":
			return fmt.Errorf("mysql write: %w", err)
		default:
			return fmt.Errorf("mysql op: %w", err)
		}
	}
	return fmt.Errorf("mysql database: %w", err)
}

// Connect to and open the database.
// This must be closed after use.
func Connect(cfg conf.Config) (*sql.DB, error) {
	if cfg == (conf.Config{}) {
		return nil, ErrConfig
	}
	dsn := Connection{
		User:      cfg.DBUser,
		Pass:      cfg.DBPass,
		Database:  cfg.DBName,
		Host:      cfg.DBHost,
		Port:      cfg.DBPort,
		NoSSLMode: !cfg.IsProduction,
		Timeout:   time.Duration(cfg.Timeout) * time.Second,
	}
	if dsn.Timeout == 0 {
		dsn.Timeout = Timeout
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

func ConnInfo(cfg conf.Config) (string, error) {
	db, err := Connect(cfg)
	if err != nil {
		return "", err
	}
	defer db.Close()
	err = db.Ping()
	if err == nil {
		return "", nil
	}
	my := &mysql.MySQLError{}
	if ok := errors.As(err, &my); ok {
		e := strings.Replace(err.Error(), cfg.DBUser, color.Primary.Sprint(cfg.DBUser), 1)
		return fmt.Sprintf("%s %v", color.Info.Sprint("MySQL"), color.Danger.Sprint(e)), nil
	}
	op := &net.OpError{}
	if ok := errors.As(err, &op); ok {
		if strings.Contains(err.Error(), "connect: connection refused") {
			return fmt.Sprintf("%s '%v' %s",
				color.Danger.Sprint("database server"),
				color.Primary.Sprintf("tcp(%s:%d)", cfg.DBHost, cfg.DBPort),
				color.Danger.Sprint("is either down or the port is blocked")), nil
		}
		return color.Danger.Sprint(err), nil
	}
	return "", nil
}
