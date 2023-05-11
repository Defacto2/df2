// Package importer parses specific .NFO and file_id.DIZ file group-packs submitted
// as .rar archives.
package importer

import (
	"archive/zip"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/importer/air"
	"github.com/Defacto2/df2/pkg/importer/amplify"
	"github.com/Defacto2/df2/pkg/importer/arcade"
	"github.com/Defacto2/df2/pkg/importer/arctic"
	"github.com/Defacto2/df2/pkg/importer/hexwars"
	"github.com/Defacto2/df2/pkg/importer/record"
	"github.com/Defacto2/df2/pkg/importer/spirit"
	"github.com/Defacto2/df2/pkg/importer/xdb"
	"github.com/Defacto2/df2/pkg/importer/zone"
	"github.com/Defacto2/df2/pkg/importer/zwt"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/google/uuid"
	"github.com/mholt/archiver"
	"go.uber.org/zap"
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
	Name      string                  // Name of the RAR archive file.
	Group     string                  // Group for the collection.
	GroupPath string                  // GroupPath is the directory path used for the group.
	DestOpen  string                  // DestOpen is the temp destination path for the extracted files.
	DestUUID  string                  // DestUUID is the temp destination path for the created UUID files.
	DIZs      int                     // DIZs file_id.diz count.
	NFOs      int                     // NFOs count.
	Others    int                     // Other types of files count.
	LastMods  Years                   // LastMods counts the last modified years.
	SubDirs   map[string]SubDirectory // Releases lists every release included in the RAR archive.
}

type Importer struct {
	RARFile string // RARFile is the absolute path to the RAR archive file.
	Insert  bool   // Insert records to the database.
	Limit   uint   // Limit the number of subdrectories to process.
	Config  conf.Config
	Logger  *zap.SugaredLogger
}

// Import the named .rar file.
func (im Importer) Import(db *sql.DB, w io.Writer) error {
	if w == nil {
		w = io.Discard
	}
	if err := check(im.RARFile); err != nil {
		return err
	}
	ctx, ticker, limit := context.Background(), time.Now(), im.Limit
	if err := checkDL(im.Config.Downloads); err != nil {
		return err
	}
	if limit > 0 {
		im.Logger.Infof("Only the first %d, randomized subdirectories of the RAR archive will be read.", limit)
	}
	// first stat the .rar file
	st := Stat{}
	if err := st.Walk(im.RARFile, im.Logger); err != nil {
		return err
	}
	// apply the limit
	st.limit(im.Limit)
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
	// Store the subdirectories as UUID archives.
	if err := st.Store(w, im.Logger); err != nil {
		return err
	}
	// Create a collection of database data.
	imports, err := st.Create()
	if err != nil {
		return err
	}
	// Copy archive and insert the data into the database.
	if im.Insert {
		inserts, err := imports.Insert(ctx, db, im.Logger, im.Config.Downloads, limit)
		if err != nil {
			return err
		}
		if inserts > 0 {
			im.Logger.Infof("Inserted %d, non-public records\n"+
				"\tThe df2 'new' and df2 'approve' commands need to be used to make these public", inserts)
		}
	} else {
		im.Logger.Warn("No database records will be created without the --insert flag")
	}
	nerdStats(w, im.Logger, st, limit)
	im.Logger.Infof("Time taken, %v sec.", time.Since(ticker).Seconds())
	return nil
}

func nerdStats(w io.Writer, l *zap.SugaredLogger, st Stat, limit uint) {
	if w == nil {
		return
	}
	l.Info("Statistics for the nerds")
	tw := new(tabwriter.Writer)
	const width = 8
	tw.Init(w, 0, width, 0, '\t', 0)
	fmt.Fprintf(tw, "\t\tThe releases for %q\n", st.Group)
	fmt.Fprint(tw, " \t\tSubDirectories: ", "\t", len(st.SubDirs))
	if limit > 0 {
		fmt.Fprint(tw, " (--limit in use)", "\n")
	} else {
		fmt.Fprint(tw, "\n")
	}
	fmt.Fprint(tw, "\t\tNFO files found: ", "\t", st.NFOs, "\n")
	fmt.Fprint(tw, "\t\tfile_id.diz files found: ", "\t", st.DIZs, "\n")
	fmt.Fprint(tw, "\t\tOther file discoveries: ", "\t", st.Others, "\n")
	fmt.Fprint(tw, "\t\tRange of years published: ", "\t")
	s := []string{}
	for year, cnt := range st.LastMods {
		s = append(s, fmt.Sprintf("%s (%d)", year, cnt))
	}
	fmt.Fprint(tw, strings.Join(s, ", "))
	fmt.Fprint(tw, "\n")
	tw.Flush()
}

func checkDL(name string) error {
	st, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("%w: %s", err, name)
	}
	if !st.IsDir() {
		return fmt.Errorf("%w: %s", ErrDownloads, name)
	}
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

