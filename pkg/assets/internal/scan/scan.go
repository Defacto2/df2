package scan

import (
	"archive/tar"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Defacto2/df2/pkg/assets/internal/file"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
)

var (
	ErrDir       = errors.New("directory to backup is missing")
	ErrDirBck    = errors.New("directory to save tar backup files is missing")
	ErrStructNil = errors.New("structure cannot be nil")
	ErrPathEmpty = errors.New("path cannot be empty")
	ErrTarget    = errors.New("unknown target")
)

type Scan struct {
	Path   string       // directory to scan
	Delete bool         // delete any detected orphan files
	Human  bool         // humanise values shown by print output
	M      database.IDs // UUID values fetched from the database
}

type Results struct {
	Count int   // Results handled
	Fails int   // Results that failed
	Bytes int64 // bytes counted
}

type (
	// Files are unique UUID values used by the database and filenames.
	Files map[string]struct{}
	part  map[string]string
)

func (sum *Results) Calculate(s Scan, d *directories.Dir) error {
	r, err := s.scanPath(d)
	if err != nil {
		return fmt.Errorf("sum calculate: %w", err)
	}
	sum.Bytes += r.Bytes
	sum.Count += r.Count
	sum.Fails += r.Fails
	return nil
}

func (s *Scan) archive(list []os.FileInfo, ignore Files) Files {
	a := make(Files)
	for _, file := range list {
		if file.IsDir() {
			continue // ignore directories
		}
		if _, file := ignore[file.Name()]; file {
			continue // ignore files
		}
		fn := file.Name()
		id := strings.TrimSuffix(fn, filepath.Ext(fn))
		// search the map `m` for `UUID`, the result is saved as a boolean to `exists`
		_, exists := s.M[id]
		if !exists {
			a[fn] = database.Empty{}
		}
	}
	return a
}

// scanPath gets a list of filenames located in s.Path and matches the Results
// against the list generated by CreateUUIDMap.
func (s Scan) scanPath(d *directories.Dir) (Results, error) {
	logs.Println(color.Primary.Sprintf("\nResults from %v", s.Path))
	// query file system
	entries, err := os.ReadDir(s.Path)
	if err != nil {
		var e *os.PathError
		if !errors.As(err, &e) {
			return Results{}, fmt.Errorf("scan path readdir %q: %w", s.Path, err)
		}
		logs.Println(color.Warn.Sprint("assets scanpath: no such directory"))
	}
	list := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return Results{}, fmt.Errorf("entry info: %w", err)
		}
		list = append(list, info)
	}
	// files to ignore
	ignore := IgnoreList(s.Path, d)
	// archive files that are to be deleted
	if s.Delete {
		if err = Backup(&s, d, ignore, list); err != nil {
			return Results{}, fmt.Errorf("scan path backup: %w", err)
		}
	}
	// list and if requested, delete orphaned files
	stat, err := parse(&s, ignore, &list)
	if err != nil {
		return stat, fmt.Errorf("scan path parse: %w", err)
	}
	dsc := fmt.Sprintf("%v B", stat.Bytes)
	if s.Human {
		dsc = humanize.Bytes(uint64(stat.Bytes))
	}
	logs.Print(fmt.Sprintf("\n%v orphaned files\n%v drive space consumed\n", stat.Count, dsc))
	// number of orphaned files discovered, deletion failures, their cumulative size in bytes
	return stat, nil
}

// IgnoreList is used by scanPath to filter files that should not be erased.
func IgnoreList(path string, d *directories.Dir) Files {
	empty := database.Empty{}
	ignore := make(Files)
	ignore["00000000-0000-0000-0000-000000000000"] = empty
	ignore["blank.png"] = empty
	if path == d.Emu {
		ignore["g_drive.zip"] = empty
		ignore["s_drive.zip"] = empty
		ignore["u_drive.zip"] = empty
		ignore["dosee-core.js"] = empty
		ignore["dosee-core.mem"] = empty
	}
	return ignore
}

// Backup is used by scanPath to backup matched orphans.
func Backup(s *Scan, d *directories.Dir, ignore Files, list []os.FileInfo) error {
	if s == nil {
		return fmt.Errorf("backup s (scan): %w", ErrStructNil)
	}
	if s.Path == "" {
		return fmt.Errorf("backup: %w", ErrPathEmpty)
	}
	var test bool
	if flag.Lookup("test.v") != nil {
		test = true
	}
	f := s.archive(list, ignore)
	// identify which files should be backed up
	d, p := backupParts(d)
	if test {
		p[s.Path] = "test"
	}
	if _, ok := p[s.Path]; ok {
		if err := s.backupPart(f, d, p, test); err != nil {
			return fmt.Errorf("backup part: %w", err)
		}
	}
	return nil
}

