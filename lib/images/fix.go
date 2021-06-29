package images

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/dustin/go-humanize" //nolint:misspell
	"github.com/gookit/color"       //nolint:misspell
)

const (
	gif  = ".gif"
	jpg  = ".jpg"
	jpeg = ".jpeg"
	_png = ".png"
	tif  = ".tif"
	tiff = ".tiff"
	webp = ".webp"
)

// imageFile is an image object.
type imageFile struct {
	ID   uint
	UUID string
	Name string
	Ext  string
	Size int
}

func (i imageFile) String() string {
	return fmt.Sprintf("(%v) %v %v ",
		color.Primary.Sprint(i.ID), i.Name,
		color.Info.Sprint(humanize.Bytes(uint64(i.Size))))
}

// Fix generates any missing assets from downloads that are images.
func Fix(simulate bool) error {
	dir := directories.Init(false)
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(`SELECT id, uuid, filename, filesize FROM files WHERE platform="image" ORDER BY id ASC`)
	if err != nil {
		return fmt.Errorf("images fix query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("images fix rows: %w", rows.Err())
	}
	defer rows.Close()
	c := 0
	for rows.Next() {
		var img imageFile
		if err = rows.Scan(&img.ID, &img.UUID, &img.Name, &img.Size); err != nil {
			return fmt.Errorf("images fix rows scan: %w", err)
		}
		if directories.ArchiveExt(img.Name) {
			continue
		}
		if !img.valid(&dir) {
			c++
			logs.Printf("%d. %v", c, img)
			if _, err := os.Stat(filepath.Join(dir.UUID, img.UUID)); os.IsNotExist(err) {
				logs.Printf("%s\n", str.X())
				continue
			} else if err != nil {
				return fmt.Errorf("images fix stat: %w", err)
			}
			if simulate {
				logs.Printf("%s\n", color.Question.Sprint("?"))
				continue
			}
			if err := Generate(filepath.Join(dir.UUID, img.UUID), img.UUID, false); err != nil {
				return fmt.Errorf("images fix generate: %w", err)
			}
			logs.Print("\n")
		}
	}
	if simulate && c > 0 {
		logs.Simulate()
	} else if c == 0 {
		logs.Println("everything is okay, there is nothing to do")
	}
	return nil
}

func (i imageFile) ext() (ok bool) {
	switch filepath.Ext(strings.ToLower(i.Name)) {
	case gif, jpg, jpeg, _png, tif, tiff:
		return true
	}
	return false
}

func (i imageFile) valid(dir *directories.Dir) (ok bool) {
	dirs := [3]string{dir.Img000, dir.Img400}
	for _, path := range dirs {
		if _, err := os.Stat(filepath.Join(path, i.UUID+_png)); !os.IsNotExist(err) {
			return true
		}
	}
	return false
}
