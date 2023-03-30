package dizzer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/dizzer/record"
	"github.com/mholt/archiver"
)

var (
	ErrDir = errors.New("named file points to a directory")
)

var Rar = archiver.Rar{
	OverwriteExisting:      false,
	MkdirAll:               true,
	ImplicitTopLevelFolder: false,
	ContinueOnError:        true,
}

// Run the dizzer on the named .rar file.
func Run(nameRar string) error {
	if st, err := os.Stat(nameRar); err != nil {
		return fmt.Errorf("%w: %s", err, nameRar)
	} else if st.IsDir() {
		return fmt.Errorf("%w: %s", err, nameRar)
	}
	// todo: confirm rar archive.

	// first stat the .rar file
	st := Stat{}
	if err := st.Walk(nameRar); err != nil {
		return err
	}

	// if okay, then uncompress it to a tmpdir
	dir, err := os.MkdirTemp(os.TempDir(), "dizzer")
	if err != nil {
		return err
	}
	//defer os.RemoveAll(dir)
	fmt.Println("pre unarchive")
	if err := Rar.Unarchive(nameRar, dir); err != nil {
		return err
	}

	fmt.Println("unarchive okay")
	// built a collection of records to insert to the db.
	itemsToAdd := st.NFOs + st.DIZs
	rec := make([]record.Record, itemsToAdd)
	// i := -1
	// file := ""
	for rel, r := range st.Releases {
		fmt.Printf("%s from %s with %d items.\n", rel, r.LastMod.Format("2006-01-02"), len(r.Files))
		// if r.DIZ {
		// 	i++
		// 	file = filepath.Join(dir, r.Path, record.FileID)
		// 	rec[i] = record.New(r.Path)
		// 	rec[i].LastMod = r.LastMod // always use the oldest last modifier
		// 	if err := rec[i].Stat(file); err != nil {
		// 		return err
		// 	}
		// 	fmt.Printf(" --- > %s\n", r.Title)
		// 	if err := rec[i].Read(file); err != nil {
		// 		return err
		// 	}
		// }
		// if r.NFO != "" {
		// 	i++
		// 	file = filepath.Join(dir, r.Path, r.NFO)
		// 	rec[i] = record.New(r.Path)
		// 	rec[i].LastMod = r.LastMod // always use the oldest last modifier
		// 	if err := rec[i].Stat(file); err != nil {
		// 		return err
		// 	}
		// 	if err := rec[i].Read(file); err != nil {
		// 		return err
		// 	}
		// }
		// todo: replace r.NFO bool with string containing filename
		// match .nfo and file containing st.Group

	}

	for i, r := range rec {
		fmt.Printf("\n%d. %+v\n", i, r)
	}

	fmt.Println("nfos:", st.NFOs)
	fmt.Println("dizes:", st.DIZs)
	fmt.Println("other files:", st.Others)
	fmt.Println("group:", st.Group)
	fmt.Printf("years: %+v\n", st.LastMods)

	return nil
}

// Stat the collection of NFO and file_id.diz files within an RAR archive.
type Stat struct {
	Name     string             // Name of the RAR archive file.
	Group    string             // Group for the collection.
	DIZs     int                // DIZs file_id.diz count.
	NFOs     int                // NFOs count.
	Others   int                // Other types of files count.
	LastMods Years              // Publish counts the last modified years.
	Releases map[string]Release // Releases lists every release included in the RAR archive.
}

// Release is a subdirectory of the RAR archive.
type Release struct {
	Title string // Title for the release.
	Path  string // Path is the subdirectory containing the release.
	//DIZ     bool      // DIZ means a file_id.diz is included with the release.
	//NFO     string    // NFO is the filename of the .nfo information file included with the release.
	Files   []string  // Files named that are included in the release.
	LastMod time.Time // The earliest file last modification time found in the release.
	Diz     record.Download
	Nfo     record.Download
}

// Walk the named .rar archive file to collect information and statistics.
func (st *Stat) Walk(name string) error {
	st.Releases = map[string]Release{}
	key := ""
	err := Rar.Walk(name, func(f archiver.File) error {
		if st.Name == "" {
			st.Name = name
		}
		if f.IsDir() {
			return nil
		}
		if st.Group == "" {
			st.Group = PathGroup(f.Name())
		}
		if filepath.Dir(f.Name()) != key {
			key = filepath.Dir(f.Name())
		}

		store, exists := st.Releases[key]
		if !exists {
			store.Title = PathTitle(key)
			store.LastMod = f.ModTime()
		} else {
			i := store.LastMod.Compare(f.ModTime())
			olderModTime := i > 0
			if olderModTime {
				store.LastMod = f.ModTime()
			}
		}

		store.Path = filepath.Dir(f.Name())
		store.Files = append(store.Files, filepath.Base(f.Name()))

		ext := strings.ToLower(filepath.Ext(f.Name()))
		switch ext {
		case ".diz":
			st.DIZs++
			//store.DIZ = true
			store.Diz = record.Download{}
			path := filepath.Join(dir, r.Path, r.NFO)
			if err := store.Diz.New(); err != nil {
				return err
			}
		case ".nfo":
			st.NFOs++
			//store.NFO = filepath.Base(f.Name()) // todo: func to confirm nfo has group name
		default:
			st.Others++
		}

		st.LastMods = st.LastMods.Add(f.ModTime())

		st.Releases[key] = store

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// Years counts the file last modified dates as years.
type Years map[string]int

// Add a last modified date to the Years map.
func (y Years) Add(mod time.Time) Years {
	if y == nil {
		y = Years{}
	}
	y[mod.Format("2006")]++
	return y
}

// Database fetches metadata for the database using the name directory,
// and the file
func Database(name string, f os.FileInfo) {
	// return a new struct?
	// Have sep func to parse directory name.
}

// PathGroup returns the group name or initalism extracted from the
// named directory path of the release. This is intended as a fallback
// when the file_id.diz cannot be parsed.
func PathGroup(name string) string {
	n := strings.Split(name, string(filepath.Separator))
	s := strings.Split(n[0], "-")
	x := len(s) - 1
	if x > 0 {
		return s[x]
	}
	return ""
}

// PathTitle returns the title of the release extracted from the
// named directory path of the release. This is intended as a fallback
// when the file_id.diz cannot be parsed.
func PathTitle(name string) string {
	n := strings.LastIndex(name, "-")
	t := name
	if n > -1 {
		t = name[0:n]
	}
	// match v1.0.0
	r := regexp.MustCompile(`v(\d+)\.(\d+)\.(\d+)`)
	t = r.ReplaceAllString(t, "v$1-$2-$3")
	// match v1.0
	r = regexp.MustCompile(`v(\d+)\.(\d+)`)
	t = r.ReplaceAllString(t, "v$1-$2")

	words := strings.Split(t, ".")
	for i, word := range words {
		switch strings.ToLower(word) {
		case "incl":
			words[i] = "including"
		case "keymaker":
			words[i] = "keymaker"
		}
	}

	t = strings.Join(words, " ")
	// restore v1.0.0
	r = regexp.MustCompile(`v(\d+)-(\d+)-(\d+)`)
	t = r.ReplaceAllString(t, "v$1.$2.$3")
	// restore v1.0
	r = regexp.MustCompile(`v(\d+)-(\d+)`)
	t = r.ReplaceAllString(t, "v$1.$2")

	return strings.TrimSpace(t)
}
