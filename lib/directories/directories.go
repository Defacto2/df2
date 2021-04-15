// Package directories interacts with the filepaths that hold files and assets.
package directories

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	m "math/rand"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"time"

	"github.com/spf13/viper"
)

// random characters used by randString().
const random string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321 .!?"

var (
	ErrPathIsFile = errors.New("path already exist as a file")
	ErrPrefix     = errors.New("invalid prefix value, it must be between 0 - 9")
)

// Dir is a collection of paths containing files.
type Dir struct {
	Img000 string // path to screencaptures and previews
	Img150 string // path to 150x150 squared thumbnails
	Img400 string // path to 400x400 squared thumbnails
	Backup string // path to the backup archives or previously removed files
	Emu    string // path to the DOSee emulation files
	Base   string // base directory path that hosts these other subdirectories
	UUID   string // path to file downloads with UUID as filenames
}

// Init initializes the subdirectories and UUID structure.
func Init(create bool) Dir {
	if viper.GetString("directory.root") == "" {
		viper.SetDefault("directory.000", "/opt/assets/000")
		viper.SetDefault("directory.150", "/opt/assets/150")
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
	d.Img150 = viper.GetString("directory.150")
	d.Img400 = viper.GetString("directory.400")
	d.Backup = viper.GetString("directory.backup")
	d.Emu = viper.GetString("directory.emu")
	d.Base = viper.GetString("directory.root")
	d.UUID = viper.GetString("directory.uuid")
	if create {
		if err := createDirectories(&d); err != nil {
			log.Fatal(err)
		}
		if err := createPlaceHolders(&d); err != nil {
			log.Fatal(err)
		}
	}
	return d
}

// Files initializes the full path filenames for a UUID.
func Files(name string) (dirs Dir) {
	dirs = Init(false)
	dirs.UUID = path.Join(dirs.UUID, name)
	dirs.Emu = path.Join(dirs.Emu, name)
	dirs.Img000 = path.Join(dirs.Img000, name)
	dirs.Img400 = path.Join(dirs.Img400, name)
	dirs.Img150 = path.Join(dirs.Img150, name)
	return dirs
}

// createDirectories generates a series of UUID subdirectories.
func createDirectories(dir *Dir) error {
	v := reflect.ValueOf(dir)
	// iterate through the D struct values
	for i := 0; i < v.NumField(); i++ {
		if d := fmt.Sprintf("%v", v.Field(i).Interface()); d != "" {
			if err := createDirectory(d); err != nil {
				return fmt.Errorf("create directory: %w", err)
			}
		}
	}
	return nil
}

// createDirectory creates a UUID subdirectory provided to dir.
func createDirectory(dir string) error {
	src, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory mkdir %q: %w", dir, err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("create directory stat %q: %w", dir, err)
	}
	if src.Mode().IsRegular() {
		return fmt.Errorf("create directory %q: %w", dir, ErrPathIsFile)
	}
	return nil
}

// createHolderFiles generates a number of placeholder files in the given directory.
func createHolderFiles(dir string, size int, number uint) error {
	const max = 9
	if number > max {
		return fmt.Errorf("create holder files number=%d: %w", number, ErrPrefix)
	}
	for i := uint(0); i <= number; i++ {
		if err := createHolderFile(dir, size, i); err != nil {
			return fmt.Errorf("create holder files: %w", err)
		}
	}
	return nil
}

// createHolderFile generates a placeholder file filled with random text in the given directory,
// the size of the file determines the number of random characters and the prefix is a digit between
// 0 and 9 is appended to the filename.
func createHolderFile(dir string, size int, prefix uint) error {
	const max = 9
	if prefix > max {
		return fmt.Errorf("create holder file prefix=%d: %w", prefix, ErrPrefix)
	}
	name := fmt.Sprintf("00000000-0000-0000-0000-00000000000%v", prefix)
	fn := path.Join(dir, name)
	if _, err := os.Stat(fn); err == nil {
		return nil // don't overwrite existing files
	}
	m.Seed(time.Now().UnixNano())
	r, err := randString(size)
	if err != nil {
		return fmt.Errorf("create holder file: %w", err)
	}
	text := []byte(r)
	if err := ioutil.WriteFile(fn, text, 0644); err != nil {
		return fmt.Errorf("write create holder file %q: %w", fn, err)
	}
	return nil
}

// createPlaceHolders generates a collection placeholder files in the UUID subdirectories.
func createPlaceHolders(dir *Dir) error {
	if err := createHolderFiles(dir.UUID, 1000000, 9); err != nil {
		return fmt.Errorf("create uuid holders: %w", err)
	}
	if err := createHolderFiles(dir.Emu, 1000000, 2); err != nil {
		return fmt.Errorf("create emu holders: %w", err)
	}
	if err := createHolderFiles(dir.Img000, 1000000, 9); err != nil {
		return fmt.Errorf("create img000 holders: %w", err)
	}
	if err := createHolderFiles(dir.Img400, 500000, 9); err != nil {
		return fmt.Errorf("create img400 holders: %w", err)
	}
	if err := createHolderFiles(dir.Img150, 100000, 9); err != nil {
		return fmt.Errorf("create img150 holders: %w", err)
	}
	return nil
}

// randString generates a random string of n x characters.
func randString(n int) (string, error) {
	s, r := make([]rune, n), []rune(random)
	for i := range s {
		p, err := rand.Prime(rand.Reader, len(r))
		if err != nil {
			return "", fmt.Errorf("random string n %d: %w", n, err)
		}
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s), nil
}

func Size(path string) (count int64, bytes uint64, err error) {
	err = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			bytes += uint64(info.Size())
			count += 1
		}
		return err
	})
	return count, bytes, err
}
