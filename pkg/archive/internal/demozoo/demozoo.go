package demozoo

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Defacto2/df2/pkg/archive/internal/content"
	"github.com/Defacto2/df2/pkg/archive/internal/file"
	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/dustin/go-humanize"
)

var ErrNoSrc = errors.New("no src filename was given")

// Usability of search, filename pattern matches.
type Usability uint

const (
	Lvl1 Usability = iota + 1 // Lvl1 is the highest usability.
	Lvl2
	Lvl3
	Lvl4
	Lvl5
	Lvl6
	Lvl7
	Lvl8
	Lvl9 // Lvl9 is the least usable.
)

const (
	bat = ".bat"
	com = ".com"
	diz = ".diz"
	exe = ".exe"
	nfo = ".nfo"
	txt = ".txt"

	FileDiz = "file_id.diz" // Filename for a common, brief description of the content of archive.
)

// Data extracted from an archive.
type Data struct {
	DOSee string // Table dosee_run_program column.
	NFO   string // Table retrotxt_readme column.
}

func (d Data) String() string {
	return fmt.Sprintf("using %q for DOSee and %q as the NFO or text", d.DOSee, d.NFO)
}

// Finds are a collection of matched filenames and their usability ranking.
type Finds map[string]Usability

// Top returns the most usable filename from a collection of finds.
func (f Finds) Top() string {
	if len(f) == 0 {
		return ""
	}
	type fp struct {
		Filename  string
		Usability Usability
	}
	ss := make([]fp, len(f))
	i := 0
	for k, v := range f {
		ss[i] = fp{k, v}
		i++
	}
	sort.SliceStable(ss, func(i, j int) bool {
		return ss[i].Usability < ss[j].Usability // '<' equals assending order
	})
	for _, kv := range ss {
		return kv.Filename // return first result
	}
	return ""
}

// DOS attempts to discover a software package starting executable from a collection of files.
func DOS(w io.Writer, name string, files content.Contents, varNames *[]string) string {
	if w == nil {
		w = io.Discard
	}
	f := make(Finds) // filename and priority values
	for _, file := range files {
		if !file.Executable {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name)) // base filename without extension
		fn := strings.ToLower(file.Name)                     // normalise filenames
		ext := strings.ToLower(file.Ext)                     // normalise file extensions
		e := findVariant(fn, exe, varNames)
		c := findVariant(fn, com, varNames)
		fmt.Fprintf(w, " > %q, %q, chk1 %s", ext, fn, base+exe)
		switch {
		case ext == bat: // [random].bat
			f[file.Name] = Lvl1
		case fn == base+exe: // [archive name].exe
			f[file.Name] = Lvl2
		case fn == base+com: // [archive name].com
			f[file.Name] = Lvl3
		case e != "":
			f[file.Name] = Lvl4
		case c != "":
			f[file.Name] = Lvl5
		case ext == exe: // [random].exe
			f[file.Name] = Lvl6
		case ext == com: // [random].com
			f[file.Name] = Lvl7
		}
	}
	return f.Top()
}

// MoveText moves the name file to a [uuid].txt named file.
func MoveText(w io.Writer, cfg conf.Config, src, uuid string) error {
	if w == nil {
		w = io.Discard
	}
	if src == "" {
		return ErrNoSrc
	}
	if err := database.CheckUUID(uuid); err != nil {
		return fmt.Errorf("movetext check uuid %q: %w", uuid, err)
	}
	f, err := directories.Files(cfg, uuid)
	if err != nil {
		return fmt.Errorf("movetext directory: %w", err)
	}
	var size int64
	if uuid != database.TestID {
		size, err = file.Move(src, f.UUID+txt)
		if err != nil {
			return fmt.Errorf("movetext filemove %q: %w", src, err)
		}
	}
	fmt.Fprintf(w, " • NFO » %s", humanize.Bytes(uint64(size)))
	return nil
}

// NFO attempts to discover a archive package NFO or information textfile from a collection of files.
func NFO(name string, files content.Contents, varNames *[]string) string {
	f := make(Finds) // filename and priority values
	for _, file := range files {
		if !file.Textfile {
			continue
		}
		obj := parseNfo(name, file, varNames)
		f.setUsable(obj, file)
	}
	return f.Top()
}

type nfoObj struct {
	base     string // base filename without extension
	filename string // normalise filenames
	ext      string // normalise file extensions
	altNFO   string
	altText  string
}

func (f Finds) setUsable(obj nfoObj, file content.File) {
	switch {
	case obj.filename == obj.base+nfo: // [archive name].nfo
		f[file.Name] = Lvl1
	case obj.altNFO != "":
		f[file.Name] = Lvl2
	case obj.filename == obj.base+txt: // [archive name].txt
		f[file.Name] = Lvl3
	case obj.altText != "":
		f[file.Name] = Lvl4
	case obj.ext == nfo: // [random].nfo
		f[file.Name] = Lvl5
	case obj.filename == FileDiz: // BBS file description
		f[file.Name] = Lvl6
	case obj.filename == obj.base+diz: // [archive name].diz
		f[file.Name] = Lvl7
	case obj.filename == txt: // [random].txt
		f[file.Name] = Lvl8
	case obj.filename == diz: // [random].diz
		f[file.Name] = Lvl9
	default: // currently lacking is [group name].nfo and [group name].txt priorities
	}
}

func parseNfo(name string, file content.File, varNames *[]string) nfoObj {
	obj := nfoObj{
		base:     strings.TrimSuffix(name, file.Ext),
		filename: strings.ToLower(file.Name),
		ext:      strings.ToLower(file.Ext),
	}
	obj.altNFO = findVariant(obj.filename, nfo, varNames)
	obj.altText = findVariant(obj.filename, txt, varNames)
	return obj
}

func findVariant(name, ext string, varNames *[]string) string {
	if varNames == nil {
		return ""
	}
	for _, v := range *varNames {
		f := v + ext
		if f == name {
			return f
		}
	}
	return ""
}
