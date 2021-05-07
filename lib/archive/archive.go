// Package archive handles collections of files that are either packaged together or compressed.
package archive

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize" //nolint:misspell
	"github.com/gabriel-vasile/mimetype"
)

const (
	bat  = ".bat"
	bmp  = ".bmp"
	com  = ".com"
	diz  = ".diz"
	exe  = ".exe"
	gif  = ".gif"
	jpg  = ".jpg"
	nfo  = ".nfo"
	png  = ".png"
	tiff = ".tiff"
	txt  = ".txt"
	webp = ".webp"
)

type content struct {
	name       string
	ext        string
	path       string
	mime       *mimetype.MIME
	size       int64
	executable bool
	textfile   bool
}

type contents map[int]content

func (c content) String() string {
	return fmt.Sprintf("%v (%v extension)", &c.name, c.ext)
}

var (
	// ErrNoCustom no customization.
	ErrNoCustom = errors.New("does not support customization")
	// ErrNotArc not an archive.
	ErrNotArc = errors.New("format specified by source filename is not an archive format")
	// ErrSameArgs same same.
	ErrSameArgs = errors.New("name and dest cannot be the same")
	// ErrWalkrFmt cannot walk archive.
	ErrWalkrFmt = errors.New("format specified by archive filename is not a walker format")
)

// filemime saves the file MIME information and name extension.
func (c *content) filemime() error {
	m, err := mimetype.DetectFile(c.path)
	if err != nil {
		return fmt.Errorf("filemime failure on %q: %w", c.path, err)
	}
	c.mime = m
	// flag useful files
	switch c.ext {
	case bat, com, exe:
		c.executable = true
	case diz, nfo, txt:
		c.textfile = true
	}
	return nil
}

// FileCopy copies a file to the destination.
func FileCopy(name, dest string) (written int64, err error) {
	src, err := os.Open(name)
	if err != nil {
		return 0, fmt.Errorf("filecopy open %q: %w", name, err)
	}
	defer src.Close()
	dst, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return 0, fmt.Errorf("filecopy dest %q: %w", dest, err)
	}
	defer dst.Close()
	if written, err = io.Copy(dst, src); err != nil {
		return 0, fmt.Errorf("filecopy io.copy: %w", err)
	}
	return written, dst.Close()
}

// FileMove copies a file to the destination and then deletes the source.
func FileMove(name, dest string) (written int64, err error) {
	if name == dest {
		return 0, fmt.Errorf("filemove: %w", ErrSameArgs)
	}
	if written, err = FileCopy(name, dest); err != nil {
		return 0, fmt.Errorf("filemove: %w", err)
	}
	if _, err = os.Stat(dest); os.IsNotExist(err) {
		return 0, fmt.Errorf("filemove dest: %w", err)
	}
	if err = os.Remove(name); err != nil {
		return 0, fmt.Errorf("filemove remove name %q: %w", name, err)
	}
	return written, nil
}

// Read returns a list of files within an rar, tar, zip or 7z archive.
// uuid is the absolute path to the archive file named as a unique id.
// filename is the original archive filename and file extension.
func Read(uuid, filename string) (files []string, err error) {
	if files, err = Readr(uuid, filename); err != nil {
		return nil, fmt.Errorf("read readr fallback: %w", err)
	}
	return files, nil
}

// Restore unpacks or decompresses a given archive file to the destination.
// The archive format is selected implicitly. Restore relies on the filename
// extension to determine which decompression format to use, which must be
// supplied using filename.
// uuid is the absolute path to the archive file named as a unique id.
// filename is the original archive filename and file extension.
func Restore(uuid, filename, destination string) (files []string, err error) {
	if err = Unarchiver(uuid, filename, destination); err != nil {
		return nil, fmt.Errorf("restore unarchiver: %w", err)
	}
	if files, err = Readr(uuid, filename); err != nil {
		return nil, fmt.Errorf("restore readr: %w", err)
	}
	return files, nil
}

// dir lists the content of a directory.
func dir(name string) error {
	files, err := ioutil.ReadDir(name)
	if err != nil {
		return fmt.Errorf("dir read name %q: %w", name, err)
	}
	for _, f := range files {
		mime, err := mimetype.DetectFile(name + "/" + f.Name())
		if err != nil {
			return fmt.Errorf("dir mime failure on %q: %w", f, err)
		}
		logs.Println(f.Name(), humanize.Bytes(uint64(f.Size())), mime)
	}
	return nil
}
