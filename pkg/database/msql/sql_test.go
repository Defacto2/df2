package msql_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database/msql"
	"github.com/stretchr/testify/assert"
)

func TestVersion_Query(t *testing.T) {
	t.Parallel()
	cfg := conf.Defaults()
	var v msql.Version
	err := v.Query(cfg)
	assert.Nil(t, err)
	assert.Contains(t, v, "MariaDB")
}
