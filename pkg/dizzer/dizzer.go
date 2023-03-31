package dizzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/dizzer/record"
	"github.com/Defacto2/df2/pkg/dizzer/zwt"
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
	if err := Rar.CheckExt(nameRar); err != nil {
		return fmt.Errorf("%w: %s", err, nameRar)
	}
	if st, err := os.Stat(nameRar); err != nil {
		return fmt.Errorf("%w: %s", err, nameRar)
	} else if st.IsDir() {
		return fmt.Errorf("%w: %s", err, nameRar)
	}

	tick := time.Now()

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
	defer os.RemoveAll(dir)

	if err := Rar.Unarchive(nameRar, dir); err != nil {
		return err
	}

	// build a collection of records to insert to the db.
	ch1 := make(chan record.Download)
	ch2 := make(chan record.Download)

	newRecs := 0
	for key, r := range st.Releases {

		go func() {
			diz := filepath.Join(dir, r.Path, record.FileID)
			if err := r.Diz.New(diz, st.Group); err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					fmt.Fprintln(os.Stderr, err)
				}
			}
			ch1 <- r.Diz
		}()

		go func() {
			if r.Nfo.Name == "" {
				ch2 <- r.Nfo
			}
			nfo := filepath.Join(dir, r.Path, r.Nfo.Name)
			if err := r.Nfo.New(nfo, st.Group); err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					fmt.Fprintln(os.Stderr, err)
				}
			}
			ch2 <- r.Nfo
		}()

		diz := <-ch1
		nfo := <-ch2

		// match lastmod date to the earliest
		if diz.LastMod.Before(nfo.LastMod) {
			nfo.LastMod = diz.LastMod
		} else if nfo.LastMod.Before(diz.LastMod) {
			diz.LastMod = nfo.LastMod
		}
		// copy readdate to nfo
		if nfo.ReadDate.IsZero() && !diz.ReadDate.IsZero() {
			nfo.ReadDate = diz.ReadDate
		}

		// make a copy of the release
		if entry, ok := st.Releases[key]; ok {
			// modify the data of the copy
			entry.Diz = diz
			entry.Nfo = nfo
			// reassign the copy to the original
			st.Releases[key] = entry
		}
		if diz != (record.Download{}) {
			newRecs++
		}
		if nfo != (record.Download{}) {
			newRecs++
		}
	}

	records := make([]record.Record, newRecs)
	i, l := 0, len(records)
	// TODO: build record collection
	for _, r := range st.Releases {
		if i >= l {
			//break
			fmt.Println("more than expacted?", i, l)
			break
		}
		//fmt.Printf("\n%d. %+v\n", i, r)
		title := ""
		if r.Diz == (record.Download{}) {
			fmt.Println("no file_id.diz for", r.Title)
		} else {
			records[i], err = record.New(r.Path, st.Group)
			if err != nil {
				fmt.Println(err)
			}
			title = r.Diz.ReadTitle
			if title == "" {
				title = r.Title
			}
			if err := records[i].Copy(&r.Diz, title); err != nil {
				fmt.Println(err)
			}
			i++
		}
		if r.Nfo == (record.Download{}) {
			fmt.Println("no readme for ", r.Title, "files inc.", r.Files)
		} else {
			records[i], err = record.New(r.Path, st.Group)
			if err != nil {
				fmt.Println(err)
			}
			if title == "" {
				title = r.Title
			}
			records[i].Title = title
			if err := records[i].Copy(&r.Nfo, title); err != nil {
				fmt.Println(err)
			}
			i++
		}
	}

	for i, r := range records {
		u, err := json.MarshalIndent(r, "", " ")
		if err != nil {
			continue
		}
		fmt.Printf("%d.\t%+v of %s\n", i, r.Title, r.FileName)
		fmt.Println(string(u))
	}

	fmt.Println("nfos:", st.NFOs)
	fmt.Println("dizes:", st.DIZs)
	fmt.Println("other files:", st.Others)
	fmt.Println("group:", st.Group)
	fmt.Printf("years: %+v\n", st.LastMods)
	fmt.Println("how many new records?", len(records), "vs i", i)

	//	defer close(ch2)
	time.Sleep(1 * time.Second)

	fmt.Println("time taken", time.Since(tick).Seconds())

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
	Files []string // Files named that are included in the release.
	//LastMod time.Time // The earliest file last modification time found in the release.
	Diz record.Download
	Nfo record.Download
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
			st.Group = Group(f.Name())
		}
		if filepath.Dir(f.Name()) != key {
			key = filepath.Dir(f.Name())
		}

		store, exists := st.Releases[key]
		if !exists {
			store.Title = PathTitle(key)
		}
		store.Path = filepath.Dir(f.Name())
		store.Files = append(store.Files, filepath.Base(f.Name()))

		ext := strings.ToLower(filepath.Ext(f.Name()))
		switch ext {
		case ".diz":
			st.DIZs++
			store.Diz = record.Download{
				Name:    filepath.Base(f.Name()),
				LastMod: f.ModTime(),
			}
		case ".nfo":
			st.NFOs++
			store.Nfo = record.Download{
				// TODO: find most appropriate NFO file
				Name:    filepath.Base(f.Name()),
				LastMod: f.ModTime(),
			}
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

func Group(key string) string {
	s := PathGroup(key)
	switch strings.ToLower(s) {
	case "zwt":
		return zwt.Name
	}
	return s
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
