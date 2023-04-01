package file

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrEmptyPath = errors.New("path cannot be empty")
	ErrWriter    = errors.New("writer pointer cannot be nil")
)

func WalkName(basepath, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("file walkname: %w", ErrEmptyPath)
	}
	s := filepath.Dir(basepath)
	if os.IsPathSeparator(path[len(path)-1]) {
		s = basepath
	}
	name, err := filepath.Rel(s, path)
	if err != nil {
		return "", fmt.Errorf("file walkname rel: %w", err)
	}
	return filepath.ToSlash(name), nil
}

// Write saves the result of a fileWalk into a TAR writer.
// Source: cloudfoundry/archiver
// https://github.com/cloudfoundry/archiver/blob/master/compressor/write_tar.go
func Write(tw *tar.Writer, path, filename string) error {
	if tw == nil {
		return ErrWriter
	}
	if path == "" {
		return ErrEmptyPath
	}
	stat, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("write tar lstat %q:%w", path, err)
	}
	var link string
	if stat.Mode()&os.ModeSymlink != 0 {
		if link, err = os.Readlink(path); err != nil {
			return fmt.Errorf("write tar mode:%w", err)
		}
	}
	head, err := tar.FileInfoHeader(stat, link)
	if err != nil {
		return fmt.Errorf("write tar header:%w", err)
	}
	if stat.IsDir() && !os.IsPathSeparator(filename[len(filename)-1]) {
		filename += "/"
	}
	name := filepath.ToSlash(filename)
	if head.Typeflag == tar.TypeReg && filename == "." {
		// archiving a single file
		name = filepath.ToSlash(filepath.Base(path))
	}
	head.Name = name
	if err := tw.WriteHeader(head); err != nil {
		return fmt.Errorf("write tar write header:%w", err)
	}
	if head.Typeflag == tar.TypeReg {
		return copier(tw, path)
	}
	return nil
}

func copier(tw *tar.Writer, name string) error {
	if tw == nil {
		return ErrWriter
	}
	file, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("writetar open %q:%w", name, err)
	}
	defer file.Close()
	if _, err = io.Copy(tw, file); err != nil {
		return fmt.Errorf("writetar io.copy %q:%w", name, err)
	}
	return nil
}
