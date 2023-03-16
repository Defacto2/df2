// Package directories interacts with the filepaths that hold files and assets.
package directories

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/directories/internal/create"
)

var ErrNoDir = errors.New("dir structure cannot be nil")

const (
	// Archives.
	z7  = ".7z"
	arc = ".arc"
	arj = ".arj"
	bz2 = ".bz2"
	cab = ".cab"
	gz  = ".gz"
	lha = ".lha"
	lzh = ".lzh"
	rar = ".rar"
	tar = ".tar"
	tgz = ".tar.gz"
	zip = ".zip"
)

// Dir is the collection of directories pointing to specific files.
type Dir struct {
	Img000 string // Img000 hold screen captures and previews.
	Img400 string // Img400 hold 400x400 squared thumbnails.
	Backup string // Backup archives or previously removed files.
	Emu    string // Emu are the DOSee emulation files.
	Base   string // Base or root directory path, the parent of these other subdirectories.
	UUID   string // UUID file downloads.
}

// Init initialises the subdirectories and UUID structure.
func Init(cfg configger.Config, create bool) Dir {
	var d Dir
	d.Img000 = cfg.Images
	d.Img400 = cfg.Thumbs
	d.Backup = cfg.Backups
	d.Emu = cfg.Emulator
	d.Base = cfg.WebRoot
	d.UUID = cfg.Downloads
	if create {
		if err := createDirectories(&d); err != nil {
			log.Print(err)
		}
		if err := PlaceHolders(&d); err != nil {
			log.Print(err)
		}
	}
	return d
}

// createDirectories generates a series of UUID subdirectories.
func createDirectories(dir *Dir) error {
	v := reflect.ValueOf(dir)
	// iterate through the D struct values
	for i := 0; i < v.NumField(); i++ {
		if d := fmt.Sprintf("%v", v.Field(i).Interface()); d != "" {
			if err := create.Dir(d); err != nil {
				return fmt.Errorf("create directory: %w", err)
			}
		}
	}
	return nil
}

// ArchiveExt checks that the named file uses a known archive extension.
func ArchiveExt(name string) bool {
	switch filepath.Ext(strings.ToLower(name)) {
	case z7, arc, arj, bz2, cab, gz, lha, lzh, rar, tar, tgz, zip:
		return true
	}
	return false
}

// Files initialises the full path filenames for a UUID.
func Files(cfg configger.Config, name string) Dir {
	dirs := Init(cfg, false)
	dirs.UUID = filepath.Join(dirs.UUID, name)
	dirs.Emu = filepath.Join(dirs.Emu, name)
	dirs.Img000 = filepath.Join(dirs.Img000, name)
	dirs.Img400 = filepath.Join(dirs.Img400, name)
	return dirs
}

// PlaceHolders generates a collection placeholder files in the UUID subdirectories.
func PlaceHolders(dir *Dir) error {
	if dir == nil {
		return fmt.Errorf("placeholder: %w", ErrNoDir)
	}
	const oneMB, halfMB, twoFiles, nineFiles = 1000000, 500000, 2, 9
	if err := create.Holders(dir.UUID, oneMB, nineFiles); err != nil {
		return fmt.Errorf("create uuid holders: %w", err)
	}
	if err := create.Holders(dir.Emu, oneMB, twoFiles); err != nil {
		return fmt.Errorf("create emu holders: %w", err)
	}
	if err := create.Holders(dir.Img000, oneMB, nineFiles); err != nil {
		return fmt.Errorf("create img000 holders: %w", err)
	}
	if err := create.Holders(dir.Img400, halfMB, nineFiles); err != nil {
		return fmt.Errorf("create img400 holders: %w", err)
	}
	return nil
}

// Size returns the number of counted files and their summed size as bytes.
func Size(root string) (int64, uint64, error) {
	var count int64
	var bytes uint64
	err := filepath.Walk(root, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			bytes += uint64(info.Size())
			count++
		}
		return err
	})
	return count, bytes, err
}

// Touch creates the named file.
func Touch(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}
