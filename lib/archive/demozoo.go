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
	unarr "github.com/gen2brain/go-unarr"
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
func ExtractDemozoo(name, uuid string, varNames []string) (Demozoo, error) {
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
	// varNames --> findNFO, findDOS = for{}
	if nfo := findNFO(name, zips); nfo != "" {
		if moveText(filepath.Join(tempDir, nfo), uuid) {
			dz.NFO = nfo
		}
	}
	if dos := findDOS(name, zips); dos != "" {
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
	print(fmt.Sprintf(" • NFO » %s", humanize.Bytes(uint64(size))))
	return true
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
