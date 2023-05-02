// Package importer parses specific .NFO and file_id.DIZ file group-packs submitted
// as .rar archives.
package importer

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/importer/record"
	"github.com/Defacto2/df2/pkg/importer/zwt"
	"github.com/google/uuid"
	"github.com/mholt/archiver"
)

var (
	ErrDir     = errors.New("named file points to a directory")
	ErrNoFiles = errors.New("release contains no files")
)

const (
	input  = "dizzer-input"
	output = "dizzer-dest"
)

// Records are a collection of Record items to insert into the database.
type Records []record.Record

// SubDirectory of the RAR archive.
type SubDirectory struct {
	Title    string      // Title for the subdirectory.
	Diz      []byte      // The body of an included file_id.diz file.
	Nfo      []byte      // The body of a root information text file.
	Files    []string    // Files named that are included in the subdirectory.
	LastMods []time.Time // The file last modification times of the Files.
	Publish  time.Time   // The earliest file last modification time found in the subdirectory.
	Path     string      // Path of the subdirectory.
	Readme   string      // Readme is the filename of the information file for display.
	UUID     string      // UUID is the new unique id for the subdirectory.
	Filename string      // Filename for the UUID download.
}

// Stat the collection of NFO and file_id.diz files within an RAR archive.
type Stat struct {
	Name     string                  // Name of the RAR archive file.
	Group    string                  // Group for the collection.
	DestOpen string                  // DestOpen is the temp destination path for the extracted files.
	DestUUID string                  // DestUUID is the temp destination path for the created UUID files.
	DIZs     int                     // DIZs file_id.diz count.
	NFOs     int                     // NFOs count.
	Others   int                     // Other types of files count.
	LastMods Years                   // LastMods counts the last modified years.
	SubDirs  map[string]SubDirectory // Releases lists every release included in the RAR archive.
}

// Import the named .rar file.
func Import(w, logger io.Writer, name string) error {
	if err := check(name); err != nil {
		return err
	}
	if w == nil {
		w = io.Discard
	}
	if logger == nil {
		logger = io.Discard
	}

	// time taken
	ticker := time.Now()

	// first stat the .rar file
	st := Stat{}
	if err := st.Walk(name); err != nil {
		return err
	}

	// if okay, then uncompress it to a tmpdir
	dest, err := os.MkdirTemp(os.TempDir(), input)
	if err != nil {
		return err
	}
	//defer os.RemoveAll(dest)
	st.DestOpen = dest

	// Unarchive prints useless errors but is much faster than using archiver.Extract(),
	// so instead discard any logged errors.
	log.SetOutput(io.Discard)
	rar := rar()
	if err := rar.Unarchive(name, dest); err != nil {
		return err
	}
	log.SetOutput(os.Stderr)

	files, err := st.Store(nil)
	if err != nil {
		return err
	}
	recs, err := st.New(logger, files)
	if err != nil {
		return err
	}

	fmt.Println("\nJSON marshal indent =", len(recs))
	// for i, r := range records {
	// 	i++
	// 	u, err := json.MarshalIndent(r, "", " ")
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fmt.Fprintf(w, "%d.\t%+v of %s\n", i, r.Title, r.FileName)
	// 	fmt.Fprintln(w, string(u))
	// }

	// TODO: check hash in DB
	// copy file to uuid dir
	// insert record

	fmt.Fprintln(w, "\nStats for nerds, totals")
	fmt.Fprintln(w, "subdirs\t", len(st.SubDirs))
	fmt.Fprintln(w, "nfos\t", st.NFOs)
	fmt.Fprintln(w, "dizes\t", st.DIZs)
	fmt.Fprintln(w, "other files\t", st.Others)
	fmt.Fprintln(w, "group\t", st.Group)
	fmt.Fprintf(w, "years: %+v\n", st.LastMods)

	fmt.Fprintln(logger, "time taken", time.Since(ticker).Seconds())
	return nil
}

func rar() archiver.Rar {
	return archiver.Rar{
		OverwriteExisting:      true,
		MkdirAll:               true,
		ImplicitTopLevelFolder: false,
		ContinueOnError:        true,
	}
}

func check(name string) error {
	rar := rar()
	if err := rar.CheckExt(name); err != nil {
		return fmt.Errorf("%w: %s", err, name)
	}
	if st, err := os.Stat(name); err != nil {
		return fmt.Errorf("%w: %s", err, name)
	} else if st.IsDir() {
		return fmt.Errorf("%w: %s", err, name)
	}
	return nil
}

// Walk the named .rar archive file to collect information and statistics.
func (st *Stat) Walk(name string) error {
	st.SubDirs = map[string]SubDirectory{}
	rar := rar()
	key := ""

	return rar.Walk(name, func(f archiver.File) error {
		if st.Name == "" {
			st.Name = name
		}
		if f.IsDir() {
			return nil
		}

		base := filepath.Base(f.Name())
		if st.Group == "" {
			st.Group = Group(f.Name())
		}
		if filepath.Dir(f.Name()) != key {
			key = filepath.Dir(f.Name())
		}

		sub, exists := st.SubDirs[key]
		if !exists {
			sub.Title = PathTitle(key)
		}
		sub.Path = filepath.Dir(f.Name())
		sub.Files = append(sub.Files, base)
		sub.LastMods = append(sub.LastMods, f.ModTime())

		uid, err := uuid.NewRandom()
		if err != nil {
			fmt.Println(err)
			return nil
		}
		sub.UUID = uid.String()

		ext := strings.ToLower(filepath.Ext(base))
		switch ext {
		case ".diz":
			st.DIZs++
			if strings.ToLower(base) == record.FileID {
				if sub.Readme == "" {
					sub.Readme = filepath.Base(base)
				}
			}
		case ".nfo":
			st.NFOs++
			g := fmt.Sprintf("%s.nfo", strings.ToLower(st.Group))
			if strings.ToLower(base) == g {
				sub.Readme = filepath.Base(g)
			}
		default:
			st.Others++
		}
		st.LastMods = st.LastMods.Add(f.ModTime())
		st.SubDirs[key] = sub
		return nil
	})
}

