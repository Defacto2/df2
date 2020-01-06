package archive

import (
	"fmt"
	"io"
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

// Extract decompresses and parses a named archive.
// uuid is used to rename the extracted assets such as image previews.
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
			tx.name = fn
			tx.size = file.Size()
			tx.cont = true
		}
	}
	if n := th.name; n != "" {
		images.Generate(n, uuid)
	}
	if n := tx.name; n != "" {
		f := directories.Files(uuid)
		size, err := FileMove(n, f.UUID+".txt")
		logs.Check(err)
		print(fmt.Sprintf("  %s ⮕ ...%s.txt %s", logs.Y(), uuid[26:36], humanize.Bytes(uint64(size))))
	}
	if x := true; !x {
		dir(tempDir)
	}
	return nil
}

// FileMove copies a file to the destination and then deletes the source.
func FileMove(name string, dest string) (int64, error) {
	src, err := os.Open(name)
	if err != nil {
		return 0, err
	}
	defer src.Close()
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	i, err := io.Copy(file, src)
	if err != nil {
		return 0, err
	}
	err = os.Remove(name)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// NewExt swaps or appends the extension to a filename.
func NewExt(name string, extension string) string {
	e := filepath.Ext(name)
	if e == "" {
		return name + extension
	}
	fn := strings.TrimSuffix(name, e)
	return fn + extension
}

// Read returns a list of files within an rar, tar, zip or 7z archive.
// In the future I would like to add support for the following archives
// "tar.gz", "gz", "lzh", "lha", "cab", "arj", "arc".
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

// dir lists the content of a directory.
func dir(name string) {
	files, err := ioutil.ReadDir(name)
	logs.Check(err)
	for _, file := range files {
		mime, err := mimetype.DetectFile(name + "/" + file.Name())
		logs.Log(err)
		fmt.Println(file.Name(), humanize.Bytes(uint64(file.Size())), mime)
	}
}
