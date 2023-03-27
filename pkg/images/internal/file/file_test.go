package file_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/images"
	"github.com/Defacto2/df2/pkg/images/internal/file"
	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"
)

func testDir() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "..", "..", "testdata", "images")
}

func TestImage(t *testing.T) {
	t.Parallel()
	color.Enable = false
	i := file.Image{}
	assert.Equal(t, "(0)  0 B ", i.String())
	i = file.Image{
		ID:   9345,
		Name: "myphoto.jpg",
		Size: 654174,
	}
	assert.Equal(t, "(9345) myphoto.jpg 654 kB ", i.String())
	assert.Equal(t, true, i.IsExt())
	b, err := i.IsDir(nil)
	assert.NotNil(t, err)
	assert.Equal(t, false, b)
	d := directories.Dir{}
	b, err = i.IsDir(&d)
	assert.Nil(t, err)
	assert.Equal(t, false, b)
}

func TestIsExt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		i      file.Image
		wantOk bool
	}{
		{"empty", file.Image{}, false},
		{"text", file.Image{Name: "some.txt"}, false},
		{"png", file.Image{Name: "some.png"}, true},
		{"jpeg", file.Image{Name: "some other.jpeg"}, true},
		{"jpeg", file.Image{Name: "some.other.jpeg"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOk := tt.i.IsExt(); gotOk != tt.wantOk {
				t.Errorf("IsExt() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	t.Parallel()
	b := file.Check("", nil)
	assert.Equal(t, false, b)
	err := errors.New("check err")
	b = file.Check("", err)
	assert.Equal(t, false, b)
	b = file.Check(filepath.Join(testDir(), "test.iff"), nil)
	assert.Equal(t, true, b)
}

func TestRemove(t *testing.T) {
	t.Parallel()
	err := file.Remove(false, "")
	assert.Nil(t, err)
	err = file.Remove(true, "")
	assert.NotNil(t, err)

	src := filepath.Join(testDir(), "test.iff")
	dst := filepath.Join(os.TempDir(), "test-check-test.iff")
	err = images.Copy(dst, src)
	assert.Nil(t, err)
	err = file.Remove(true, dst)
	assert.Nil(t, err)

	src = filepath.Join(testDir(), "test-0byte.000")
	err = directories.Touch(src)
	assert.Nil(t, err)
	err = file.Remove0byte(src)
	assert.Nil(t, err)
}

func TestVendor(t *testing.T) {
	t.Parallel()
	s := file.Vendor()
	assert.NotEqual(t, "", s)
}
