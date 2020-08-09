package assets

import (
	"archive/tar"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	_ "github.com/go-sql-driver/mysql" // MySQL database driver
	"github.com/gookit/color"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
)

type Target int

const (
	All Target = iota
	Download
	Emulation
	Image
)

type (
	// files are unique UUID values used by the database and filenames.
	files map[string]struct{}
	part  map[string]string
)

var (
	ignore files
	paths  []string // a collection of directories
)

var (
	ErrStructNil = errors.New("structure cannot be nil")
	ErrPathEmpty = errors.New("path cannot be empty")
	ErrTarget    = errors.New("unknown target")
)

// Clean walks through and scans directories containing UUID files
// and erases any orphans that cannot be matched to the database.
func Clean(target string, remove, human bool) error {
	d := directories.Init(false)
	return clean(targetfy(target), d, remove, human)
}

// CreateUUIDMap builds a map of all the unique UUID values stored in the Defacto2 database.
func CreateUUIDMap() (total int, uuids database.IDs, err error) {
	db := database.Connect()
	defer db.Close()
	// count rows
	count := 0
	if err = db.QueryRow("SELECT COUNT(*) FROM `files`").Scan(&count); err != nil {
		return 0, nil, fmt.Errorf("create uuid map query row: %w", err)
	}
	// query database
	var id, uuid string
	rows, err := db.Query("SELECT `id`,`uuid` FROM `files`")
	if err != nil {
		return 0, nil, fmt.Errorf("create uuid map query: %w", err)
	}
	defer rows.Close()
	uuids = make(database.IDs, count)
	for rows.Next() {
		if err = rows.Scan(&id, &uuid); err != nil {
			return 0, nil, fmt.Errorf("create uuid map row: %w", err)
		}
		// store record `uuid` value as a key name in the map `m` with an empty value
		uuids[uuid] = database.Empty{}
		total++
	}
	return total, uuids, db.Close()
}

// backup is used by scanPath to backup matched orphans.
func backup(s *scan, d directories.Dir, list []os.FileInfo) error {
	if s == nil {
		return fmt.Errorf("backup s (scan): %w", ErrStructNil)
	} else if s.path == "" {
		return fmt.Errorf("backup: %w", ErrPathEmpty)
	}
	var test bool
	if flag.Lookup("test.v") != nil {
		test = true
	}
	f := s.archive(list)
	// identify which files should be backed up
	d, p := backupParts(d)
	if test {
		p[s.path] = "test"
	}
	if _, ok := p[s.path]; ok {
		if err := s.backupPart(f, d, p, test); err != nil {
			return fmt.Errorf("backup part: %w", err)
		}
	}
	return nil
}

func (s *scan) backupPart(f files, d directories.Dir, p part, test bool) error {
	t := time.Now().Format("2006-Jan-2-150405") // Mon Jan 2 15:04:05 MST 2006
	name, basepath := filepath.Join(d.Backup, fmt.Sprintf("bak-%v-%v.tar", p[s.path], t)), s.path
	// create tar archive
	newTar, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("create %q: %w", name, err)
	}
	tw := tar.NewWriter(newTar)
	defer tw.Close()
	c := 0
	// walk through `path` and match any files marked for deletion
	// Partial source: https://github.com/cloudfoundry/archiver/blob/master/compressor/write_tar.go
	err = filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk %q: %w", path, err)
		}
		name, err = walkName(basepath, path)
		if err != nil {
			return fmt.Errorf("walk name %q: %w", path, err)
		}
		if _, ok := f[name]; ok || test {
			c++
			if c == 1 {
				logs.Print("archiving these files before deletion\n\n")
			}
			if err := writeTar(path, name, tw); err != nil {
				return fmt.Errorf("write tar %q: %w", path, err)
			}
		}
		return nil // no match
	})
	// if backup fails, then abort deletion
	if err != nil || c == 0 {
		// clean up any loose archives
		newTar.Close()
		if err := os.Remove(name); err != nil {
			return fmt.Errorf("remove %q: %w", name, err)
		}
	}
	return nil
}

func backupParts(d directories.Dir) (directories.Dir, part) {
	p := make(part)
	p[d.UUID] = "uuid"
	p[d.Img150] = "img-150xthumbs"
	p[d.Img400] = "img-400xthumbs"
	p[d.Img000] = "img-captures"
	return d, p
}

func clean(t Target, d directories.Dir, remove, human bool) error {
	if ts := targets(t, d); ts == nil {
		return fmt.Errorf("check target %q: %w", t, ErrTarget)
	}
	fmt.Printf("mess: %+v", d)
	// connect to the database
	rows, m, err := CreateUUIDMap()
	if err != nil {
		return fmt.Errorf("clean uuid map: %w", err)
	}
	logs.Println("The following files do not match any UUIDs in the database")
	// parse directories
	var sum results
	for p := range paths {
		s := scan{path: paths[p], delete: remove, human: human, m: m}
		if err := sum.calculate(s, d); err != nil {
			return fmt.Errorf("clean sum calculate: %w", err)
		}
	}
	// output a summary of the results
	logs.Println(color.Notice.Sprintf("\nTotal orphaned files discovered %v out of %v",
		humanize.Comma(int64(sum.count)), humanize.Comma(int64(rows))))
	if sum.fails > 0 {
		logs.Print(fmt.Sprintf("assets clean: due to errors %v files could not be deleted\n", sum.fails))
	}
	if len(paths) > 1 && sum.bytes > 0 {
		var pts string
		if human {
			pts = humanize.Bytes(uint64(sum.bytes))
		} else {
			pts = fmt.Sprintf("%v B", sum.bytes)
		}
		logs.Print(fmt.Sprintf("%v drive space consumed\n", pts))
	}
	return nil
}

