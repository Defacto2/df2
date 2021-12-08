package task

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/gabriel-vasile/mimetype"
)

const (
	bat  = ".bat"
	bmp  = ".bmp"
	com  = ".com"
	diz  = ".diz"
	exe  = ".exe"
	gif  = ".gif"
	jpg  = ".jpg"
	nfo  = ".nfo"
	png  = ".png"
	tiff = ".tiff"
	txt  = ".txt"
	webp = ".webp"
)

type Task struct {
	Name string // filename
	Size int64  // file size
	Cont bool   // continue, don't scan anymore images
}

func Init() Task {
	return Task{
		Name: "",
		Size: 0,
		Cont: false}
}

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
