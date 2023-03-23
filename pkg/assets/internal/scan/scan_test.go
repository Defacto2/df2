package scan_test

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/assets"
	"github.com/Defacto2/df2/pkg/assets/internal/scan"
	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/stretchr/testify/assert"
)

// createTempDir creates a temporary directory and copies some test images into it.
// dir is the absolute directory path while sum is the sum total of bytes copied.
// Returns the sum of bytes written and the created directory path.
func createTempDir() (int64, string, error) {
	// make dir
	dir, err := os.MkdirTemp(os.TempDir(), "test-addtardir")
	if err != nil {
		return 0, "", err
	}
	// copy files
	path := filepath.Join("..", "..", "..", "..", "testdata", "images")
	src, err := filepath.Abs(path)
	if err != nil {
		return 0, dir, err
	}
	imgs := []string{"test.gif", "test.png", "test.jpg"}
	done, sum := make(chan error), int64(0)
	for _, f := range imgs {
		go func(f string) {
			sum, err = archive.Copy(filepath.Join(src, f), filepath.Join(dir, f))
			if err != nil {
				done <- err
			}
			done <- nil
		}(f)
	}
	done1, done2, done3 := <-done, <-done, <-done
	if done1 != nil {
		return 0, dir, done1
	}
	if done2 != nil {
		return 0, dir, done2
	}
	if done3 != nil {
		return 0, dir, done3
	}
	return sum, dir, nil
}

func TestBackup(t *testing.T) {
	_, tmp, err := createTempDir()
	assert.Nil(t, err)
	defer os.RemoveAll(tmp)

	entries, err := os.ReadDir(tmp)
	assert.Nil(t, err)
	list := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		i, err := entry.Info()
		assert.Nil(t, err)
		list = append(list, i)
	}

	d, err := directories.Init(configger.Defaults(), false)
	assert.Nil(t, err)
	d.Backup = os.TempDir() // test override

	err = scan.Backup(nil, nil, nil, nil, nil)
	assert.NotNil(t, err)

	skip, err := scan.Skip("", &d)
	assert.Nil(t, err)
	err = scan.Backup(io.Discard, skip, nil, nil, nil)
	assert.NotNil(t, err)
	err = scan.Backup(io.Discard, skip, &list, nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	_, uuids, err := assets.CreateUUIDMap(db)
	assert.Nil(t, err)
	s := scan.Scan{
		Path:   tmp,
		Delete: false,
		Human:  true,
		IDs:    uuids,
	}
	err = scan.Backup(io.Discard, skip, &list, &s, nil)
	assert.NotNil(t, err)
	err = scan.Backup(io.Discard, skip, &list, &s, &d)
	assert.Nil(t, err)
}
