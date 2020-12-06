package archive

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/gabriel-vasile/mimetype"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/images"
	"github.com/Defacto2/df2/lib/logs"
)

// Extract decompresses and parses an archive.
// uuid is used to rename the extracted assets such as image previews.
func Extract(archive, filename, uuid string) error {
	if err := database.CheckUUID(uuid); err != nil {
		return fmt.Errorf("extract archive uuid %q: %w", uuid, err)
	}
	// create temp dir
	tempDir, err := ioutil.TempDir("", "proofextract-")
	if err != nil {
		return fmt.Errorf("extract archive tempdir %q: %w", tempDir, err)
	}
	defer os.RemoveAll(tempDir)
	if err = Unarchiver(archive, filename, tempDir); err != nil {
		return fmt.Errorf("extract unarchiver: %w", err)
	}
	th, tx, err := tasks(tempDir)
	if err != nil {
		return err
	}
	if n := th.name; n != "" {
		if err := images.Generate(n, uuid, true); err != nil {
			return fmt.Errorf("extract archive generate img: %w", err)
		}
	}
	if n := tx.name; n != "" {
		f := directories.Files(uuid)
		if _, err := FileMove(n, f.UUID+txt); err != nil {
			return fmt.Errorf("extract archive filemove %q: %w", n, err)
		}
		logs.Print("  »txt")
	}
	if x := true; !x {
		if err := dir(tempDir); err != nil {
			return fmt.Errorf("extract archive dir %q: %w", tempDir, err)
		}
	}
	return nil
}

func tasks(tempDir string) (th task, tx task, err error) {
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return th, tx, fmt.Errorf("extract archive read tempdir %q: %w", tempDir, err)
	}
	th, tx = taskInit(), taskInit()
	for _, file := range files {
		if th.cont && tx.cont {
			break
		}
		fn := path.Join(tempDir, file.Name())
		fmime, err := mimetype.DetectFile(fn)
		if err != nil {
			return th, tx, fmt.Errorf("extract archive detect mime %q: %w", fn, err)
		}
		switch fmime.Extension() {
		case bmp, gif, jpg, png, tiff, webp:
			if th.cont {
				continue
			}
			if file.Size() > th.size {
				th.name = fn
				th.size = file.Size()
			}
		case txt:
			if tx.cont {
				continue
			}
			tx.name = fn
			tx.size = file.Size()
			tx.cont = true
		}
	}
	return th, tx, nil
}

type task struct {
	name string // filename
	size int64  // file size
	cont bool   // continue, don't scan anymore images
}

func taskInit() task {
	return task{name: "", size: 0, cont: false}
}