func (st *Stat) Store(w io.Writer) (int, error) {
	if w == nil {
		w = io.Discard
	}
	const noArchive = 1
	i := 0

	p, err := os.MkdirTemp(os.TempDir(), output)
	if err != nil {
		return 0, err
	}
	st.DestUUID = p

	for key, sub := range st.SubDirs {
		if len(sub.Files) == 0 {
			return 0, fmt.Errorf("%w: %s", ErrNoFiles, sub.Title)
		}

		i++
		sources := []string{}
		for _, f := range sub.Files {
			sources = append(sources, filepath.Join(st.DestOpen, sub.Path, f))
		}
		name := sub.UUID
		dest := filepath.Join(p, name)

		// read the content of any group nfo or file_id.diz
		for _, src := range sources {
			base := filepath.Base(src)
			g := fmt.Sprintf("%s.nfo", strings.ToLower(st.Group))
			if strings.ToLower(base) == g {
				sub.Nfo, err = os.ReadFile(src)
				if err != nil {
					return 0, err
				}
			}
			if strings.ToLower(base) == record.FileID {
				sub.Diz, err = os.ReadFile(src)
				if err != nil {
					return 0, err
				}
			}
		}

		if len(sub.Files) == noArchive {
			dest = filepath.Join(p, sub.UUID)
			_, err := record.Copy(dest, sources[0])
			if err != nil {
				return 0, err
			}
			err = os.Chtimes(dest, sub.LastMods[0], sub.LastMods[0])
			if err != nil {
				return 0, err
			}
			sub.Filename = filepath.Base(sources[0])
			fmt.Fprintf(w, "new text file with %d bytes, %s\n", w, dest)
			continue
		}

		_, err := sub.Zip(dest, sources...)
		if err != nil {
			return 0, err
		}
		sub.Filename = record.Zip(sub.Path)
		fmt.Fprintf(w, "new zip archive with %d bytes, %s\n", w, dest)

		// apply the changes
		st.SubDirs[key] = sub
	}
	return i, nil
}

func (st *Stat) New(logger io.Writer, stored int) (Records, error) {
	var err error
	records := make(Records, stored)
	c, l := 0, len(records)
	fmt.Printf("records %d-%d\n", c, l)
	for key, meta := range st.SubDirs {
		c++
		if c != l {
			continue
		}
		i := c - 1

		records[i], err = record.New(meta.UUID, meta.Path, st.Group)
		if err != nil {
			return nil, err
		}

		d := record.Download{}
		name := filepath.Join(st.DestUUID, meta.UUID)
		err = d.New(name, st.Group)
		if err != nil {
			return nil, err
		}

		err = d.ReadDIZ(string(meta.Diz), st.Group)
		if err != nil {
			return nil, err
		}

		records[i].Title = d.ReadTitle
		if records[i].Title == "" {
			records[i].Title = meta.Title
		}
		records[i].FileName = meta.Filename // this needs to be set earliy either zip filename or nfo filename
		records[i].FileSize = d.Bytes
		records[i].FileMagic = d.Magic
		records[i].HashStrong = d.HashStrong
		records[i].HashWeak = d.HashWeak
		records[i].LastMod = time.Now()
		records[i].Published = d.ReadDate

		fmt.Printf("META = %+v\n\n", meta)
		fmt.Printf("%s\n%s\n\n", key, meta.Diz)
		fmt.Printf("RECORD = %+v\n\n%+v", records[i], d)
		// match lastmod date to the earliest
		// if diz.LastMod.Before(nfo.LastMod) {
		// 	nfo.LastMod = diz.LastMod
		// } else if nfo.LastMod.Before(diz.LastMod) {
		// 	diz.LastMod = nfo.LastMod
		// }
		// // copy readdate to nfo
		// if nfo.ReadDate.IsZero() && !diz.ReadDate.IsZero() {
		// 	nfo.ReadDate = diz.ReadDate
		// }
	}
	return records, nil
}

func (sub SubDirectory) Zip(dst string, sources ...string) (written int64, err error) {
	arch, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer arch.Close()

	zipWriter := zip.NewWriter(arch)
	for i, src := range sources {
		f, err := os.Open(src)
		if err != nil {
			return 0, err
		}
		defer f.Close()

		base := filepath.Base(src)

		fh := zip.FileHeader{
			Name:     base,
			NonUTF8:  true,
			Modified: sub.LastMods[i],
		}
		w, err := zipWriter.CreateHeader(&fh)
		if err != nil {
			return 0, err
		}

		wr, err := io.Copy(w, f)
		if err != nil {
			return 0, err
		}
		written = written + wr
	}
	err = zipWriter.Close()
	if err != nil {
		return 0, err
	}
	return written, nil
}

func Group(key string) string {
	s := PathGroup(key)
	switch strings.ToLower(s) {
	case "df2":
		return "Defacto2"
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

// PathGroup returns the group name or initialism extracted from the
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
