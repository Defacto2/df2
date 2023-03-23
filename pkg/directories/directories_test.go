package directories_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	d, err := directories.Init(configger.Defaults(), false)
	assert.Nil(t, err)
	assert.NotEmpty(t, d)
	d, err = directories.Init(configger.MockDirs(), false)
	assert.Nil(t, err)
	assert.NotEmpty(t, d)
	d, err = directories.Init(configger.MockDirs(), true)
	assert.Nil(t, err)
	assert.NotEmpty(t, d)
	tmp := filepath.Join(os.TempDir(), "df2-mocker")
	defer os.RemoveAll(tmp)
}

func TestArchiveExt(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{}, false},
		{"no period", args{"arj"}, false},
		{"okay", args{".arj"}, true},
		{"caps", args{".ARJ"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := directories.ArchiveExt(tt.args.name); got != tt.want {
				t.Errorf("ArchiveExt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFiles(t *testing.T) {
	d, err := directories.Files(configger.Defaults(), "")
	assert.Nil(t, err)
	assert.NotEmpty(t, d)
	d, err = directories.Files(configger.Defaults(), "qwerty")
	assert.Nil(t, err)
	assert.Contains(t, d.Img000, "qwerty")
}

func TestSize(t *testing.T) {
	empty := filepath.Join("..", "..", "testdata", "empty")
	valid := filepath.Join("..", "..", "testdata", "demozoo")
	i, u := int64(0), uint64(0)

	c, b, err := directories.Size("")
	assert.NotNil(t, err)
	assert.Equal(t, i, c)
	assert.Equal(t, u, b)

	c, b, err = directories.Size("/dev/null/no-such-dir")
	assert.NotNil(t, err)
	assert.Equal(t, i, c)
	assert.Equal(t, u, b)

	if _, err := os.Stat(empty); errors.Is(err, fs.ErrNotExist) {
		if err := os.Mkdir(empty, 0o755); err != nil {
			assert.Nil(t, err)
		}
		defer os.Remove(empty)
	}
	c, b, err = directories.Size(empty)
	assert.Nil(t, err)
	assert.Equal(t, i, c)
	assert.Equal(t, u, b)

	c, b, err = directories.Size(valid)
	assert.Nil(t, err)
	assert.Equal(t, int64(18), c)
	assert.Equal(t, uint64(9602), b)
}

func TestTouch(t *testing.T) {
	tmp := filepath.Join(os.TempDir(), "directories-touch-test.tmp")
	err := directories.Touch(tmp)
	assert.Nil(t, err)
	defer os.Remove(tmp)
}
