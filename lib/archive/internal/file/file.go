package file

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
	"github.com/gabriel-vasile/mimetype"
)

var (
	ErrSameArgs = errors.New("name and dest cannot be the same")
)

const (
	CreateMode = 0o666
)

// Copy copies a file to the destination.
func Copy(name, dest string) (written int64, err error) {
	src, err := os.Open(name)
	if err != nil {
		return 0, fmt.Errorf("filecopy open %q: %w", name, err)
	}
	defer src.Close()
	dst, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, CreateMode)
	if err != nil {
		return 0, fmt.Errorf("filecopy dest %q: %w", dest, err)
	}
	defer dst.Close()
	if written, err = io.Copy(dst, src); err != nil {
		return 0, fmt.Errorf("filecopy io.copy: %w", err)
	}
	return written, dst.Close()
}

// Dir lists the content of a directory.
func Dir(name string) error {
	files, err := ioutil.ReadDir(name)
	if err != nil {
		return fmt.Errorf("dir read name %q: %w", name, err)
	}
	for _, f := range files {
		mime, err := mimetype.DetectFile(name + "/" + f.Name())
		if err != nil {
			return fmt.Errorf("dir mime failure on %q: %w", f, err)
		}
		logs.Println(f.Name(), humanize.Bytes(uint64(f.Size())), mime)
	}
	return nil
}

// Move copies a file to the destination and then deletes the source.
func Move(name, dest string) (written int64, err error) {
	if name == dest {
		return 0, fmt.Errorf("filemove: %w", ErrSameArgs)
	}
	if written, err = Copy(name, dest); err != nil {
		return 0, fmt.Errorf("filemove: %w", err)
	}
	if _, err = os.Stat(dest); os.IsNotExist(err) {
		return 0, fmt.Errorf("filemove dest: %w", err)
	}
	if err = os.Remove(name); err != nil {
		return 0, fmt.Errorf("filemove remove name %q: %w", name, err)
	}
	return written, nil
}
