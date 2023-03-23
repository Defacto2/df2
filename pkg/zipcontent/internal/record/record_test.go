package record_test

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/zipcontent/internal/record"
	"github.com/Defacto2/df2/pkg/zipcontent/internal/scan"
	"github.com/stretchr/testify/assert"
)

var dzDir = filepath.Join("..", "..", "..", "..", "testdata", "demozoo")

func TestNew(t *testing.T) {
	r, err := record.New(nil, "")
	assert.NotNil(t, err)
	assert.Empty(t, r)

	const ids, uuids = "345674", "b4ef0174-57b4-11ec-bf63-0242ac130002"
	const id, uuid, filename, readme = 0, 1, 4, 6
	vals := make([]sql.RawBytes, 7)
	vals[id] = sql.RawBytes(ids)
	vals[uuid] = sql.RawBytes(uuids)
	vals[filename] = sql.RawBytes("somefile.zip")
	vals[readme] = sql.RawBytes("readme.txt")

	r, err = record.New(vals, "some-directory")
	assert.Nil(t, err)
	assert.NotEmpty(t, r)
}

func TestRecord_Iterate(t *testing.T) {
	r := record.Record{}
	err := r.Iterate(nil, nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()

	err = r.Iterate(db, io.Discard, nil)
	assert.NotNil(t, err)

	vals := []sql.RawBytes{
		sql.RawBytes("0"),
		sql.RawBytes(time.Now().String()),
		sql.RawBytes("somefile.zip"),
		sql.RawBytes(""),
	}
	badSt := scan.Stats{
		Columns: []string{"1ee21218-5898-11ec-bf63-0242ac130002"},
		Values:  &vals,
	}
	err = r.Iterate(db, io.Discard, &badSt)
	assert.NotNil(t, err)

	okSt := scan.Stats{
		Columns: []string{"id", "createdat", "filename", "uuid"},
		Values:  &vals,
	}
	err = r.Iterate(db, io.Discard, &okSt)
	assert.NotNil(t, err)
}

func TestRecord_Archive(t *testing.T) {
	const uuid = "1ee21218-5898-11ec-bf63-0242ac130002"
	r := record.Record{}
	err := r.Archive(nil, nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = r.Archive(db, io.Discard, nil)
	assert.NotNil(t, err)

	vals := []sql.RawBytes{
		sql.RawBytes("0"),
		sql.RawBytes(time.Now().String()),
		sql.RawBytes("somefile.zip"),
		sql.RawBytes(""),
	}
	badSt := scan.Stats{
		Columns: []string{uuid},
		Values:  &vals,
	}
	err = r.Archive(db, io.Discard, &badSt)
	assert.NotNil(t, err)

	okSt := scan.Stats{
		Columns: []string{"id", "createdat", "filename", "uuid"},
		Values:  &vals,
	}

	r = record.Record{
		ID:   "1",
		UUID: uuid,
		File: filepath.Join(dzDir, "test.zip"),
		Name: "test.png",
	}
	err = r.Archive(db, io.Discard, &okSt)
	assert.Nil(t, err)
	defer os.Remove(uuid + ".txt")
}
