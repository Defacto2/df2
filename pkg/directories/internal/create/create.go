// Package create handles the making of directories.
package create

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var (
	ErrPathIsFile = errors.New("path already exist as a file")
	ErrPrefix     = errors.New("invalid prefix value, it must be between 0 - 9")
)

const (
	dirMode  fs.FileMode = 0o755
	fileMode fs.FileMode = 0o644
)

// MkDir creates a UUID subdirectory in the directory path.
func MkDir(path string) error {
	src, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		if err = os.MkdirAll(path, dirMode); err != nil {
			return fmt.Errorf("create directory mkdir %q: %w", path, err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("create directory stat %q: %w", path, err)
	}
	if src.Mode().IsRegular() {
		return fmt.Errorf("create directory %q: %w", path, ErrPathIsFile)
	}
	return nil
}

// Holder generates a placeholder file filled with random text in the given directory path,
// the size of the file determines the number of random characters and the prefix is a digit between
// 0 and 9 is appended to the filename.
func Holder(path string, size int, prefix uint) error {
	const max = 9
	if prefix > max {
		return fmt.Errorf("create holder file prefix=%d: %w", prefix, ErrPrefix)
	}
	name := fmt.Sprintf("00000000-0000-0000-0000-00000000000%v", prefix)
	fn := filepath.Join(path, name)
	if _, err := os.Stat(fn); err == nil {
		return nil // don't overwrite existing files
	}
	if err := os.WriteFile(fn, Zeros(size), fileMode); err != nil {
		return fmt.Errorf("write create holder file %q: %w", fn, err)
	}
	return nil
}

// Holders generates a number of placeholder files in the given directory path.
func Holders(path string, size int, count uint) error {
	const max = 9
	if count > max {
		return fmt.Errorf("create holder files number=%d: %w", count, ErrPrefix)
	}
	for i := uint(0); i <= count; i++ {
		if err := Holder(path, size, i); err != nil {
			return fmt.Errorf("create holder files: %w", err)
		}
	}
	return nil
}

// Zeros generates a string of n x 0 characters.
func Zeros(n int) []byte {
	return bytes.Repeat([]byte("0"), n)
}
