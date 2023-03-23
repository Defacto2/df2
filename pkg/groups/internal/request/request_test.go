package request_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups/internal/request"
	"github.com/stretchr/testify/assert"
)

func TestFlags_DataList(t *testing.T) {
	t.Parallel()
	f := request.Flags{}
	err := f.DataList(nil, nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = f.DataList(db, io.Discard, io.Discard)
	assert.NotNil(t, err)

	f = request.Flags{Filter: "bbs"}
	b := bytes.Buffer{}
	err = f.DataList(db, io.Discard, &b)
	assert.Nil(t, err)
	assert.Contains(t, b.String(), `<option value="`)
}

func TestFlags_HTML(t *testing.T) {
	t.Parallel()
	f := request.Flags{}
	err := f.HTML(nil, nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = f.HTML(db, io.Discard, io.Discard)
	assert.NotNil(t, err)

	f = request.Flags{Filter: "bbs"}
	b := bytes.Buffer{}
	err = f.HTML(db, io.Discard, &b)
	assert.Nil(t, err)
	assert.Contains(t, b.String(), `/a></h2><h2><a href="`)
}

func TestFlags_Files(t *testing.T) {
	t.Parallel()
	f := request.Flags{}
	_, err := f.Files(nil, "")
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	i, err := f.Files(db, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, i)
	f = request.Flags{Counts: true}
	i, err = f.Files(db, "Defacto")
	assert.Nil(t, err)
	assert.Greater(t, i, 1)
}

func TestFlags_Initialism(t *testing.T) {
	t.Parallel()
	f := request.Flags{}
	_, err := f.Initialism(nil, "")
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	s, err := f.Initialism(db, "")
	assert.Nil(t, err)
	assert.Equal(t, "", s)
	f = request.Flags{Initialisms: true}
	s, err = f.Initialism(db, "Defacto")
	assert.Nil(t, err)
	assert.Equal(t, "df", s)
}

func TestPrint(t *testing.T) {
	t.Parallel()
	r := request.Flags{}
	i, err := request.Print(nil, nil, r)
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	i, err = request.Print(db, io.Discard, r)
	assert.Nil(t, err)
	assert.Greater(t, i, 1)
}
