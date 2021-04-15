package archive

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
)

type finds map[string]int

// Demozoo data extracted from an archive.
type Demozoo struct {
	DOSee string // dosee_run_program column
	NFO   string // retrotxt_readme column
}

func (d Demozoo) String() string {
	return fmt.Sprintf("using %q for DOSee and %q as the NFO or text", d.DOSee, d.NFO)
}

// ExtractDemozoo decompresses and parses archives fetched from Demozoo.org.
func ExtractDemozoo(name, uuid string, varNames *[]string) (dz Demozoo, err error) {
	if err = database.CheckUUID(uuid); err != nil {
		return Demozoo{}, fmt.Errorf("extract demozoo checkuuid %q: %w", uuid, err)
	}
	// create temp dir
	tempDir, err := ioutil.TempDir("", "extarc-")
	if err != nil {
		return Demozoo{}, fmt.Errorf("extract demozoo tempdir %q: %w", tempDir, err)
	}
	defer os.RemoveAll(tempDir)
	filename, err := database.LookupFile(uuid)
	if err != nil {
		return Demozoo{}, fmt.Errorf("extract demozoo lookup id %q: %w", uuid, err)
	}
	if _, err = Restore(name, filename, tempDir); err != nil {
		return Demozoo{}, fmt.Errorf("extract demozoo restore %q: %w", filename, err)
	}
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return Demozoo{}, fmt.Errorf("extract demozoo readdir %q: %w", tempDir, err)
	}
	zips := make(contents)
	for i, f := range files {
		var zip content
		zip.path = tempDir // filename gets appended by z.scan()
		zip.filescan(f)
		if err = zip.filemime(); err != nil {
			return Demozoo{}, fmt.Errorf("extract demozoo filemime %q: %w", f, err)
		}
		zips[i] = zip
	}
	if nfo := findNFO(name, zips, varNames); nfo != "" {
		if ok, err := moveText(filepath.Join(tempDir, nfo), uuid); err != nil {
			return Demozoo{}, fmt.Errorf("extract demozo move nfo: %w", err)
		} else if !ok {
			dz.NFO = nfo
		}
	}
	if dos := findDOS(name, zips, varNames); dos != "" {
		dz.DOSee = dos
	}
	return dz, nil
}

// filescan saves the filename, size and temporary path.
func (c *content) filescan(f os.FileInfo) {
	c.name = f.Name()
	c.ext = strings.ToLower(filepath.Ext(f.Name()))
	c.size = f.Size()
	c.path = path.Join(c.path, f.Name())
}

func moveText(name, uuid string) (ok bool, err error) {
	if name == "" {
		return false, nil
	}
	if err = database.CheckUUID(uuid); err != nil {
		return false, fmt.Errorf("movetext check uuid %q: %w", uuid, err)
	}
	f := directories.Files(uuid)
	size, err := FileMove(name, f.UUID+txt)
	if err != nil {
		return false, fmt.Errorf("movetext filemove %q: %w", name, err)
	}
	logs.Printf(" • NFO » %s", humanize.Bytes(uint64(size)))
	return true, nil
}

// top returns the highest prioritised filename from a collection of finds.
func (f finds) top() string {
	if len(f) == 0 {
		return ""
	}
	type fp struct {
		Filename string
		Priority int
	}
	ss := make([]fp, len(f))
	i := 0
	for k, v := range f {
		ss[i] = fp{k, v}
		i++
	}
	sort.SliceStable(ss, func(i, j int) bool {
		return ss[i].Priority < ss[j].Priority // '<' equals assending order
	})
	for _, kv := range ss {
		return kv.Filename // return first result
	}
	return ""
}

func findDOS(name string, files contents, varNames *[]string) string {
	f := make(finds) // filename and priority values
	for _, file := range files {
		if !file.executable {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name)) // base filename without extension
		fn := strings.ToLower(file.name)                     // normalise filenames
		ext := strings.ToLower(file.ext)                     // normalise file extensions
		e := findVariant(fn, exe, varNames)
		c := findVariant(fn, com, varNames)
		fmt.Printf(" > %q, %q, chk1 %s", ext, fn, base+exe)
		switch {
		case ext == bat: // [random].bat
			f[file.name] = 1
		case fn == base+exe: // [archive name].exe
			f[file.name] = 2
		case fn == base+com: // [archive name].com
			f[file.name] = 3
		case e != "":
			f[file.name] = 4
		case c != "":
			f[file.name] = 5
		case ext == exe: // [random].exe
			f[file.name] = 6
		case ext == com: // [random].com
			f[file.name] = 7
		}
	}
	return f.top()
}

func findNFO(name string, files contents, varNames *[]string) string {
	f := make(finds) // filename and priority values
	for _, file := range files {
		if !file.textfile {
			continue
		}
		base := strings.TrimSuffix(name, file.ext) // base filename without extension
		fn := strings.ToLower(file.name)           // normalise filenames
		ext := strings.ToLower(file.ext)           // normalise file extensions
		n := findVariant(fn, ".nfo", varNames)
		t := findVariant(fn, ".txt", varNames)
		switch {
		case fn == base+".nfo": // [archive name].nfo
			f[file.name] = 1
		case n != "":
			f[file.name] = 2
		case fn == base+".txt": // [archive name].txt
			f[file.name] = 3
		case t != "":
			f[file.name] = 4
		case ext == ".nfo": // [random].nfo
			f[file.name] = 5
		case fn == "file_id.diz": // BBS file description
			f[file.name] = 6
		case fn == base+".diz": // [archive name].diz
			f[file.name] = 7
		case fn == ".txt": // [random].txt
			f[file.name] = 8
		case fn == ".diz": // [random].diz
			f[file.name] = 9
		default: // currently lacking is [group name].nfo and [group name].txt priorities
		}
	}
	return f.top()
}

func findVariant(name, ext string, varNames *[]string) string {
	for _, v := range *varNames {
		f := v + ext
		if f == name {
			return f
		}
	}
	return ""
}
