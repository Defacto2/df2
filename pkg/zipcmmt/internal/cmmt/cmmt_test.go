package cmmt_test

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/zipcmmt/internal/cmmt"
	"github.com/stretchr/testify/assert"
)

const (
	mockCmmt = "Test is some placeholder text."
	uniCmmt  = "Test is some \xCD\xB9 text."
	uuid     = "ef73b9dc-58b5-11ec-bf63-0242ac130002"
)

var path = filepath.Join("..", "..", "..", "..", "testdata", "uuid")

func TestZipfile_Exist(t *testing.T) {
	t.Parallel()
	z := cmmt.Zipfile{}
	b, err := z.Exist("")
	assert.NotNil(t, err)
	assert.False(t, b)

	z = cmmt.Zipfile{
		UUID: "foo",
	}
	b, err = z.Exist(path)
	assert.NotNil(t, err)
	assert.False(t, b)

	z = cmmt.Zipfile{
		UUID:      uuid,
		Overwrite: true,
	}
	b, err = z.Exist("")
	assert.NotNil(t, err)
	assert.False(t, b)

	b, err = z.Exist(path)
	assert.Nil(t, err)
	assert.True(t, b)
}

func TestZipfile_Save(t *testing.T) {
	t.Parallel()
	z := cmmt.Zipfile{}
	s, err := z.Save(nil, "")
	assert.NotNil(t, err)
	assert.Equal(t, "", s)

	z = cmmt.Zipfile{
		UUID: "foo",
	}
	s, err = z.Save(io.Discard, "bar")
	assert.NotNil(t, err)
	assert.Equal(t, "", s)

	z = cmmt.Zipfile{
		ID:   1,
		UUID: uuid,
	}
	s, err = z.Save(io.Discard, path)
	assert.Nil(t, err)
	assert.NotEqual(t, "", s)
}

func TestZipfile_Format(t *testing.T) {
	t.Parallel()
	z := cmmt.Zipfile{}
	bb, err := z.Format(nil)
	assert.NotNil(t, err)
	assert.Empty(t, bb)

	s := mockCmmt
	bb, err = z.Format(&s)
	assert.NotNil(t, err)
	assert.Empty(t, bb)

	z = cmmt.Zipfile{
		ID: 1,
	}
	bb, err = z.Format(&s)
	assert.Nil(t, err)
	assert.Contains(t, bb.String(), mockCmmt)

	s = ""
	bb, err = z.Format(&s)
	assert.Nil(t, err)
	assert.Equal(t, "", bb.String())

	z = cmmt.Zipfile{
		ID:    1,
		CP437: true,
	}
	s = uniCmmt
	bb, err = z.Format(&s)
	assert.Nil(t, err)
	assert.Contains(t, bb.String(), "Test is some ═╣ text.")
}