// Limit the subdirectories.
// It is best to read the whole rar archive before asserting the subdirectory limit,
// otherwise some files might go missing.
func (st *Stat) limit(i uint) {
	if i == 0 {
		return
	}
	dirs := 0
	for key := range st.SubDirs {
		dirs++
		if dirs > int(i) {
			delete(st.SubDirs, key)
		}
	}
}

// Walk the named .rar archive file to collect information and statistics.
func (st *Stat) Walk(name string, l *zap.SugaredLogger) error {
	st.SubDirs = map[string]SubDirectory{}
	rar := rar()
	key := ""
	return rar.Walk(name, func(f archiver.File) error {
		if f.IsDir() {
			return nil
		}
		if st.Name == "" {
			st.Name = name
		}
		// find metadata
		base := filepath.Base(f.Name())
		if st.Group == "" {
			st.Group = Group(f.Name())
		}
		if st.GroupPath == "" {
			st.GroupPath = strings.ToLower(PathGroup(f.Name()))
		}
		key = fixKey(key, f.Name())
		// create subdirectory entry
		sub, exists := st.SubDirs[key]
		if !exists {
			sub.Title = str.PathTitle(key)
		}
		sub.Path = root(f.Name())
		relPath := strings.Replace(f.Name(), key+string(os.PathSeparator), "", 1)
		sub.Files = append(sub.Files, relPath)
		sub.LastMods = append(sub.LastMods, f.ModTime())
		uid, err := uuid.NewRandom()
		if err != nil {
			l.Errorf("new random error: %w", err)
			return nil
		}
		sub.UUID = uid.String()
		// find readme text
		ext := strings.ToLower(filepath.Ext(base))
		switch ext {
		case ".diz":
			st.DIZs++
			if strings.ToLower(base) == record.FileID {
				if sub.Readme == "" {
					// the base value with the exact filename case must be used.
					sub.Readme = filepath.Base(base)
				}
			}
		case ".nfo":
			st.NFOs++
			g := fmt.Sprintf("%s.nfo", st.GroupPath)
			if strings.ToLower(base) == g {
				// the base value with the exact filename case must be used.
				sub.Readme = filepath.Base(base)
			}
		default:
			st.Others++
		}
		st.LastMods = st.LastMods.Add(f.ModTime())
		st.SubDirs[key] = sub
		return nil
	})
}

func root(p string) string {
	if p == "" {
		return ""
	}
	return strings.Split(filepath.Dir(p), string(os.PathSeparator))[0]
}

func fixKey(key, name string) string {
	if filepath.Dir(name) == key {
		return key
	}
	key = filepath.Dir(name)
	s := strings.Split(key, string(filepath.Separator))
	if len(s) > 1 {
		key = s[0]
	}
	return strings.TrimSpace(key)
}

// Store creates a collection of UUID named, uncompressed zip archives
// from the extracted sub-directories, of the imported RAR file archive.
func (st *Stat) Store(w io.Writer, l *zap.SugaredLogger) error { //nolint:funlen,gocognit
	if w == nil {
		w = io.Discard
	}
	const noArchive = 1
	i := 0
	// make temp dest
	p, err := os.MkdirTemp(os.TempDir(), output)
	if err != nil {
		return fmt.Errorf("store mkdir: %w", err)
	}
	st.DestUUID = p
	defer os.Remove(p)
	// archive the subdirectories
	l.Infof("Storing %d subdirectories", len(st.SubDirs))
	const width = 8
	for key, sub := range st.SubDirs {
		if len(sub.Files) == 0 {
			return fmt.Errorf("%w: %s", ErrNoFiles, sub.Title)
		}
		i++
		tw := new(tabwriter.Writer)
		tw.Init(w, 0, width, 0, '\t', 0)
		sources := []string{}
		for _, f := range sub.Files {
			sources = append(sources, filepath.Join(st.DestOpen, sub.Path, f))
		}
		name := sub.UUID
		dest := filepath.Join(p, name)
		// read the content of any group nfo or file_id.diz
		for _, src := range sources {
			base := filepath.Base(src)
			g := nfos(st.GroupPath, sub.Path)
			if strings.ToLower(base) == g[0] {
				sub.Nfo, err = os.ReadFile(src)
				if err != nil {
					return fmt.Errorf("store readnfo0: %w", err)
				}
			}
			if len(sub.Nfo) == 0 && strings.ToLower(base) == g[1] {
				sub.Nfo, err = os.ReadFile(src)
				if err != nil {
					return fmt.Errorf("store readnfo1: %w", err)
				}
			}
			if UseDIZ(g[0], base) {
				sub.Diz, err = os.ReadFile(src)
				if err != nil {
					return fmt.Errorf("store readdiz: %w", err)
				}
				if len(sub.Nfo) > 0 && len(sub.Diz) > 0 {
					// as only one of these byte arrays are required,
					// save system memory by deleting the unused one
					sub.Nfo = nil
				}
			}
		}
		// handle subdirectories containing only a single text file
		if len(sub.Files) == noArchive {
			dest = filepath.Join(p, sub.UUID)
			br, err := record.Copy(dest, sources[0])
			if err != nil {
				return err
			}
			err = os.Chtimes(dest, sub.LastMods[0], sub.LastMods[0])
			if err != nil {
				return fmt.Errorf("store chtimes: %w", err)
			}

			sub.Filename = filepath.Base(sources[0])
			fmt.Fprintf(tw, " \t\tTEXT file, %d bytes: %s\n", br, dest)
			tw.Flush()
			// save the changes
			st.SubDirs[key] = sub
			continue
		}
		// process subdirectory
		br, err := sub.Zip(l, dest, sources...)
		if err != nil {
			return err
		}
		sub.Filename = record.Zip(sub.Path)
		fmt.Fprintf(tw, " \t\tZIP archive, %d bytes: %s\n", br, dest)
		tw.Flush()
		// save the changes
		st.SubDirs[key] = sub
	}
	return nil
}

