package record_test

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/proof/internal/record"
	"github.com/Defacto2/df2/pkg/proof/internal/stat"
	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/buffer"
)

const uuid = "d37e5b5f-f5bf-4138-9078-891e41b10a12"

func TestNew(t *testing.T) {
	t.Parallel()
	r := record.New(nil, "")
	assert.Empty(t, r)
	v := []sql.RawBytes{
		sql.RawBytes("1"),
		sql.RawBytes(uuid),
		sql.RawBytes("placeholder"),
		sql.RawBytes("placeholder"),
		sql.RawBytes("file.txt"),
	}
	r = record.New(v, "somePath")
	assert.NotEmpty(t, r)
}

func TestRecord_Approve(t *testing.T) {
	t.Parallel()
	r := record.Record{}
	err := r.Approve(nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = r.Approve(db, io.Discard)
	assert.NotNil(t, err)

	r = record.Record{
		ID: "1",
	}
	err = r.Approve(db, io.Discard)
	assert.Nil(t, err)
}

func TestRecord_Iterate(t *testing.T) {
	t.Parallel()
	r := record.Record{}
	err := r.Iterate(nil, nil, conf.Config{}, stat.Proof{})
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = r.Iterate(db, io.Discard, conf.Config{}, stat.Proof{})
	assert.NotNil(t, err)

	r = record.Record{
		ID: "1",
	}
	raw := []sql.RawBytes{
		sql.RawBytes("1"),
		sql.RawBytes(time.Now().Format("2006-01-02T15:04:05Z")),
		sql.RawBytes("file.txt"),
		sql.RawBytes("readme.txt,prog.exe"),
	}
	p := stat.Proof{
		Columns: []string{"id", "createdat", "filename", "file_zip_content"},
		Values:  &raw,
	}
	err = r.Iterate(db, io.Discard, conf.Config{}, p)
	assert.Nil(t, err)
}

func TestUpdateZipContent(t *testing.T) {
	t.Parallel()
	err := record.UpdateZipContent(nil, nil, "", "", "", -999)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = record.UpdateZipContent(db, io.Discard, "", "", "", -999)
	assert.NotNil(t, err)
	err = record.UpdateZipContent(db, io.Discard, "1", "", "", 0)
	assert.Nil(t, err)
}

func TestRecord_Zip(t *testing.T) {
	t.Parallel()
	r := record.Record{}
	err := r.Zip(nil, nil, conf.Config{}, false)
	assert.NotNil(t, err)

	cfg := conf.Defaults()
	db, err := database.Connect(cfg)
	assert.Nil(t, err)
	defer db.Close()
	err = r.Zip(db, io.Discard, conf.Config{}, false)
	assert.NotNil(t, err)
	err = r.Zip(db, io.Discard, cfg, false)
	assert.NotNil(t, err)
	r = record.Record{
		ID:   "1",
		UUID: uuid,
		File: "someDir/file.txt",
		Name: "file.txt",
	}
	err = r.Zip(db, io.Discard, cfg, true)
	assert.NotNil(t, err)
	wd, err := os.Getwd()
	assert.Nil(t, err)
	path := filepath.Join(wd, "..", "..", "..", "..", "testdata", "demozoo", "test.zip")
	r = record.Record{
		ID:   "1",
		UUID: uuid,
		File: path,
		Name: "test.zip",
	}
	err = r.Zip(db, io.Discard, cfg, true)
	assert.Nil(t, err)
}

func TestSkip(t *testing.T) {
	t.Parallel()
	color.Enable = false
	b, err := record.Skip(nil, stat.Proof{}, record.Record{})
	assert.NotNil(t, err)
	assert.Equal(t, false, b)

	r := record.Record{
		ID:   "1",
		UUID: uuid,
		File: "no-such-file",
	}
	p := stat.Proof{
		Total: 5,
		Count: 1,
	}
	bb := buffer.Buffer{}
	b, err = record.Skip(&bb, p, r)
	assert.Nil(t, err)
	assert.Equal(t, true, b)
	assert.Contains(t, bb.String(), "1 is missing")
}
