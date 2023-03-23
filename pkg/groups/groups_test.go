package groups_test

import (
	"io"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/stretchr/testify/assert"
)

func TestRequest_DataList(t *testing.T) {
	t.Parallel()
	r := groups.Request{}
	err := r.DataList(nil, nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = r.DataList(db, io.Discard, nil)
	assert.NotNil(t, err)
}

func TestCronjob(t *testing.T) {
	t.Parallel()
	err := groups.Cronjob(nil, nil, nil, "", false)
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = groups.Cronjob(db, io.Discard, io.Discard, "", false)
	assert.NotNil(t, err)
	err = groups.Cronjob(db, io.Discard, io.Discard, "invalid-tag", false)
	assert.NotNil(t, err)
}

func TestExact(t *testing.T) {
	t.Parallel()
	i, err := groups.Exact(nil, "")
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	i, err = groups.Exact(db, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, i)
	i, err = groups.Exact(db, "Defacto2")
	assert.Nil(t, err)
	assert.Greater(t, i, 1)
}

func TestFix(t *testing.T) {
	t.Parallel()
	err := groups.Fix(nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = groups.Fix(db, io.Discard)
	assert.Nil(t, err)
}

func TestVariations(t *testing.T) {
	t.Parallel()
	s, err := groups.Variations(nil, "")
	assert.NotNil(t, err)
	assert.Empty(t, s)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	s, err = groups.Variations(db, "")
	assert.Nil(t, err)
	assert.Empty(t, s)
	s, err = groups.Variations(db, "hello")
	assert.Nil(t, err)
	assert.Contains(t, s, "hello")
	s, err = groups.Variations(db, "hello world")
	assert.Nil(t, err)
	assert.Contains(t, s, "hello world")
	assert.Contains(t, s, "helloworld")
	assert.Contains(t, s, "hello-world")
	assert.Contains(t, s, "hello_world")
	assert.Contains(t, s, "hello.world")
}
