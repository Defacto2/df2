package recd_test

import (
	"bytes"
	"database/sql"
	"io"
	"os"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/database/internal/recd"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

func TestRecord_String(t *testing.T) {
	t.Parallel()
	r := recd.Record{}
	assert.NotEqual(t, "", r.String())
}

func TestRecord_Approve(t *testing.T) {
	t.Parallel()
	r := recd.Record{}
	err := r.Approve(nil)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = r.Approve(db)
	assert.Nil(t, err)

	r = recd.Record{
		ID: 1,
	}
	err = r.Approve(db)
	assert.Nil(t, err)

	i := r.AutoID("1")
	assert.Equal(t, uint(1), i)
	err = r.Approve(db)
	assert.Nil(t, err)
}

func TestRecord_Check(t *testing.T) {
	t.Parallel()
	r := recd.Record{}
	b, err := r.Check(nil, "", nil, nil)
	assert.NotNil(t, err)
	assert.False(t, b)

	dir, err := directories.Init(conf.Defaults(), false)
	assert.Nil(t, err)
	bb := &bytes.Buffer{}
	b, err = r.Check(bb, "", nil, &dir)
	assert.NotNil(t, err)
	assert.False(t, b)
	vals := make([]sql.RawBytes, 15)
	b, err = r.Check(bb, "", vals, &dir)
	assert.Nil(t, err)
	assert.False(t, b)
}

func TestRecord_Checks(t *testing.T) {
	t.Parallel()
	r := recd.Record{}
	b := r.CheckDownload(io.Discard, "", "")
	assert.False(t, b)

	r = recd.Record{
		Filename: "test.zip",
	}
	r.CheckFileContent("")
	assert.False(t, b)

	b = r.CheckFileSize(internal.RandStr)
	assert.False(t, b)

	b = r.CheckFileSize("1024")
	assert.True(t, b)

	b = r.CheckImage(internal.RandStr)
	assert.False(t, b)
	b = r.CheckImage(internal.TestImg(4))
	assert.True(t, b)

	b = r.RecoverDownload(nil, "", "")
	assert.False(t, b)

	r = recd.Record{Filename: internal.Zip}
	f, err := os.CreateTemp(os.TempDir(), "recover-download")
	assert.Nil(t, err)
	defer f.Close()
	b = r.RecoverDownload(io.Discard, internal.TestArchives(4), f.Name())
	assert.True(t, b)
	defer os.Remove(f.Name())
}

func TestRecord_Summary(t *testing.T) {
	t.Parallel()
	r := recd.Record{}
	r.Summary(nil, 0)
	r.Summary(nil, 1)
}

func TestNewApprove(t *testing.T) {
	t.Parallel()
	b, err := recd.NewApprove(nil)
	assert.NotNil(t, err)
	assert.False(t, b)
	raw := make([]sql.RawBytes, 4)
	b, err = recd.NewApprove(raw)
	assert.Nil(t, err)
	assert.False(t, b)
	now := time.Now().Format(time.RFC3339)
	raw[2] = []byte(now)
	raw[3] = []byte(now)
	b, err = recd.NewApprove(raw)
	assert.Nil(t, err)
	assert.True(t, b)
}

func TestVerbose(t *testing.T) {
	t.Parallel()
	recd.Verbose(io.Discard, false, "true")
	bb := &bytes.Buffer{}
	recd.Verbose(bb, true, internal.RandStr)
	assert.Contains(t, bb.String(), internal.RandStr)
}

func TestQueries(t *testing.T) {
	t.Parallel()
	err := recd.Queries(nil, nil, conf.Config{}, false)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = recd.Queries(db, io.Discard, conf.Defaults(), false)
	assert.Nil(t, err)
}

func TestCheckGroups(t *testing.T) {
	t.Parallel()
	r := recd.Record{}
	b := r.CheckGroups("", "")
	assert.False(t, b)
	b = r.CheckGroups("CHANGEME", "")
	assert.False(t, b)
	b = r.CheckGroups("", "CHANGEME")
	assert.False(t, b)
	b = r.CheckGroups("A group", "")
	assert.True(t, b)
	b = r.CheckGroups("", "A group")
	assert.True(t, b)
}

func TestValid(t *testing.T) {
	t.Parallel()
	b, err := recd.Valid(nil, nil)
	assert.NotNil(t, err)
	assert.False(t, b)

	new, offset := []byte("2006-01-02T15:04:06Z"), []byte("2006-01-02T15:04:05Z")
	b, err = recd.Valid(new, new)
	assert.Nil(t, err)
	assert.True(t, b)
	b, err = recd.Valid(new, offset)
	assert.Nil(t, err)
	assert.True(t, b)
	new, offset = []byte("2026-01-02T15:04:06Z"), []byte("2006-01-02T15:04:05Z")
	b, err = recd.Valid(new, offset)
	assert.Nil(t, err)
	assert.False(t, b)
	b, err = recd.Valid(offset, new)
	assert.Nil(t, err)
	assert.False(t, b)
}

func TestReverseInt(t *testing.T) {
	t.Parallel()
	i, err := recd.ReverseInt(0)
	assert.Nil(t, err)
	assert.Equal(t, 0, i)
	i, err = recd.ReverseInt(12345)
	assert.Nil(t, err)
	assert.Equal(t, 54321, i)
	i, err = recd.ReverseInt(555)
	assert.Nil(t, err)
	assert.Equal(t, 555, i)
	i, err = recd.ReverseInt(662211)
	assert.Nil(t, err)
	assert.Equal(t, 112266, i)
}
