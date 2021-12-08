package content

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

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

type File struct {
	Name       string
	Ext        string
	Path       string
	Mime       *mimetype.MIME
	Size       int64
	Executable bool
	Textfile   bool
}

type Contents map[int]File

func (c File) String() string {
	return fmt.Sprintf("%v (%v extension)", &c.Name, c.Ext)
}

// MIME saves the file MIME information and name extension.
func (c *File) MIME() error {
	m, err := mimetype.DetectFile(c.Path)
	if err != nil {
		return fmt.Errorf("filemime failure on %q: %w", c.Path, err)
	}
	c.Mime = m
	// flag useful files
	switch c.Ext {
	case bat, com, exe:
		c.Executable = true
	case diz, nfo, txt:
		c.Textfile = true
	}
	return nil
}

// Scan saves the filename, size and temporary path.
func (c *File) Scan(f os.FileInfo) {
	c.Name = f.Name()
	c.Ext = strings.ToLower(filepath.Ext(f.Name()))
	c.Size = f.Size()
	c.Path = path.Join(c.Path, f.Name())
}
