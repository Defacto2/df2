// Package importer parses specific .NFO and file_id.DIZ file group-packs submitted
// as .rar archives.
package importer

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/importer/record"
	"github.com/Defacto2/df2/pkg/importer/zwt"
	"github.com/google/uuid"
	"github.com/mholt/archiver"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"

	models "github.com/Defacto2/df2/pkg/models/mysql"
)

var (
	ErrDir       = errors.New("named file points to a directory")
	ErrDownloads = errors.New("downloads directory points to a file")
	ErrNoFiles   = errors.New("release contains no files")
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

type Importer struct {
	RARFile string
	Insert  bool
	Limit   uint
	Config  conf.Config
	Logger  *zap.SugaredLogger
}

// Import the named .rar file.
func (im Importer) Import(db *sql.DB, w io.Writer) error {
	if err := check(im.RARFile); err != nil {
		return err
	}
	if w == nil {
		w = io.Discard
	}

	ticker := time.Now()
	ctx := context.Background()
	downloads := im.Config.Downloads
	limit := im.Limit

	// todo: move to internal/args?
	if x, err := os.Stat(downloads); err != nil {
		return fmt.Errorf("%w: %s", err, downloads)
	} else if !x.IsDir() {
		return fmt.Errorf("%w: %s", ErrDownloads, downloads)
	}
	if im.Limit > 0 {
		fmt.Fprintf(w, "\n%d item limit applied to the total number of found text files to process.\n", limit)
	}

	// first stat the .rar file
	st := Stat{}
	if err := st.Walk(im.RARFile, limit); err != nil {
		return err
	}

	// if okay, then uncompress it to a tmpdir
	dest, err := os.MkdirTemp(os.TempDir(), input)
	if err != nil {
		return err
	}
	st.DestOpen = dest
	defer os.RemoveAll(dest)

	// Unarchive prints useless errors, but is much faster than using archiver.Extract().
	// so set SetOutput to discard all logged errors.
	log.SetOutput(io.Discard)
	rar := rar()
	if err := rar.Unarchive(im.RARFile, dest); err != nil {
		return err
	}
	log.SetOutput(os.Stderr)

	if err := st.Store(nil, limit); err != nil {
		return err
	}
	imports, err := st.Create(w, limit)
	if err != nil {
		return err
	}

	// TODO:
	// copy file to uuid dir
	// insert record

	for i, rec := range imports {
		if limit > 0 && i > int(limit) {
			break
		}
		// TODO: make funcs
		cnt, err := models.Files(qm.Where("file_integrity_strong=?", rec.HashStrong)).Count(ctx, db)
		if err != nil {
			return err
		}
		if cnt != 0 {
			fmt.Fprintf(w, "Skipped %q as the file hash matches an existing database record", rec.Title)
			continue
		}
		b, err := json.MarshalIndent(rec, "", " ")
		if err != nil {
			return err
		}

		if !im.Insert {
			continue
		}

		newpath := filepath.Join(im.Config.Downloads, rec.UUID.String)
		if err := os.Rename(rec.TempPath, newpath); err != nil {
			return err
		}

		/*
			type Record struct {
				UUID       string    `json:"uuid"`
				Slug       string    `json:"slug"`
				Title      string    `json:"record_title"`
				Group      string    `json:"group_brand_for"`
				FileName   string    `json:"filename"`
				FileSize   int64     `json:"filesize"`
				FileMagic  string    `json:"file_magic_type"`
				HashStrong string    `json:"file_integrity_strong"`
				HashWeak   string    `json:"file_integrity_weak"`
				LastMod    time.Time `json:"file_last_modified"`
				Published  time.Time `json:"date_issued"`
				Section    string    `json:"section"`
				Platform   string    `json:"platform"`
				Comment    string    `json:"comment"`
				TempPath   string    `json:"temp_path"` // TempPath to the temporary UUID named file download.
			}
		*/

		f1 := models.File{}
		f1.UUID = rec.UUID
		f1.RecordTitle = null.NewString(rec.Title, true)
		f1.GroupBrandFor = rec.Group
		if !rec.Published.IsZero() {
			f1.DateIssuedYear = null.Int16From(int16(rec.Published.Year()))
			f1.DateIssuedMonth = null.Int8From(int8(rec.Published.Month()))
			f1.DateIssuedDay = null.Int8From(int8(rec.Published.Day()))
		}
		f1.Filename = null.NewString(rec.FileName, true)
		f1.Filesize = null.NewInt(int(rec.FileSize), true)
		f1.FileMagicType = null.NewString(rec.FileMagic, true)
		//		f1.FileZipContent = null.NewString() ZipContent
		f1.FileIntegrityStrong = null.NewString(rec.HashStrong, true)
		f1.FileIntegrityWeak = null.NewString(rec.HashWeak, true)
		f1.FileLastModified = null.NewTime(rec.LastMod, true)
		f1.Platform = null.NewString(rec.Platform, true)
		f1.Section = null.NewString(rec.Section, true)
		f1.Comment = null.NewString(rec.Comment, true)
		// DeletedAt
		// UpdatedBy
		// retrotxt_readme

		err = f1.Insert(ctx, db, boil.Infer()) // Insert the first pilot with name "Larry"
		if err != nil {
			defer os.Remove(newpath)
			return err
		}
		// p1 now has an ID field set to 1

		// todo:
		// check temppath
		// mv file
		// insert data
		// if db throws error, delete copied file

		fmt.Fprintf(w, "\n%d.\t%s\n", i, string(b))
		fmt.Fprintf(w, "File:\t%s", rec.TempPath)
	}

	if !im.Insert {
		fmt.Fprintln(w, "\nno records are to be inserted without the --insert flag")
	}

	fmt.Fprintln(w, "\nStats for nerds, totals")
	fmt.Fprintln(w, "subdirs\t", len(st.SubDirs))
	fmt.Fprintln(w, "nfos\t", st.NFOs)
	fmt.Fprintln(w, "dizes\t", st.DIZs)
	fmt.Fprintln(w, "other files\t", st.Others)
	fmt.Fprintln(w, "group\t", st.Group)
	fmt.Fprintf(w, "years: %+v\n", st.LastMods)

	fmt.Fprintln(w, "time taken", time.Since(ticker).Seconds())
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
func (st *Stat) Walk(name string, limit uint) error {
	st.SubDirs = map[string]SubDirectory{}
	rar := rar()
	key := ""
	count := -1

	return rar.Walk(name, func(f archiver.File) error {
		if st.Name == "" {
			st.Name = name
		}
		if f.IsDir() {
			return nil
		}
		count++
		if count > int(limit) {
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

// Store creates a collection of UUID named, uncompressed zip archives
// from the extracted sub-directories, of the imported RAR file archive.
func (st *Stat) Store(w io.Writer, limit uint) error {
	if w == nil {
		w = io.Discard
	}
	const noArchive = 1
	i := 0

	p, err := os.MkdirTemp(os.TempDir(), output)
	if err != nil {
		return err
	}
	st.DestUUID = p

	for key, sub := range st.SubDirs {
		if len(sub.Files) == 0 {
			return fmt.Errorf("%w: %s", ErrNoFiles, sub.Title)
		}
		fmt.Println(i, ">", limit)
		fmt.Println(key, sub)
		fmt.Println()
		if limit > 0 && uint(i) >= limit {
			fmt.Println("=== BREAK")
			break
		}

		i++
		sources := []string{}
		for _, f := range sub.Files {
			sources = append(sources, filepath.Join(st.DestOpen, sub.Path, f))
		}
		name := sub.UUID
		dest := filepath.Join(p, name)

		fmt.Println("------>", dest)

		// read the content of any group nfo or file_id.diz
		for _, src := range sources {
			base := filepath.Base(src)
			g := fmt.Sprintf("%s.nfo", strings.ToLower(st.Group))
			if strings.ToLower(base) == g {
				sub.Nfo, err = os.ReadFile(src)
				if err != nil {
					return err
				}
			}
			if strings.ToLower(base) == record.FileID {
				sub.Diz, err = os.ReadFile(src)
				if err != nil {
					return err
				}
			}
		}

		if len(sub.Files) == noArchive {
			dest = filepath.Join(p, sub.UUID)
			_, err := record.Copy(dest, sources[0])
			if err != nil {
				return err
			}
			err = os.Chtimes(dest, sub.LastMods[0], sub.LastMods[0])
			if err != nil {
				return err
			}
			sub.Filename = filepath.Base(sources[0])
			fmt.Fprintf(w, "new text file with %d bytes, %s\n", w, dest)
			continue
		}

		_, err := sub.Zip(dest, sources...)
		if err != nil {
			return err
		}
		sub.Filename = record.Zip(sub.Path)
		fmt.Fprintf(w, "new zip archive with %d bytes, %s\n", w, dest)

		// apply the changes
		st.SubDirs[key] = sub
	}
	return nil
}

// Create a collection of records based on the extracted sub-directories,
// of the imported RAR file archive.
func (st *Stat) Create(w io.Writer, limit uint) (records Records, err error) {
	if w == nil {
		w = io.Discard
	}

	records = make(Records, len(st.SubDirs))
	c, l := 0, len(records)
	for key, meta := range st.SubDirs {
		c++
		i := c - 1
		if limit > 0 && uint(i) >= limit {
			fmt.Fprintf(w, " Item limit argument %d of %d reached\n", i, limit)
			break
		}

		fmt.Println("---------->", key)
		fmt.Println("--->>", meta.UUID)

		records[i], err = record.New(meta.UUID, meta.Path, st.Group)
		if err != nil {
			return nil, err
		}

		d := record.Download{}
		name := filepath.Join(st.DestUUID, meta.UUID)
		err = d.Create(name, st.Group)
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
		records[i].TempPath = name
		fmt.Fprintf(w, "%d-%d.\t new record %q\n", c, l, key)
	}
	return records, nil
}

// Zip creates an uncompressed zip archive at dst using the the named file sources.
func (sub SubDirectory) Zip(dst string, sources ...string) (written int64, err error) {
	arch, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer arch.Close()

	zipWriter := zip.NewWriter(arch)
	if sub.Path != "" {
		_ = zipWriter.SetComment(fmt.Sprintf("release directory: %s", sub.Path))
	}

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

// Group return the proper group title of the key.
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
