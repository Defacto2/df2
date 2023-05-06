package update_test

import (
	"io"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/database/internal/update"
	"github.com/stretchr/testify/assert"
)

func TestUpdate_Execute(t *testing.T) {
	t.Parallel()
	u := update.Update{}
	i, err := u.Execute(nil)
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	i, err = u.Execute(db)
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)

	u = update.Update{
		Query: "UPDATE files SET updatedat=? WHERE id=?",
		Args:  []any{time.Now(), 1},
	}
	i, err = u.Execute(db)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), i)
}

func TestColumn_NamedTitles(t *testing.T) {
	t.Parallel()
	var c update.Column
	err := c.NamedTitles(nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = c.NamedTitles(db, io.Discard)
	assert.Nil(t, err)
}

func TestDistinct(t *testing.T) {
	t.Parallel()
	s, err := update.Distinct(nil, "")
	assert.NotNil(t, err)
	assert.Empty(t, s)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	s, err = update.Distinct(db, "")
	assert.NotNil(t, err)
	assert.Empty(t, s)

	s, err = update.Distinct(db, "section")
	assert.Nil(t, err)
	assert.Len(t, s, 26)
	s, err = update.Distinct(db, "platform")
	assert.Nil(t, err)
	assert.Len(t, s, 17)
}

func TestSections(t *testing.T) {
	t.Parallel()
	err := update.Sections(nil, nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = update.Sections(db, io.Discard, nil)
	assert.NotNil(t, err)
	empty := []string{}
	err = update.Sections(db, io.Discard, &empty)
	assert.Nil(t, err)
	audio := []string{"audio", "audio"}
	err = update.Sections(db, io.Discard, &audio)
	assert.Nil(t, err)
}

func TestPlatforms(t *testing.T) {
	t.Parallel()
	err := update.Platforms(nil, nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = update.Platforms(db, io.Discard, nil)
	assert.NotNil(t, err)
	empty := []string{}
	err = update.Platforms(db, io.Discard, &empty)
	assert.Nil(t, err)
	dos := []string{"dos", "dos"}
	err = update.Platforms(db, io.Discard, &dos)
	assert.Nil(t, err)
}