func nfos(groupPath, subPath string) [2]string {
	g := fmt.Sprintf("%s.nfo", groupPath)
	// handle any edge-cases
	switch subPath { //nolint:gocritic
	case `KV331.Synthmaster.2.v2.6.21.MacOSX.Incl.Keyfile-HEXWARS`:
		g = `HEXWARS-OSX.nfo`
	}
	s := [2]string{strings.ToLower(g), ""}
	if groupPath == `air` {
		s[1] = "airiso.nfo"
	}
	return s
}

// Do not read the file_id.diz if these named nfo files are discovered,
// as the included file_id.diz doesn't contain the required metadata.
func UseDIZ(g, base string) bool {
	switch g {
	case `air.nfo`, `airiso.nfo`, `arcade.nfo`, `arctic.nfo`, `arctic (2).nfo`, `xdb.nfo`:
		return false
	}
	return strings.ToLower(base) == record.FileID
}

// Create a collection of records based on the extracted sub-directories,
// of the imported RAR file archive.
func (st *Stat) Create() (records record.Records, err error) { //nolint:nonamedreturns
	records = make(record.Records, len(st.SubDirs))
	i := -1
	for _, meta := range st.SubDirs {
		i++
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
		if d.ReadDate.IsZero() {
			// fallback when there is no file_id.diz or no date within the diz
			err = d.ReadNfo(string(meta.Nfo), st.Group)
			if err != nil {
				return nil, err
			}
		}
		records[i].Title = d.ReadTitle
		if records[i].Title == "" {
			records[i].Title = meta.Title
		}
		records[i].FileName = meta.Filename
		records[i].FileSize = d.Bytes
		records[i].FileMagic = d.Magic
		records[i].HashStrong = d.HashStrong
		records[i].HashWeak = d.HashWeak
		records[i].Published = d.ReadDate
		if len(meta.Files) > 1 {
			records[i].LastMod = time.Now()
			records[i].ZipContent = strings.Join(meta.Files, "\n")
		}
		if len(meta.Files) == 1 {
			records[i].LastMod = d.ModTime
		}
		records[i].Readme = meta.Readme
		records[i].TempPath = name
	}
	return records, nil
}

// Zip creates an uncompressed zip archive at dst using the named file sources.
func (sub SubDirectory) Zip(l *zap.SugaredLogger, dst string, sources ...string) (written int64, err error) { //nolint:nonamedreturns,lll
	arch, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer arch.Close()
	// create a new zip file writer
	zipWr := zip.NewWriter(arch)
	if sub.Path != "" {
		// apply an optional zip file comment
		s := fmt.Sprintf("release directory: %s", sub.Path)
		if err = zipWr.SetComment(s); err != nil {
			l.Warnf("failed to set the optional zip file comment")
		}
	}
	// read, copy and store source files into the zip file writer
	for i, src := range sources {
		f, err := os.Open(src)
		if err != nil {
			return 0, fmt.Errorf("zip open: %w", err)
		}
		defer f.Close()
		// file header for each source
		base := filepath.Base(src)
		fh := zip.FileHeader{
			Name:     base,
			NonUTF8:  true,
			Modified: sub.LastMods[i],
		}
		w, err := zipWr.CreateHeader(&fh)
		if err != nil {
			return 0, fmt.Errorf("zip header: %w", err)
		}
		// copy source to the zip file
		wr, err := io.Copy(w, f)
		if err != nil {
			return 0, fmt.Errorf("zip io copy: %w", err)
		}
		written += wr
	}
	if err = zipWr.Close(); err != nil {
		return 0, fmt.Errorf("zip close: %w", err)
	}
	return written, nil
}

// Group return the proper group title of the key.
func Group(key string) string {
	s := PathGroup(key)
	switch strings.ToLower(s) {
	case "air":
		return air.Name
	case "amplify":
		return amplify.Name
	case "arcade":
		return arcade.Name
	case "arctic":
		return arctic.Name
	case "df2":
		return "Defacto2"
	case "hexwars":
		return hexwars.Name
	case "spirit":
		return spirit.Name
	case "xdb":
		return xdb.Name
	case "zone":
		return zone.Name
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
