package people_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/people"
	"github.com/Defacto2/df2/pkg/people/internal/role"
	"github.com/stretchr/testify/assert"
)

func TestCronjob(t *testing.T) {
	err := people.Cronjob(nil, nil, "", false)
	assert.NotNil(t, err)
	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = people.Cronjob(db, io.Discard, "", false)
	assert.NotNil(t, err)
}

func TestDataList(t *testing.T) {
	err := people.DataList(nil, nil, "", people.Flags{})
	assert.NotNil(t, err)
	db, err := database.Connect(configger.Defaults())
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
	f := people.Filters()
	assert.Len(t, f, 4)
	s := people.Roles()
	assert.Contains(t, s, role.Writers.String())
}

func TestHTML(t *testing.T) {
	err := people.HTML(nil, nil, "", people.Flags{})
	assert.NotNil(t, err)
	db, err := database.Connect(configger.Defaults())
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
	err := people.Print(nil, nil, people.Flags{})
	assert.NotNil(t, err)
	db, err := database.Connect(configger.Defaults())
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
	err := people.Fix(nil, nil)
	assert.NotNil(t, err)
	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	bb := &bytes.Buffer{}
	err = people.Fix(db, bb)
	assert.Nil(t, err)
	assert.Contains(t, bb.String(), `time taken`)
}
