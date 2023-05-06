// Package task handles the scanning of directories for release proofs.
package task

import (
	"fmt"
	"os"
	"path"

	"github.com/gabriel-vasile/mimetype"
)

const (
	bmp  = ".bmp"
	gif  = ".gif"
	jpg  = ".jpg"
	png  = ".png"
	tiff = ".tiff"
	txt  = ".txt"
	webp = ".webp"
)

// Task for fetching both text and image proofs.
type Task struct {
	Name string // Name of file.
	Size int64  // Size of the file in bytes.
	Cont bool   // Continue, stops the scan for anymore similar file types.
}

// Init initializes the task.
func Init() Task {
	return Task{
		Name: "",
		Size: 0,
		Cont: false,
	}
}

// Run a scan for proofs in the provided temp directory.
func Run(tempDir string) ( //nolint:cyclop,nonamedreturns
	th Task, tx Task, err error,
) {
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return th, tx, fmt.Errorf("extract archive read tempdir %q: %w", tempDir, err)
	}
	th, tx = Init(), Init()
	for _, entry := range entries {
		if th.Cont && tx.Cont {
			break
		}
		if entry.IsDir() {
			continue
		}
		file, err := entry.Info()
		if err != nil {
			fmt.Fprintf(os.Stdout, "extract archive entry error: %s\n", err)
			continue
		}
		fn := path.Join(tempDir, entry.Name())
		fmime, err := mimetype.DetectFile(fn)
		if err != nil {
			return th, tx, fmt.Errorf("extract archive detect mime %q: %w", fn, err)
		}
		switch fmime.Extension() {
		case bmp, gif, jpg, png, tiff, webp:
			if th.Cont {
				continue
			}
			if file.Size() > th.Size {
				th.Name = fn
				th.Size = file.Size()
			}
		case txt:
			if tx.Cont {
				continue
			}
			tx.Name = fn
			tx.Size = file.Size()
			tx.Cont = true
		}
	}
	return th, tx, nil
}
