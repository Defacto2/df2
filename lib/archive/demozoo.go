package archive

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
)

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
	if err := database.CheckUUID(uuid); err != nil {
		return Demozoo{}, fmt.Errorf("extract demozoo checkuuid %q: %w", uuid, err)
	}
	// create temp dir
	tempDir, err := ioutil.TempDir("", "extarc-")
	if err != nil {
		return Demozoo{}, fmt.Errorf("extract demozoo tempdir %q: %w", tempDir, err)
	}
	defer os.RemoveAll(tempDir)
	filename, err := database.LookupFile(uuid)
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
	if err := database.CheckUUID(uuid); err != nil {
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

// top returns the highest prioritized filename from a collection of finds.
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

type finds map[string]int

func findDOS(name string, files contents, varNames *[]string) string {
	finds := make(finds) // filename and priority values
	for _, file := range files {
		if !file.executable {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name)) // base filename without extension
		fn := strings.ToLower(file.name)                     // normalize filenames
		ext := strings.ToLower(file.ext)                     // normalize file extensions
		e := findVariant(fn, exe, varNames)
		c := findVariant(fn, com, varNames)
		fmt.Printf(" > %q, %q, chk1 %s", ext, fn, base+exe)
		switch {
		case ext == bat: // [random].bat
			finds[file.name] = 1
		case fn == base+exe: // [archive name].exe
			finds[file.name] = 2
		case fn == base+com: // [archive name].com
			finds[file.name] = 3
		case e != "":
			finds[file.name] = 4
		case c != "":
			finds[file.name] = 5
		case ext == exe: // [random].exe
			finds[file.name] = 6
		case ext == com: // [random].com
			finds[file.name] = 7
		}
	}
	return finds.top()
}

func findNFO(name string, files contents, varNames *[]string) string {
	finds := make(finds) // filename and priority values
	for _, file := range files {
		if !file.textfile {
			continue
		}
		base := strings.TrimSuffix(name, file.ext) // base filename without extension
		fn := strings.ToLower(file.name)           // normalize filenames
		ext := strings.ToLower(file.ext)           // normalize file extensions
		n := findVariant(fn, ".nfo", varNames)
		t := findVariant(fn, ".txt", varNames)
		switch {
		case fn == base+".nfo": // [archive name].nfo
			finds[file.name] = 1
		case n != "":
			finds[file.name] = 2
		case fn == base+".txt": // [archive name].txt
			finds[file.name] = 3
		case t != "":
			finds[file.name] = 4
		case ext == ".nfo": // [random].nfo
			finds[file.name] = 5
		case fn == "file_id.diz": // BBS file description
			finds[file.name] = 6
		case fn == base+".diz": // [archive name].diz
			finds[file.name] = 7
		case fn == ".txt": // [random].txt
			finds[file.name] = 8
		case fn == ".diz": // [random].diz
			finds[file.name] = 9
		default: // currently lacking is [group name].nfo and [group name].txt priorities
		}
	}
	return finds.top()
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
