package content_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/archive/internal/content"
	"github.com/stretchr/testify/assert"
)

func testDir(name string) string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "..", "..", "testdata", name)
}
func TestFile_MIME(t *testing.T) {
	t.Parallel()
	f := content.File{}
	err := f.MIME()
	assert.NotNil(t, err)
	f = content.File{Path: testDir("demozoo/test.tar.bz2")}
	err = f.MIME()
	assert.Nil(t, err)
	assert.Equal(t, "application/x-bzip2", f.Mime.String())
	f = content.File{Path: testDir("demozoo/test/test.txt")}
	err = f.MIME()
	assert.Nil(t, err)
	assert.Equal(t, "text/plain; charset=utf-8", f.Mime.String())
}

func TestFile_Scan(t *testing.T) {
	t.Parallel()
	f := content.File{}
	st, err := os.Stat(testDir("demozoo/test/test.txt"))
	assert.Nil(t, err)
	f.Scan(st)
	assert.Equal(t, ".txt", f.Ext)
	assert.Equal(t, true, f.Textfile)
	assert.Equal(t, "test.txt (.txt extension)", f.String())
}
