package data_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/shrink/internal/data"
	"github.com/stretchr/testify/assert"
)

const (
	uuid = "29e0ca1f-c0a6-4b1a-b019-94a54243c093"
)

var (
	testdata = filepath.Join("..", "..", "..", "..", "tests", "uuid")
)

func TestMonth(t *testing.T) {
	m := data.Month("")
	assert.Equal(t, data.Months(0), m)
	m = data.Month("ja")
	assert.Equal(t, data.Months(0), m)
	m = data.Month("jan")
	assert.Equal(t, data.Months(1), m)
	m = data.Month("dec")
	assert.Equal(t, data.Months(12), m)

}

func TestApprovals_Approve(t *testing.T) {
	p := data.Preview
	err := p.Approve(nil)
	assert.NotNil(t, err)
	i := data.Preview
	err = i.Approve(nil)
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = p.Approve(db)
	assert.NotNil(t, err)
}

func TestInit(t *testing.T) {
	ds := string(filepath.Separator)
	d, err := filepath.Abs(ds)
	if err != nil {
		t.Error(err)
	}
	if err := data.Init(io.Discard, ""); err == nil {
		t.Errorf("Init() should have an error: %v", err)
	}
	if err := data.Init(io.Discard, d); err != nil {
		t.Errorf("Init() should have an error: %v", err)
	}
}

func TestSaveDir(t *testing.T) {
	s, err := data.SaveDir()
	assert.Nil(t, err)
	assert.NotEqual(t, "", s)
}

func TestApprovals_Store(t *testing.T) {
	path, err := data.Preview.Store(nil, "", "", false)
	assert.NotNil(t, err)
	assert.Equal(t, "", path)
	path, err = data.Preview.Store(io.Discard, testdata, "store-test", false)
	assert.Nil(t, err)
	assert.NotEqual(t, "", path)
	defer os.Remove(path)
}

func TestCompress(t *testing.T) {
	// archive file
	path := filepath.Join("..", "..", "..", "..", "tests", "empty")
	imgs := filepath.Join("..", "..", "..", "..", "tests", "images")
	d, err := filepath.Abs(path)
	if err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(d); os.IsNotExist(err) {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Error(err)
		}
		defer os.Remove(d)
	}
	tgz := filepath.Join(d, uuid)
	// use images as the content of the archive
	img, err := filepath.Abs(imgs)
	if err != nil {
		t.Error(err)
	}
	i1 := filepath.Join(img, "test_0x.png")
	i2 := filepath.Join(img, "test-clone.png")

	err = data.Compress(io.Discard, nil, "")
	assert.NotNil(t, err)
	err = data.Compress(io.Discard, nil, tgz)
	assert.Nil(t, err)
	err = data.Compress(io.Discard, []string{i1, i2}, tgz)
	assert.Nil(t, err)
	defer os.Remove(tgz)
}

func TestRemove(t *testing.T) {
	err := data.Remove(io.Discard, nil)
	assert.Nil(t, err)
	err = data.Remove(io.Discard, []string{"blahblahblah"})
	assert.NotNil(t, err)
}
