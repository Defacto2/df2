package archive

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/images"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
	"github.com/gabriel-vasile/mimetype"
	unarr "github.com/gen2brain/go-unarr"
)

type task struct {
	name string // filename
	size int64  // file size
	cont bool   // continue, don't scan anymore images
}

// var archives = []string{"zip", "tar.gz", "tar", "rar", "gz", "lzh", "lha", "cab", "arj", "arc", "7z"}

// Extract decompresses and parses an archive
func Extract(name string, uuid string) error {
	// create temp dir
	tempDir, err := ioutil.TempDir("", "extarc-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	// extract archive
	a, err := unarr.NewArchive(name)
	if err != nil {
		return err
	}
	defer a.Close()
	err = a.Extract(tempDir)
	if err != nil {
		return err
	}
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return err
	}
	th := task{name: "", size: 0, cont: false}
	tx := task{name: "", size: 0, cont: false}
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
			// todo copy file
			tx.name = fn
			tx.size = file.Size()
			tx.cont = true
		default:
			//fmt.Println(fmime.Extension())
		}
	}
	if n := th.name; n != "" {
		d := directories.Init(false)
		images.Generate(n, uuid, d)
	}
	if x := false; !x {
		dir(tempDir)
	}
	return nil
}

// NewExt replaces or appends the extension to a file name.
func NewExt(name string, extension string) string {
	e := filepath.Ext(name)
	if e == "" {
		return name + extension
	}
	fn := strings.TrimSuffix(name, e)
	return fn + extension
}

// Read returns a list of files within rar, tar, zip, 7z archives
func Read(name string) ([]string, error) {
	a, err := unarr.NewArchive(name)
	if err != nil {
		return nil, err
	}
	defer a.Close()
	list, err := a.List()
	if err != nil {
		return nil, err
	}
	return list, nil
}

func dir(name string) {
	files, err := ioutil.ReadDir(name)
	logs.Check(err)
	for _, file := range files {
		mime, err := mimetype.DetectFile(name + "/" + file.Name())
		logs.Log(err)
		fmt.Println(file.Name(), humanize.Bytes(uint64(file.Size())), mime)
	}
}
