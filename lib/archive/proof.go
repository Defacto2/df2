package archive

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/images"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/gabriel-vasile/mimetype"
	unarr "github.com/gen2brain/go-unarr"
)

func extract(archive, tempDir string) error {
	// extract archive
	ua, err := unarr.NewArchive(archive)
	if err != nil {
		return err
	}
	defer ua.Close()
	if _, err = ua.Extract(tempDir); err != nil {
		return err
	}
	return ua.Close()
}

// Extract decompresses and parses an archive.
// uuid is used to rename the extracted assets such as image previews.
func Extract(archive, filename, uuid string) error {
	if err := database.CheckUUID(uuid); err != nil {
		return err
	}
	// create temp dir
	tempDir, err := ioutil.TempDir("", "extarc-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	if err := extract(archive, tempDir); err != nil {
		if err := extractr(archive, filename, tempDir); err != nil {
			return err
		}
	}
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return err
	}
	th, tx := taskInit(), taskInit()
	for _, file := range files {
		if th.cont && tx.cont {
			break
		}
		fn := path.Join(tempDir, file.Name())
		fmime, err := mimetype.DetectFile(fn)
		if err != nil {
			return err
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
	if n := th.name; n != "" {
		images.Generate(n, uuid, true)
	}
	if n := tx.name; n != "" {
		f := directories.Files(uuid)
		if _, err := FileMove(n, f.UUID+txt); err != nil {
			return err
		}
		logs.Print("  »txt")
	}
	if x := true; !x {
		if err := dir(tempDir); err != nil {
			return err
		}
	}
	return nil
}

type task struct {
	name string // filename
	size int64  // file size
	cont bool   // continue, don't scan anymore images
}

func taskInit() task {
	return task{name: "", size: 0, cont: false}
}

// TODO remove
func extErr(err error) error {
	return fmt.Errorf("archive extract: %v", err)
}
