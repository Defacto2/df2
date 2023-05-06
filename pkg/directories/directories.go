// Package directories interacts with the filepaths that hold both the user files
// for downloads and website assets.
package directories

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/directories/internal/create"
)

var (
	ErrDir = errors.New("directory cannot be an empty value")
	ErrNil = errors.New("directories structure pointer cannot be nil")
)

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
	Img000 string // Img000 contain record screenshots and previews.
	Img400 string // Img400 contain 400x squared thumbnails of the screenshots.
	Backup string // Backup archives for SQL data and previously removed files.
	Emu    string // Emu has the DOSee emulation files.
	Base   string // Base or root directory path, the parent of these other subdirectories.
	UUID   string // UUID file downloads.
}

// Init initialises the subdirectories and UUID structure.
func Init(cfg conf.Config, create bool) (Dir, error) {
	d := Dir{
		Img000: cfg.Images,
		Img400: cfg.Thumbs,
		Backup: cfg.Backups,
		Emu:    cfg.Emulator,
		Base:   cfg.WebRoot,
		UUID:   cfg.Downloads,
	}
	if cfg.Images == "" {
		return Dir{}, fmt.Errorf("init %w cfg.images", ErrDir)
	}
	if cfg.Thumbs == "" {
		return Dir{}, fmt.Errorf("init %w cfg.thumbs", ErrDir)
	}
	if cfg.Backups == "" {
		return Dir{}, fmt.Errorf("init %w cfg.backups", ErrDir)
	}
	if cfg.Emulator == "" {
		return Dir{}, fmt.Errorf("init %w cfg.emulator", ErrDir)
	}
	if cfg.WebRoot == "" {
		return Dir{}, fmt.Errorf("init %w cfg.webroot", ErrDir)
	}
	if cfg.Downloads == "" {
		return Dir{}, fmt.Errorf("init %w cfg.downloads", ErrDir)
	}
	if !create {
		return d, nil
	}
	if err := createDirectories(&d); err != nil {
		return Dir{}, err
	}
	if err := PlaceHolders(&d); err != nil {
		return Dir{}, err
	}
	return d, nil
}

// createDirectories generates a series of UUID subdirectories.
func createDirectories(dir *Dir) error {
	if dir == nil {
		return fmt.Errorf("create directories: %w", ErrNil)
	}
	mkdir := func(path string) error {
		if path == "" {
			return nil
		}
		if err := create.MkDir(path); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
		return nil
	}
	if err := mkdir(dir.Img000); err != nil {
		return err
	}
	if err := mkdir(dir.Img400); err != nil {
		return err
	}
	if err := mkdir(dir.Backup); err != nil {
		return err
	}
	if err := mkdir(dir.Emu); err != nil {
		return err
	}
	if err := mkdir(dir.Base); err != nil {
		return err
	}
	return mkdir(dir.UUID)
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
func Files(cfg conf.Config, name string) (Dir, error) {
	dirs, err := Init(cfg, false)
	if err != nil {
		return Dir{}, err
	}
	dirs.UUID = filepath.Join(dirs.UUID, name)
	dirs.Emu = filepath.Join(dirs.Emu, name)
	dirs.Img000 = filepath.Join(dirs.Img000, name)
	dirs.Img400 = filepath.Join(dirs.Img400, name)
	return dirs, nil
}

// PlaceHolders generates a collection placeholder files in the UUID subdirectories.
func PlaceHolders(dir *Dir) error {
	if dir == nil {
		return fmt.Errorf("place holders: %w", ErrNil)
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
