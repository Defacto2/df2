package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

func Compress(files []string, buf io.Writer) error {
	w := tar.NewWriter(buf)
	defer w.Close()

	for _, file := range files {
		if err := add(w, file); err != nil {
			return fmt.Errorf("compress %s: %w", file, err)
		}
	}

	return nil
}

func Store(files []string, buf io.Writer) error {
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, file := range files {
		if err := add(tw, file); err != nil {
			return fmt.Errorf("store %s: %w", file, err)
		}
	}

	return nil
}

func add(tw *tar.Writer, filename string) error {
	file, err := os.Open(filename)
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
	header.Name = filename

	if err = tw.WriteHeader(header); err != nil {
		return fmt.Errorf("header: %w", err)
	}

	if _, err = io.Copy(tw, file); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	return nil
}
