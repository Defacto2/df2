package msql_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database/msql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestConnection_String(t *testing.T) {
	c := msql.Connection{}
	assert.Equal(t, ":@tcp(:0)/?allowCleartextPasswords=true&parseTime=true&timeout=5s", c.String())
	c = msql.Connection{
		User:      "root",
		Pass:      "qwerty",
		Host:      "example.com",
		Port:      3360,
		NoSSLMode: true,
	}
	assert.Equal(t, "root:qwerty@tcp(example.com:3360)/?allowCleartextPasswords=false&parseTime=true&timeout=5s", c.String())
}

func TestConnect(t *testing.T) {
	cfg := configger.Defaults()
	got, err := msql.Connect(cfg)
	assert.NotNil(t, got)
	assert.Nil(t, err)
}
