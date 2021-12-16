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

func WalkName(basepath, path string) (name string, err error) {
	if path == "" {
		return "", fmt.Errorf("walkname: %w", ErrPathEmpty)
	}
	if os.IsPathSeparator(path[len(path)-1]) {
		name, err = filepath.Rel(basepath, path)
	} else {
		name, err = filepath.Rel(filepath.Dir(basepath), path)
	}
	if err != nil {
		return "", fmt.Errorf("walkname rel-path: %w", err)
	}
	return filepath.ToSlash(name), nil
}

// writeTar saves the result of a fileWalk into a TAR writer.
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
	if head.Typeflag == tar.TypeReg && filename == "." {
		// archiving a single file
		head.Name = filepath.ToSlash(filepath.Base(absPath))
	} else {
		head.Name = filepath.ToSlash(filename)
	}
	if err := tw.WriteHeader(head); err != nil {
		return fmt.Errorf("writetar write header:%w", err)
	}
	if head.Typeflag == tar.TypeReg {
		file, err := os.Open(absPath)
		if err != nil {
			return fmt.Errorf("writetar open %q:%w", absPath, err)
		}
		defer file.Close()
		if _, err = io.Copy(tw, file); err != nil {
			return fmt.Errorf("writetar io.copy %q:%w", absPath, err)
		}
	}
	return nil
}
