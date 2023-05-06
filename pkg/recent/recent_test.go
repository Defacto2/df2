package recent_test

import (
	"bytes"
	"database/sql"
	"io"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/recent"
	"github.com/stretchr/testify/assert"
)

const uuid = "d37e5b5f-f5bf-4138-9078-891e41b10a12"

func TestThumb_Scan(t *testing.T) {
	t.Parallel()
	f := recent.Thumb{}
	f.Scan(nil)
	assert.Equal(t, "", f.URLID)
	n := time.Now().Format(time.RFC3339)
	v := []sql.RawBytes{
		sql.RawBytes("1"),
		sql.RawBytes(uuid),
		sql.RawBytes("Placeholder title"),
		sql.RawBytes("For some group"),
		sql.RawBytes("By some group"),
		sql.RawBytes("file.txt"),
		sql.RawBytes("1990"),
		sql.RawBytes(n),
	}
	f.Scan(v)
	assert.NotEmpty(t, f)
}

func TestList(t *testing.T) {
	t.Parallel()
	err := recent.List(nil, nil, 1, false)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
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
