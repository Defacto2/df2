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

	"github.com/Defacto2/df2/lib/directories/internal/create"
	"github.com/spf13/viper"
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
func Init(create bool) Dir {
	if viper.GetString("directory.root") == "" {
		viper.SetDefault("directory.000", "/opt/assets/000")
		viper.SetDefault("directory.400", "/opt/assets/400")
		viper.SetDefault("directory.backup", "/opt/assets/backups")
		viper.SetDefault("directory.emu", "/opt/assets/emularity.zip")
		viper.SetDefault("directory.html", "/opt/assets/html")
		viper.SetDefault("directory.incoming.files", "/opt/incoming/files")
		viper.SetDefault("directory.incoming.previews", "/opt/incoming/previews")
		viper.SetDefault("directory.root", "/opt/assets")
		viper.SetDefault("directory.sql", "/opt/assets/sql")
		viper.SetDefault("directory.uuid", "/opt/assets/downloads")
		viper.SetDefault("directory.views", "/opt/assets/views")
	}
	var d Dir
	d.Img000 = viper.GetString("directory.000")
	d.Img400 = viper.GetString("directory.400")
	d.Backup = viper.GetString("directory.backup")
	d.Emu = viper.GetString("directory.emu")
	d.Base = viper.GetString("directory.root")
	d.UUID = viper.GetString("directory.uuid")
	if create {
		if err := createDirectories(&d); err != nil {
			log.Fatal(err)
		}
		if err := PlaceHolders(&d); err != nil {
			log.Fatal(err)
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
func Files(name string) Dir {
	dirs := Init(false)
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
func Size(root string) (count int64, bytes uint64, err error) {
	err = filepath.Walk(root, func(_ string, info os.FileInfo, err error) error {
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
