package file

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var ErrPathEmpty = errors.New("path cannot be empty")

func WalkName(basepath, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("walkname: %w", ErrPathEmpty)
	}
	s := filepath.Dir(basepath)
	if os.IsPathSeparator(path[len(path)-1]) {
		s = basepath
	}
	name, err := filepath.Rel(s, path)
	if err != nil {
		return "", fmt.Errorf("walkname rel-path: %w", err)
	}
	return filepath.ToSlash(name), nil
}

// WriteTar saves the result of a fileWalk into a TAR writer.
// Source: cloudfoundry/archiver
// https://github.com/cloudfoundry/archiver/blob/master/compressor/write_tar.go
func WriteTar(absPath, filename string, tw *tar.Writer) error {
	stat, err := os.Lstat(absPath)
	if err != nil {
		return fmt.Errorf("writetar %q:%w", absPath, err)
	}
	var link string
	if stat.Mode()&os.ModeSymlink != 0 {
		if link, err = os.Readlink(absPath); err != nil {
			return fmt.Errorf("writetar mode:%w", err)
		}
	}
	head, err := tar.FileInfoHeader(stat, link)
	if err != nil {
		return fmt.Errorf("writetar header:%w", err)
	}
	if stat.IsDir() && !os.IsPathSeparator(filename[len(filename)-1]) {
		filename += "/"
	}
	name := filepath.ToSlash(filename)
	if head.Typeflag == tar.TypeReg && filename == "." {
		// archiving a single file
		name = filepath.ToSlash(filepath.Base(absPath))
	}
	head.Name = name
	if err := tw.WriteHeader(head); err != nil {
		return fmt.Errorf("writetar write header:%w", err)
	}
	if head.Typeflag == tar.TypeReg {
		return copyTar(absPath, tw)
	}
	return nil
}

func copyTar(name string, tw *tar.Writer) error {
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
