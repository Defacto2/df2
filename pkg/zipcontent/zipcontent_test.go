package zipcontent_test

import (
	"bytes"
	"testing"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/zipcontent"
	"github.com/stretchr/testify/assert"
)

func TestFix(t *testing.T) {
	t.Parallel()
	err := zipcontent.Fix(nil, nil, nil, conf.Config{}, false)
	assert.NotNil(t, err)

	cfg := conf.Defaults()
	db, err := database.Connect(cfg)
	assert.Nil(t, err)
	defer db.Close()

	bb := &bytes.Buffer{}
	err = zipcontent.Fix(db, bb, nil, cfg, false)
	assert.Nil(t, err)
	assert.NotContains(t, bb.String(), "Total archives scanned")

	bb = &bytes.Buffer{}
	err = zipcontent.Fix(db, bb, nil, cfg, true)
	assert.Nil(t, err)
	assert.Contains(t, bb.String(), "Total archives scanned")
}
