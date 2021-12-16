package task

import (
	"fmt"
	"io/ioutil"
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
func Run(tempDir string) (th, tx Task, err error) {
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return th, tx, fmt.Errorf("extract archive read tempdir %q: %w", tempDir, err)
	}
	th, tx = Init(), Init()
	for _, file := range files {
		if th.Cont && tx.Cont {
			break
		}
		fn := path.Join(tempDir, file.Name())
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
