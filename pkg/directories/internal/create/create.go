package create

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	m "math/rand"
	"os"
	"path/filepath"
	"time"
)

var (
	ErrPathIsFile = errors.New("path already exist as a file")
	ErrPrefix     = errors.New("invalid prefix value, it must be between 0 - 9")
)

const (
	dirMode  fs.FileMode = 0o755
	fileMode fs.FileMode = 0o644

	// random characters used by randString().
	random = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321 .!?"
)

// Dir creates a UUID subdirectory in the directory path.
func Dir(path string) error {
	src, err := os.Stat(path)
	if os.IsNotExist(err) {
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
	m.Seed(time.Now().UnixNano())
	r, err := RandString(size)
	if err != nil {
		return fmt.Errorf("create holder file: %w", err)
	}
	text := []byte(r)
	if err := os.WriteFile(fn, text, fileMode); err != nil {
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

// RandString generates a random string of n characters.
func RandString(n int) (string, error) {
	s, r := make([]rune, n), []rune(random)
	for i := range s {
		p, err := rand.Prime(rand.Reader, len(r))
		if err != nil {
			return "", fmt.Errorf("random string n %d: %w", n, err)
		}
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s), nil
}