// ignoreList is used by scanPath to filter files that should not be erased.
func ignoreList(path string, d directories.Dir) (ignore files) {
	var empty = database.Empty{}
	ignore = make(files)
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

func targetfy(t string) Target {
	switch strings.ToLower(t) {
	case "all":
		return All
	case "download":
		return Download
	case "emulation":
		return Emulation
	case "image":
		return Image
	}
	return -1
}

func targets(t Target, d directories.Dir) []string {
	if d.Base == "" {
		d = directories.Init(false)
	}
	paths = nil
	switch t {
	case All:
		paths = append(paths, d.UUID, d.Emu, d.Backup, d.Img000, d.Img400, d.Img150)
	case Download:
		paths = append(paths, d.UUID, d.Backup)
	case Emulation:
		paths = append(paths, d.Emu)
	case Image:
		paths = append(paths, d.Img000, d.Img400, d.Img150)
	}
	return paths
}

func walkName(basepath, path string) (name string, err error) {
	if path == "" {
		return "", fmt.Errorf("walkname: %w", ErrPathEmpty)
	}
	if os.IsPathSeparator(path[len(path)-1]) {
		name, err = filepath.Rel(basepath, path)
	} else {
		name, err = filepath.Rel(filepath.Dir(basepath), path)
	}
	if err != nil {
		return "", fmt.Errorf("walkname rel-path: %w", err)
	}
	return filepath.ToSlash(name), nil
}

// writeTar saves the result of a fileWalk into a TAR writer.
// Source: cloudfoundry/archiver
// https://github.com/cloudfoundry/archiver/blob/master/compressor/write_tar.go
func writeTar(absPath, filename string, tw *tar.Writer) error {
	stat, err := os.Lstat(absPath)
	if err != nil {
		return fmt.Errorf("writetar %q:%w", absPath, err)
	}
	var link string
	if stat.Mode()&os.ModeSymlink != 0 {
		if link, err = os.Readlink(absPath); err != nil {
			return fmt.Errorf("writetar mode:%w", err)
		}
	}
	head, err := tar.FileInfoHeader(stat, link)
	if err != nil {
		return fmt.Errorf("writetar header:%w", err)
	}
	if stat.IsDir() && !os.IsPathSeparator(filename[len(filename)-1]) {
		filename += "/"
	}
	if head.Typeflag == tar.TypeReg && filename == "." {
		// archiving a single file
		head.Name = filepath.ToSlash(filepath.Base(absPath))
	} else {
		head.Name = filepath.ToSlash(filename)
	}
	if err := tw.WriteHeader(head); err != nil {
		return fmt.Errorf("writetar write header:%w", err)
	}
	if head.Typeflag == tar.TypeReg {
		file, err := os.Open(absPath)
		if err != nil {
			return fmt.Errorf("writetar open %q:%w", absPath, err)
		}
		defer file.Close()
		if _, err = io.Copy(tw, file); err != nil {
			return fmt.Errorf("writetar io.copy %q:%w", absPath, err)
		}
	}
	return nil
}

type results struct {
	count int   // results handled
	fails int   // results that failed
	bytes int64 // bytes counted
}

func (sum *results) calculate(s scan, d directories.Dir) error {
	r, err := s.scanPath(d)
	if err != nil {
		return fmt.Errorf("sum calculate: %w", err)
	}
	sum.bytes += r.bytes
	sum.count += r.count
	sum.fails += r.fails
	return nil
}

type scan struct {
	path   string       // directory to scan
	delete bool         // delete any detected orphan files
	human  bool         // humanize values shown by print output
	m      database.IDs // UUID values fetched from the database
}

func (s *scan) archive(list []os.FileInfo) files {
	a := make(files)
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
		_, exists := s.m[id]
		if !exists {
			a[fn] = database.Empty{}
		}
	}
	return a
}

// scanPath gets a list of filenames located in s.path and matches the results
// against the list generated by CreateUUIDMap.
func (s scan) scanPath(d directories.Dir) (stat results, err error) {
	logs.Println(color.Primary.Sprintf("\nResults from %v", s.path))
	// query file system
	list, err := ioutil.ReadDir(s.path)
	if err != nil {
		var e *os.PathError
		if errors.As(err, &e) {
			logs.Println(color.Warn.Sprint("assets scanpath: no such directory"))
		} else {
			return stat, fmt.Errorf("scan path readdir %q: %w", s.path, err)
		}
	}
	// files to ignore
	ignore = ignoreList(s.path, d)
	// archive files that are to be deleted
	if s.delete {
		if err = backup(&s, d, list); err != nil {
			return stat, fmt.Errorf("scan path backup: %w", err)
		}
	}
	// list and if requested, delete orphaned files
	stat, err = parse(&s, &list)
	if err != nil {
		return stat, fmt.Errorf("scan path parse: %w", err)
	}
	var dsc string
	if s.human {
		dsc = humanize.Bytes(uint64(stat.bytes))
	} else {
		dsc = fmt.Sprintf("%v B", stat.bytes)
	}
	logs.Print(fmt.Sprintf("\n%v orphaned files\n%v drive space consumed\n", stat.count, dsc))
	// number of orphaned files discovered, deletion failures, their cumulative size in bytes
	return stat, nil
}
