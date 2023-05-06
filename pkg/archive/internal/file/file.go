// Package file handles common file system operations such as move, copy and
// dir.
package file

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/gabriel-vasile/mimetype"
)

var (
	ErrSameArgs = errors.New("name and dest cannot be the same")
	ErrWriter   = errors.New("writer must be a file object")
)

const (
	CreateMode = 0o666
)

// Add a file to the tar writer.
func Add(tw *tar.Writer, src string) error {
	if tw == nil {
		return fmt.Errorf("archive add file to tar: %w", ErrWriter)
	}
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open archive add file to tar: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat archive add file to tar: %w", err)
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return fmt.Errorf("info archive add file to tar: %w", err)
	}
	// https://golang.org/src/archive/tar/common.go?#L626
	header.Name = src

	if err = tw.WriteHeader(header); err != nil {
		return fmt.Errorf("header archive add file to tar: %w", err)
	}

	if _, err = io.Copy(tw, file); err != nil {
		return fmt.Errorf("copy archive add file to tar: %w", err)
	}

	return nil
}

// Copy copies a file to the destination.
func Copy(name, dest string) (int64, error) {
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
	written, err := io.Copy(dst, src)
	if err != nil {
		return 0, fmt.Errorf("filecopy io.copy: %w", err)
	}
	return written, dst.Close()
}

// Dir writes the content of a directory to the writer.
func Dir(w io.Writer, name string) error {
	if w == nil {
		w = io.Discard
	}
	files, err := os.ReadDir(name)
	if err != nil {
		return fmt.Errorf("dir read name %q: %w", name, err)
	}
	for _, f := range files {
		fi, err := f.Info()
		if err != nil {
			return fmt.Errorf("dir failure with %q: %w", fi, err)
		}
		if fi.IsDir() {
			continue
		}
		mime, err := mimetype.DetectFile(name + "/" + fi.Name())
		if err != nil {
			return fmt.Errorf("dir mime failure on %q: %w", fi, err)
		}
		fmt.Fprintln(w, fi.Name(), humanize.Bytes(uint64(fi.Size())), mime)
	}
	return nil
}

// Move copies a file to the destination and then deletes the source.
func Move(name, dest string) (int64, error) {
	if name == dest {
		return 0, fmt.Errorf("file move: %w", ErrSameArgs)
	}
	b, err := Copy(name, dest)
	if err != nil {
		return 0, fmt.Errorf("file move copy: %w", err)
	}
	if _, err = os.Stat(dest); errors.Is(err, fs.ErrNotExist) {
		return 0, fmt.Errorf("file move: %s: %w", dest, err)
	}
	if err = os.Remove(name); err != nil {
		return 0, fmt.Errorf("file move remove %q: %w", name, err)
	}
	return b, nil
}
