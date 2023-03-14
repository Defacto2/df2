package msql_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/database/msql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestConnection_String(t *testing.T) {
	c := msql.Connection{}
	assert.Equal(t, "user:password@tcp(localhost:3306)/?allowCleartextPasswords=true&parseTime=true", c.String())
	c = msql.Connection{
		User:      "root",
		Password:  "qwerty",
		HostName:  "example.com",
		HostPort:  3360,
		NoSSLMode: true,
	}
	assert.Equal(t, "root:qwerty@tcp(example.com:3360)/?allowCleartextPasswords=false&parseTime=true", c.String())
}

func TestConnectDB(t *testing.T) {
	got, err := msql.ConnectDB()
	assert.NotNil(t, got)
	assert.Nil(t, err)
}
