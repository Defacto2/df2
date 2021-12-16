package file

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
	gap "github.com/muesli/go-app-paths"
)

const (
	gif  = ".gif"
	jpg  = ".jpg"
	jpeg = ".jpeg"
	_png = ".png"
	tif  = ".tif"
	tiff = ".tiff"
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

// Check the existence of the named file and
// confirm it is not a directory or zero-byte file.
func Check(name string, err error) bool {
	if err != nil {
		return false
	}
	s, err := os.Stat(name)
	if err != nil {
		return false
	}
	if s.IsDir() || s.Size() < 1 {
		return false
	}
	return true
}

// Remove the named file, only when confirm is true.
func Remove(confirm bool, name string) error {
	if !confirm {
		return nil
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("remove %q: %w", name, err)
	}
	return nil
}

func RemoveWebP(name string) error {
	s, err := os.Stat(name)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("removewebp stat: %w", err)
	}
	if s.Size() == 0 {
		if err := os.Remove(name); err != nil {
			return fmt.Errorf("removewebp: %w", err)
		}
	}
	return nil
}

// Vendor is the absolute path to store webpbin vendor downloads.
func Vendor() string {
	fp, err := gap.NewScope(gap.User, logs.GapUser).CacheDir()
	if err != nil {
		h, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(fmt.Errorf("vendorPath userhomedir: %w", err))
		}
		return path.Join(h, ".vendor/df2")
	}
	return fp
}
