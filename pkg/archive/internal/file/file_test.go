package file_test

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/archive/internal/file"
	"github.com/stretchr/testify/assert"
)

func join(s string) string {
	x := filepath.Join("text", s)
	return testDir(x)
}

func testDir(name string) string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "..", "..", "tests", name)
}

func TestAdd(t *testing.T) {
	err := file.Add(nil, "")
	assert.NotNil(t, err)
	err = file.Add(nil, testDir("file-does-not-exists"))
	assert.NotNil(t, err)

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	err = file.Add(tw, testDir("text/test.png"))
	assert.Nil(t, err)
	err = file.Add(tw, testDir("text/test.txt"))
	assert.Nil(t, err)
}

func TestDir(t *testing.T) {
	err := file.Dir(nil, "")
	assert.NotNil(t, err)
	err = file.Dir(io.Discard, testDir(""))
	assert.Nil(t, err)
}

func TestMove(t *testing.T) {
	i, err := file.Move("", "")
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)
	i, err = file.Move(join("test.txt"), "")
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)
	i, err = file.Move("", join("test.txt"))
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)

	src := filepath.Join(os.TempDir(), "test-file-move.txt")
	i, err = file.Copy(join("test.txt"), src)
	assert.Nil(t, err)
	assert.Equal(t, int64(12), i)

	i, err = file.Copy(src, src+"~")
	assert.Nil(t, err)
	assert.Equal(t, int64(12), i)
	defer os.Remove(src + "~")
}

func TestCopy(t *testing.T) {
	i, err := file.Copy("", "")
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)

	i, err = file.Copy(join("test.txt"), "")
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)
	i, err = file.Copy("", join("test.txt"))
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)
	i, err = file.Copy(join("test.txt"), join("test.txt"))
	assert.Nil(t, err)
	assert.Equal(t, int64(12), i)
	i, err = file.Copy(join("test.txt"), join("test.txt~"))
	assert.Nil(t, err)
	assert.Equal(t, int64(12), i)
	defer os.Remove(join("test.txt~"))

	i, err = file.Copy(join("test.txt"), os.TempDir())
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), i)
}
