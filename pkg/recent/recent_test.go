package recent_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/recent"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	t.Parallel()
	err := recent.List(nil, nil, 1, false)
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()

	err = recent.List(db, io.Discard, 5, false)
	assert.Nil(t, err)

	bb := bytes.Buffer{}
	err = recent.List(db, &bb, 1, false)
	assert.Nil(t, err)
	assert.Contains(t, bb.String(), "COLUMNS")
	assert.Nil(t, err)
	err = recent.List(db, &bb, 1, true)
	assert.Nil(t, err)
	assert.Contains(t, bb.String(), "COLUMNS")
}
