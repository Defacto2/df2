package tf_test

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/text/internal/tf"
	"github.com/stretchr/testify/assert"
)

const (
	uuid          = "21cb94d3-ffc1-4055-8398-b7b4ed1e67e8"
	fileToExtract = "test.txt"
)

var (
	dzDir = filepath.Join("..", "..", "..", "..", "testdata", "demozoo")
)

func TestTextFile(t *testing.T) {
	t.Parallel()
	s := tf.TextFile{}
	assert.Equal(t, "(0)  0 B ", s.String())
	s = tf.TextFile{
		ID:   1,
		Name: "somefile.txt",
		Size: 512,
	}
	assert.Equal(t, "(1) somefile.txt 512 B ", s.String())
	assert.False(t, s.Archive())
	s.Name = "somefile.zip"
	assert.True(t, s.Archive())

	b, err := s.Exist(nil)
	assert.NotNil(t, err)
	assert.False(t, b)

	dir, err := directories.Init(conf.Defaults(), false)
	assert.Nil(t, err)
	b, err = s.Exist(&dir)
	assert.Nil(t, err)
	assert.False(t, b)
}

func TestTextFile_Extract(t *testing.T) {
	t.Parallel()
	s := tf.TextFile{}
	err := s.Extract(nil, nil)
	assert.NotNil(t, err)

	dir, err := directories.Init(conf.Defaults(), false)
	assert.Nil(t, err)
	err = s.Extract(io.Discard, &dir)
	assert.NotNil(t, err)

	dir.UUID = dzDir // overwrite the UUID to use testdata
	s = tf.TextFile{
		ID:     1,
		UUID:   uuid,
		Name:   "test.zip",
		Ext:    ".zip",
		Readme: sql.NullString{String: fileToExtract, Valid: true},
	}
	err = s.Extract(io.Discard, &dir)
	assert.Nil(t, err)
	// test.zip only contains 1 text file that should be removed after extraction
	defer os.Remove(filepath.Join(dzDir, uuid+".txt"))
}

func TestTextFile_ExtractedImgs(t *testing.T) {
	t.Parallel()
	s := tf.TextFile{}
	err := s.ExtractedImgs(nil, conf.Config{}, "")
	assert.NotNil(t, err)
	err = s.ExtractedImgs(io.Discard, conf.Defaults(), "")
	assert.NotNil(t, err)

	s = tf.TextFile{
		ID:     1,
		UUID:   uuid,
		Name:   "test.zip",
		Ext:    ".zip",
		Readme: sql.NullString{String: fileToExtract, Valid: true},
	}
	err = s.ExtractedImgs(io.Discard, conf.Defaults(), filepath.Join(dzDir, "extracted"))
	assert.Nil(t, err)
}

func TestTextFile_TextPNG(t *testing.T) {
	t.Parallel()
	s := tf.TextFile{}
	err := s.TextPNG(nil, conf.Config{}, 0, "")
	assert.NotNil(t, err)
	err = s.TextPNG(io.Discard, conf.Defaults(), 0, "")
	assert.NotNil(t, err)

	s = tf.TextFile{
		ID:     1,
		UUID:   uuid,
		Name:   "test.zip",
		Ext:    ".zip",
		Readme: sql.NullString{String: fileToExtract, Valid: true},
	}
	err = s.TextPNG(io.Discard, conf.Defaults(), 0, dzDir)
	assert.NotNil(t, err)
	// further tests can be done using the those created for img.Make()
}

func TestTextFile_Webp(t *testing.T) {
	t.Parallel()
	s := tf.TextFile{}
	i, err := s.WebP(nil, 0, "")
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)

	i, err = s.WebP(io.Discard, 0, "")
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)

	s = tf.TextFile{
		ID:     1,
		UUID:   uuid,
		Name:   "test.zip",
		Ext:    ".zip",
		Readme: sql.NullString{String: fileToExtract, Valid: true},
	}
	i, err = s.WebP(io.Discard, 0, dzDir)
	assert.Nil(t, err)
	assert.Equal(t, 1, i)
}
