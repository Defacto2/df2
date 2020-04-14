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

// Extract decompresses and parses a named archive.
// uuid is used to rename the extracted assets such as image previews.
func Extract(name, uuid string) error {
	if err := database.CheckUUID(uuid); err != nil {
		return extErr(err)
	}
	// create temp dir
	tempDir, err := ioutil.TempDir("", "extarc-")
	if err != nil {
		return extErr(err)
	}
	defer os.RemoveAll(tempDir)
	// extract archive
	a, err := unarr.NewArchive(name)
	if err != nil {
		return extErr(err)
	}
	defer a.Close()
	// TODO: use contents strings[] instead of scanning the directory
	_, err = a.Extract(tempDir)
	if err != nil {
		return extErr(err)
	}
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return extErr(err)
	}
	th := taskInit()
	tx := taskInit()
	for _, file := range files {
		if th.cont && tx.cont {
			break
		}
		fn := path.Join(tempDir, file.Name())
		fmime, err := mimetype.DetectFile(fn)
		if err != nil {
			return extErr(err)
		}
		switch fmime.Extension() {
		case ".bmp", ".gif", ".jpg", ".png", ".tiff", ".webp":
			if th.cont {
				continue
			}
			switch {
			case file.Size() > th.size:
				th.name = fn
				th.size = file.Size()
			}
		case ".txt":
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
		_, err := FileMove(n, f.UUID+".txt")
		logs.Check(err)
		print(fmt.Sprint("  »txt"))
	}
	if x := true; !x {
		dir(tempDir)
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

func extErr(err error) error {
	return fmt.Errorf("archive extract: %v", err)
}
