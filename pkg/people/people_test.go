package people_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/people"
	"github.com/Defacto2/df2/pkg/people/internal/role"
	"github.com/stretchr/testify/assert"
)

func TestCronjob(t *testing.T) {
	t.Parallel()
	err := people.Cronjob(nil, nil, "", false)
	assert.NotNil(t, err)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = people.Cronjob(db, io.Discard, "", false)
	assert.NotNil(t, err)
}

func TestDataList(t *testing.T) {
	t.Parallel()
	err := people.DataList(nil, nil, "", people.Flags{})
	assert.NotNil(t, err)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = people.DataList(db, io.Discard, "", people.Flags{})
	assert.Nil(t, err)
	bb := bytes.Buffer{}
	err = people.DataList(db, &bb, "", people.Flags{})
	assert.Nil(t, err)
	assert.Contains(t, bb.String(), `<option value="`)
}

func TestFilters(t *testing.T) {
	t.Parallel()
	f := people.Filters()
	assert.Len(t, f, 4)
	s := people.Roles()
	assert.Contains(t, s, role.Writers.String())
}

func TestHTML(t *testing.T) {
	t.Parallel()
	err := people.HTML(nil, nil, "", people.Flags{})
	assert.NotNil(t, err)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = people.HTML(db, io.Discard, "", people.Flags{})
	assert.Nil(t, err)
	bb := bytes.Buffer{}
	err = people.HTML(db, &bb, "", people.Flags{})
	assert.Nil(t, err)
	assert.Contains(t, bb.String(), `<h2><a href="/p/`)
}

func TestPrint(t *testing.T) {
	t.Parallel()
	err := people.Print(nil, nil, people.Flags{})
	assert.NotNil(t, err)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = people.Print(db, io.Discard, people.Flags{})
	assert.Nil(t, err)
	bb := &bytes.Buffer{}
	err = people.Print(db, bb, people.Flags{})
	assert.Nil(t, err)
	assert.Contains(t, bb.String(), `Total authors`)
}

func TestFix(t *testing.T) {
	t.Parallel()
	err := people.Fix(nil, nil)
	assert.NotNil(t, err)
	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	bb := &bytes.Buffer{}
	err = people.Fix(db, bb)
	assert.Nil(t, err)
	assert.Contains(t, bb.String(), `time taken`)
}