func backupParts(d *directories.Dir) (*directories.Dir, part) {
	p := make(part)
	p[d.UUID] = "uuid"
	p[d.Img400] = "img-400xthumbs"
	p[d.Img000] = "img-captures"
	return d, p
}

func (s *Scan) backupPart(f Files, d *directories.Dir, p part, test bool) error {
	t := time.Now().Format("2006-Jan-2-150405") // Mon Jan 2 15:04:05 MST 2006
	tarName, basepath := filepath.Join(d.Backup, fmt.Sprintf("bak-%v-%v.tar", p[s.Path], t)), s.Path
	_, err := os.Stat(basepath)
	if os.IsNotExist(err) {
		return fmt.Errorf("create %q: %w", basepath, ErrDir)
	}
	_, err = os.Stat(d.Backup)
	if os.IsNotExist(err) {
		return fmt.Errorf("create %q: %w", d.Backup, ErrDirBck)
	}
	// create tar archive
	newTar, err := os.Create(tarName)
	if err != nil {
		return fmt.Errorf("create %q: %w", tarName, err)
	}
	tw := tar.NewWriter(newTar)
	defer tw.Close()
	c := 0
	// walk through `path` and match any files marked for deletion
	// Partial source: https://github.com/cloudfoundry/archiver/blob/master/compressor/write_tar.go
	err = filepath.Walk(s.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk %q: %w", path, err)
		}
		wlkname, err := file.WalkName(basepath, path)
		if err != nil {
			return fmt.Errorf("walk name %q: %w", path, err)
		}
		if _, ok := f[wlkname]; ok || test {
			c++
			if c == 1 {
				logs.Print("archiving these files before deletion\n\n")
			}
			if err := file.WriteTar(path, tarName, tw); err != nil {
				return fmt.Errorf("write tar %q: %w", path, err)
			}
		}
		return nil // no match
	})
	// if backup fails, then abort deletion
	if err != nil || c == 0 {
		// clean up any loose archives
		newTar.Close()
		if err := os.Remove(tarName); err != nil {
			return fmt.Errorf("cleanup remove %q: %w", tarName, err)
		}
	}
	return nil
}

type item struct {
	name  string // os.FileInfo.Name()
	path  string // filepath
	flag  string // check tick, mark or blank
	human bool   // humanise sizes
	cnt   string // loop count
	fm    string // file mode
	fs    string // file size
	mt    string // file modified time
}

// parse is used by scanPath to remove matched orphans.
func parse(s *Scan, ignore Files, list *[]os.FileInfo) (Results, error) {
	const padding = 2
	stat := Results{Count: 0, Fails: 0, Bytes: 0}
	for _, file := range *list {
		if file.IsDir() {
			continue // ignore directories
		}
		if _, ign := ignore[file.Name()]; ign {
			continue // ignore files
		}
		i := item{human: s.Human, name: file.Name()}
		uuid := strings.TrimSuffix(i.name, filepath.Ext(i.name))
		w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
		// search the map `m` for `UUID`, the result is saved as a boolean to `exists`
		_, exists := s.M[uuid]
		if !exists {
			stat.totals(file)
			if s.Delete {
				i.path = path.Join(s.Path, file.Name())
				i.erase(stat)
			}
			i.count(stat.Count)
			i.mod(file)
			i.size(file)
			i.bits(file)
			fmt.Fprintf(w, "%v\t%v %v\t%v\t%v\t%v\n", i.cnt, i.flag, i.name, i.fs, i.fm, i.mt)
		}
		if err := w.Flush(); err != nil {
			return stat, fmt.Errorf("parse tabwriter flush: %w", err)
		}
	}
	return stat, nil
}

func (i *item) bits(f os.FileInfo) {
	i.fm = color.Note.Sprint(f.Mode())
}

func (i *item) count(c int) {
	i.cnt = color.Secondary.Sprint(strconv.Itoa(c) + ".")
}

func (i *item) erase(r Results) {
	i.flag = str.Y()
	if err := os.Remove(i.path); err != nil {
		i.flag = str.X()
		r.Fails++
	}
}

func (i *item) mod(f os.FileInfo) {
	s := fmt.Sprint(f.ModTime())
	if i.human {
		// show date and time
		s = f.ModTime().Format("02 Jan 15:04")
		if time.Now().Year() != f.ModTime().Year() {
			// otherwise, hide the time and show the year
			s = f.ModTime().Format("02 Jan 2006")
		}
	}
	i.mt = color.Primary.Sprint(s)
}

func (i *item) size(f os.FileInfo) {
	s := fmt.Sprint(f.Size())
	if i.human {
		s = humanize.Bytes(uint64(f.Size()))
	}
	i.fs = color.Comment.Sprint(s)
}

func (sum *Results) totals(f os.FileInfo) {
	sum.Count++
	sum.Bytes += f.Size()
}
