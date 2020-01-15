package archive

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/images"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
	"github.com/gabriel-vasile/mimetype"
	unarr "github.com/gen2brain/go-unarr"
)

type task struct {
	name string // filename
	size int64  // file size
	cont bool   // continue, don't scan anymore images
}

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
	//mode ioutil.FileMode
}

type contents map[int]content

type Demozoo struct {
	DOSee string
	NFO   string
}

func (d Demozoo) String() string {
	return fmt.Sprintf("using %q for DOSee and %q as the NFO or text", d.DOSee, d.NFO)
}

// ExtractDemozoo decompresses and parses archives fetched from Demozoo.org.
func ExtractDemozoo(name, uuid string) (Demozoo, error) {
	var dz Demozoo
	if err := database.CheckUUID(uuid); err != nil {
		return dz, err
	}
	// create temp dir
	tempDir, err := ioutil.TempDir("", "extarc-")
	if err != nil {
		return dz, err
	}
	defer os.RemoveAll(tempDir)
	// extract archive
	a, err := unarr.NewArchive(name)
	if err != nil {
		return dz, err
	}
	defer a.Close()
	err = a.Extract(tempDir)
	if err != nil {
		return dz, err
	}
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return dz, err
	}
	var zips = make(contents)
	for i, f := range files {
		var zip content
		zip.path = tempDir // filename gets appended by z.scan()
		zip.filescan(f)
		err = zip.filemime(f)
		if err != nil {
			return dz, err
		}
		zips[i] = zip
	}
	if nfo := findNFO(name, zips); nfo != "" {
		println("use as a NFO:", nfo)
		if moveText(filepath.Join(tempDir, nfo), uuid) {
			dz.NFO = nfo
		}
	}
	if dos := findDOS(name, zips); dos != "" {
		println("use as a DOS binary:", dos)
		dz.DOSee = dos
	}
	return dz, nil
}

func moveText(name, uuid string) bool {
	if name == "" {
		return false
	}
	if err := database.CheckUUID(uuid); err != nil {
		logs.Check(err)
	}
	f := directories.Files(uuid)
	size, err := FileMove(name, f.UUID+".txt")
	logs.Check(err)
	print(fmt.Sprintf("  %s » ...%s.txt %s", logs.Y(), uuid[26:36], humanize.Bytes(uint64(size))))
	return true
}

func (c content) String() string {
	return fmt.Sprintf("%v (%v extension)", &c.name, c.ext)
}

// filescan saves the filename, size and temporary path.
func (c *content) filescan(f os.FileInfo) {
	c.name = f.Name()
	c.ext = strings.ToLower(filepath.Ext(f.Name()))
	c.size = f.Size()
	c.path = path.Join(c.path, f.Name())
}

type finds map[string]int

func findDOS(name string, files contents) string {
	finds := make(finds) // filename and priority values
	for _, file := range files {
		if !file.executable {
			continue
		}
		base := strings.TrimSuffix(name, file.ext) // base filename without extension
		fn := strings.ToLower(file.name)           // normalize filenames
		ext := strings.ToLower(file.ext)           // normalize file extensions
		switch {
		case ext == ".bat": // [random].bat
			finds[file.name] = 1
		case fn == base+".exe": // [archive name].exe
			finds[file.name] = 2
		case fn == base+".com": // [archive name].com
			finds[file.name] = 3
		case ext == ".exe": // [random].exe
			finds[file.name] = 4
		case ext == ".com": // [random].com
			finds[file.name] = 5
		}
	}
	return finds.top()
}

func findNFO(name string, files contents) string {
	finds := make(finds) // filename and priority values
	for _, file := range files {
		if !file.textfile {
			continue
		}
		base := strings.TrimSuffix(name, file.ext) // base filename without extension
		fn := strings.ToLower(file.name)           // normalize filenames
		ext := strings.ToLower(file.ext)           // normalize file extensions
		switch {
		case fn == base+".nfo": // [archive name].nfo
			finds[file.name] = 1
		case ext == ".nfo": // [random].nfo
			finds[file.name] = 2
		case fn == base+".txt": // [archive name].txt
			finds[file.name] = 3
		case fn == "file_id.diz": // BBS file description
			finds[file.name] = 4
		case fn == base+".diz": // [archive name].diz
			finds[file.name] = 5
		case fn == base+".txt": // [random].txt
			finds[file.name] = 6
		case fn == base+".diz": // [random].diz
			finds[file.name] = 7
		default: // currently lacking is [group name].nfo and [group name].txt priorities
		}
	}
	return finds.top()
}

// top returns the highest prioritized filename from a collection of finds.
func (f finds) top() string {
	if len(f) == 0 {
		return ""
	}
	type fp struct {
		Filename string
		Priority int
	}
	var ss []fp
	for k, v := range f {
		ss = append(ss, fp{k, v})
	}
	sort.SliceStable(ss, func(i, j int) bool {
		return ss[i].Priority < ss[j].Priority // '<' equals assending order
	})
	for _, kv := range ss {
		return kv.Filename // return first result
	}
	return ""
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

// Extract decompresses and parses a named archive.
// uuid is used to rename the extracted assets such as image previews.
func Extract(name, uuid string) error {
	if err := database.CheckUUID(uuid); err != nil {
		return err
	}
	// create temp dir
	tempDir, err := ioutil.TempDir("", "extarc-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	// extract archive
	a, err := unarr.NewArchive(name)
	if err != nil {
		return err
	}
	defer a.Close()
	err = a.Extract(tempDir)
	if err != nil {
		return err
	}
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return err
	}
	th := taskInit()
	tx := taskInit()
	for _, file := range files {
		if th.cont && tx.cont {
			break
		}
		fn := path.Join(tempDir, file.Name())
		fmime, err := mimetype.DetectFile(fn)
		if err != nil {
			return err
		}
		switch fmime.Extension() {
		case ".bmp", ".gif", ".jpg", ".png", ".tiff", ".webp":
			if th.cont {
				continue
			}
			switch {
			case file.Size() > th.size:
				th.name = fn
				th.size = file.Size()
			}
		case ".txt":
			if tx.cont {
				continue
			}
			tx.name = fn
			tx.size = file.Size()
			tx.cont = true
		}
	}
	if n := th.name; n != "" {
		images.Generate(n, uuid)
	}
	if n := tx.name; n != "" {
		f := directories.Files(uuid)
		size, err := FileMove(n, f.UUID+".txt")
		logs.Check(err)
		print(fmt.Sprintf("  %s » ...%s.txt %s", logs.Y(), uuid[26:36], humanize.Bytes(uint64(size))))
	}
	if x := true; !x {
		dir(tempDir)
	}
	return nil
}

func taskInit() task {
	return task{name: "", size: 0, cont: false}
}

// FileMove copies a file to the destination and then deletes the source.
func FileMove(name, dest string) (int64, error) {
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
