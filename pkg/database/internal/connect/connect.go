package connect

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/Defacto2/df2/pkg/logs"
	"github.com/go-sql-driver/mysql"
	"github.com/gookit/color"
	"github.com/spf13/viper"
)

var ErrConnect = errors.New("could not connect to the mysql database server")

const (
	// Datetime MySQL format.
	Datetime = "2006-01-02T15:04:05Z"
	// UpdateID is a user id to use with the updatedby column.
	UpdateID = "b66dc282-a029-4e99-85db-2cf2892fffcc"

	hide = "****"
)

// Connection information for a MySQL database.
type Connection struct {
	Name     string // Name of the database
	User     string // User name access.
	Pass     string // Pass is the user password.
	Server   string // Server is URI to connect to the database, using the protocol, address and port.
	Protocol string // Protocol to connect to the database.
	Address  string // Address to connect to the database.
	Port     string // Port to connect to the database.
}

func (c *Connection) String() string {
	return fmt.Sprintf("%v:%v@%v/%v?timeout=5s&parseTime=true",
		c.User, c.Pass, c.Server, c.Name)
}

// Init initialises the database connection using stored settings.
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
		e := strings.Replace(err.Error(), c.Pass, hide, 1)
		log.Fatal(fmt.Errorf("connect database open: %s: %w", e, ErrConnect))
	}
	// ping the server to make sure the connection works
	if err = db.Ping(); err != nil {
		logs.Println(color.Secondary.Sprint(strings.Replace(c.String(), c.Pass, hide, 1)))
		// filter the password and then print the datasource connection info
		// to discover more errors, use log.Printf("%T", err)
		me := &mysql.MySQLError{}
		nop := &net.OpError{}
		switch {
		case errors.As(err, &me):
			log.Fatal(fmt.Errorf("connect mysql error: %w", err))
		case errors.As(err, &nop):
			switch nop.Op {
			case "dial":
				log.Fatal(fmt.Errorf("database server %v is either down or the %v %v port is blocked: %w",
					c.Address, c.Protocol, c.Port, ErrConnect))
			case "read":
				log.Fatal(fmt.Errorf("connect database read: %w", err))
			case "write":
				log.Fatal(fmt.Errorf("connect database write: %w", err))
			default:
				log.Fatal(fmt.Errorf("connect database op: %w", err))
			}
		}
		log.Fatal(fmt.Errorf("connect database: %w", err))
	}
	return db
}

// defaults initialises default connection settings.
func defaults() {
	viper.SetDefault("connection.name", "defacto2-inno")
	viper.SetDefault("connection.user", "root")
	viper.SetDefault("connection.password", "password")
	viper.SetDefault("connection.server.protocol", "tcp")
	viper.SetDefault("connection.server.host", "localhost")
	viper.SetDefault("connection.server.port", "3306")
}
