package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
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

// Image is an image object.
type Image struct {
	ID   uint
	UUID string
	Name string
	Ext  string
	Size int
}

func (i Image) String() string {
	return fmt.Sprintf("(%v) %v %v ",
		color.Primary.Sprint(i.ID), i.Name,
		color.Info.Sprint(humanize.Bytes(uint64(i.Size))))
}

func (i Image) IsExt() (ok bool) {
	switch filepath.Ext(strings.ToLower(i.Name)) {
	case gif, jpg, jpeg, _png, tif, tiff:
		return true
	}
	return false
}

func (i Image) IsDir(dir *directories.Dir) (ok bool) {
	dirs := [2]string{dir.Img000, dir.Img400}
	for _, path := range dirs {
		if _, err := os.Stat(filepath.Join(path, i.UUID+_png)); !os.IsNotExist(err) {
			return true
		}
	}
	return false
}
