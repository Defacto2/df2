package msql_test

import (
	"fmt"
	"testing"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database/msql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

const pass = "hello-world-password"

func conn() string {
	return "root:" + pass + "@tcp(example.com:3360)/?allowCleartextPasswords=false&parseTime=true&timeout=30s"
}

func TestConnection_String(t *testing.T) {
	t.Parallel()
	c := msql.Connection{}
	assert.Equal(t, ":@tcp(:0)/?allowCleartextPasswords=true&parseTime=true&timeout=30s", c.String())
	c = msql.Connection{
		User:      "root",
		Pass:      pass,
		Host:      "example.com",
		Port:      3360,
		NoSSLMode: true,
	}
	assert.Equal(t, conn, c.String())
}

func TestConnect(t *testing.T) {
	t.Parallel()
	cfg := conf.Defaults()
	got, err := msql.Connect(cfg)
	assert.NotNil(t, got)
	assert.Nil(t, err)
}

func TestMaskPass(t *testing.T) {
	t.Parallel()
	c := msql.Connection{
		Pass: pass,
	}
	err := fmt.Errorf("%w, %s", msql.ErrConnect, conn())
	err1 := c.MaskPass(err)
	assert.NotContains(t, err1.Error(), pass)
}

func TestConnInfo(t *testing.T) {
	t.Parallel()
	s, err := msql.ConnInfo(conf.Config{})
	assert.NotNil(t, err)
	assert.Equal(t, "", s)
	s, err = msql.ConnInfo(conf.Defaults())
	assert.Nil(t, err)
	assert.Equal(t, "", s)
}
