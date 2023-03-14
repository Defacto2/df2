package file

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/gabriel-vasile/mimetype"
)

var ErrSameArgs = errors.New("name and dest cannot be the same")

const (
	CreateMode = 0o666
)

// Add a file to the tar writer.
func Add(tw *tar.Writer, src string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return fmt.Errorf("info: %w", err)
	}
	// https://golang.org/src/archive/tar/common.go?#L626
	header.Name = src

	if err = tw.WriteHeader(header); err != nil {
		return fmt.Errorf("header: %w", err)
	}

	if _, err = io.Copy(tw, file); err != nil {
		return fmt.Errorf("copy: %w", err)
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

// Dir lists the content of a directory.
func Dir(w io.Writer, name string) error {
	files, err := os.ReadDir(name)
	if err != nil {
		return fmt.Errorf("dir read name %q: %w", name, err)
	}
	for _, file := range files {
		f, err := file.Info()
		if err != nil {
			return fmt.Errorf("dir failure with %q: %w", f, err)
		}
		mime, err := mimetype.DetectFile(name + "/" + f.Name())
		if err != nil {
			return fmt.Errorf("dir mime failure on %q: %w", f, err)
		}
		fmt.Fprintln(w, f.Name(), humanize.Bytes(uint64(f.Size())), mime)
	}
	return nil
}

// Move copies a file to the destination and then deletes the source.
func Move(name, dest string) (int64, error) {
	if name == dest {
		return 0, fmt.Errorf("filemove: %w", ErrSameArgs)
	}
	written, err := Copy(name, dest)
	if err != nil {
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
