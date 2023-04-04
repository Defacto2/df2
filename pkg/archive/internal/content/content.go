// Package content stats the content of an archive.
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
	bat = ".bat"
	com = ".com"
	diz = ".diz"
	exe = ".exe"
	nfo = ".nfo"
	txt = ".txt"
)

// File details and metadata.
type File struct {
	Name       string         // Name of the file.
	Ext        string         // Extension of the file.
	Path       string         // Path to the file.
	Mime       *mimetype.MIME // File MIME type.
	Size       int64          // Size of file in bytes.
	Executable bool           // Executable program file.
	Textfile   bool           // Human readable text file.
}

type Contents map[int]File

func (c File) String() string {
	return fmt.Sprintf("%v (%v extension)", c.Name, c.Ext)
}

// MIME sets the file MIME information and name extension.
func (c *File) MIME() error {
	m, err := mimetype.DetectFile(c.Path)
	if err != nil {
		return fmt.Errorf("content filemime failure on %q: %w", c.Path, err)
	}
	c.Mime = m
	return nil
}

// Scan sets the filename, size and temporary path.
func (c *File) Scan(f os.FileInfo) {
	c.Name = f.Name()
	c.Ext = strings.ToLower(filepath.Ext(f.Name()))
	c.Size = f.Size()
	c.Path = path.Join(c.Path, f.Name())
	// flag useful files
	switch strings.ToLower(c.Ext) {
	case bat, com, exe:
		c.Executable = true
	case diz, nfo, txt:
		c.Textfile = true
	}
}
