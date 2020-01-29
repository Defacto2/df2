package archive

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
	"github.com/gabriel-vasile/mimetype"
	unarr "github.com/gen2brain/go-unarr"
)

type content struct {
	name       string
	file       int
	ext        string
	path       string
	mime       *mimetype.MIME
	modtime    time.Time
	size       int64
	executable bool
	textfile   bool
}

type contents map[int]content

func (c content) String() string {
	return fmt.Sprintf("%v (%v extension)", &c.name, c.ext)
}

// filemime saves the file MIME information and name extension.
func (c *content) filemime(f os.FileInfo) error {
	m, err := mimetype.DetectFile(c.path)
	if err != nil {
		return err
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
func FileCopy(name, dest string) (int64, error) {
	src, err := os.Open(name)
	if err != nil {
		return 0, err
	}
	defer src.Close()
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	i, err := io.Copy(file, src)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// FileMove copies a file to the destination and then deletes the source.
func FileMove(name, dest string) (int64, error) {
	i, err := FileCopy(name, dest)
	err = os.Remove(name)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// NewExt swaps or appends the extension to a filename.
func NewExt(name, extension string) string {
	e := filepath.Ext(name)
	if e == "" {
		return name + extension
	}
	fn := strings.TrimSuffix(name, e)
	return fn + extension
}

// Read returns a list of files within an rar, tar, zip or 7z archive.
// In the future I would like to add support for the following archives
// "tar.gz", "gz", "lzh", "lha", "cab", "arj", "arc".
func Read(name string) ([]string, error) {
	a, err := unarr.NewArchive(name)
	if err != nil {
		return nil, err
	}
	defer a.Close()
	list, err := a.List()
	if err != nil {
		return nil, err
	}
	return list, nil
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
