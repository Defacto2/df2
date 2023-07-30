// Package file handles the images as files.
package file

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
	gap "github.com/muesli/go-app-paths"
)

var ErrPointer = errors.New("pointer value cannot be nil")

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
	return fmt.Sprintf("%v  %v  %v ",
		color.Primary.Sprint(i.ID),
		color.Info.Sprint(humanize.Bytes(uint64(i.Size))),
		i.Name,
	)
}

func (i Image) IsExt() bool {
	switch filepath.Ext(strings.ToLower(i.Name)) {
	case gif, jpg, jpeg, _png, tif, tiff:
		return true
	}
	return false
}

func (i Image) IsDir(dir *directories.Dir) (bool, error) {
	if dir == nil {
		return false, fmt.Errorf("dir %w", ErrPointer)
	}
	dirs := [2]string{dir.Img000, dir.Img400}
	for _, path := range dirs {
		_, err := os.Stat(filepath.Join(path, i.UUID+_png))
		if !errors.Is(err, fs.ErrNotExist) {
			return true, nil
		}
	}
	return false, nil
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

// Remove0byte removes the named file but only if it is 0 bytes in size.
func Remove0byte(name string) error {
	s, err := os.Stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("remove 0byte stat: %w", err)
	}
	if s.Size() == 0 {
		if err := os.Remove(name); err != nil {
			return fmt.Errorf("remove 0byte: %w", err)
		}
	}
	return nil
}

// Vendor is the absolute path to store webpbin vendor downloads.
func Vendor() string {
	fp, err := gap.NewScope(gap.User, conf.GapUser).CacheDir()
	if err != nil {
		h, err := os.UserHomeDir()
		if err != nil {
			log.Print(fmt.Errorf("vendorPath userhomedir: %w", err))
		}
		return path.Join(h, ".vendor/df2")
	}
	return fp
}
