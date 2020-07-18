package archive

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
	"github.com/gabriel-vasile/mimetype"
	unarr "github.com/gen2brain/go-unarr"
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

// filemime saves the file MIME information and name extension.
func (c *content) filemime() error {
	m, err := mimetype.DetectFile(c.path)
	if err != nil {
		return genErr("filemime", err)
	}
	c.mime = m
	// flag useful files
	switch c.ext {
	case ".exe", ".bat", ".com":
		c.executable = true
	case ".nfo", ".diz", ".txt":
		c.textfile = true
	}
	return nil
}

// FileCopy copies a file to the destination.
func FileCopy(name, dest string) (written int64, err error) {
	src, err := os.Open(name)
	if err != nil {
		return 0, genErr("filecopy", err)
	}
	defer src.Close()
	dst, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return 0, genErr("filecopy", err)
	}
	defer dst.Close()
	written, err = io.Copy(dst, src)
	if err != nil {
		return 0, genErr("filecopy", err)
	}
	return written, dst.Close()
}

// FileMove copies a file to the destination and then deletes the source.
func FileMove(name, dest string) (written int64, err error) {
	if name == dest {
		return written, err
	}
	if written, err = FileCopy(name, dest); err != nil {
		return 0, genErr("filemove", err)
	}
	if _, err = os.Stat(dest); os.IsNotExist(err) {
		return 0, err
	}
	if err = os.Remove(name); err != nil {
		return 0, genErr("filemove", err)
	}
	return written, err
}

// NewExt swaps or appends the extension to a filename.
func NewExt(name, extension string) (filename string) {
	e := filepath.Ext(name)
	if e == "" {
		return name + extension
	}
	fn := strings.TrimSuffix(name, e)
	return fn + extension
}

// Read returns a list of files within an rar, tar, zip or 7z archive.
// archive is the absolute path to the archive file named as a uuid
// filename is the original archive filename and file extension
func Read(archive string, filename string) (files []string, err error) {
	a, err := unarr.NewArchive(archive)
	if err != nil {
		// using archiver as a fallback
		files, err = Readr(archive, filename)
		if err != nil {
			return nil, genErr("readr", err)
		}
		return files, nil
	}
	defer a.Close()
	files, err = a.List()
	if err != nil {
		return nil, genErr("read", err)
	}
	return files, nil
}

// Restore unpacks or decompresses a given archive file to the destination.
// The archive format is selected implicitly.
// Restore relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func Restore(source, filename, destination string) (files []string, err error) {
	a, err := unarr.NewArchive(source)
	if err != nil {
		if err = Unarchiver(source, filename, destination); err != nil {
			return nil, genErr("unarchiver", err)
		}
		if files, err = Readr(source, filename); err != nil {
			return nil, genErr("readr", err)
		}
	} else {
		defer a.Close()
		files, err = a.Extract(destination)
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

// dir lists the content of a directory.
func dir(name string) {
	files, err := ioutil.ReadDir(name)
	logs.Check(err)
	for _, file := range files {
		mime, err := mimetype.DetectFile(name + "/" + file.Name())
		logs.Log(err)
		logs.Println(file.Name(), humanize.Bytes(uint64(file.Size())), mime)
	}
}

func genErr(name string, err error) error {
	return fmt.Errorf("archive %s: %v", name, err)
}
