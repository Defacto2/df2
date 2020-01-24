package assets

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"

	"github.com/Defacto2/df2/lib/directories"

	"github.com/dustin/go-humanize"

	_ "github.com/go-sql-driver/mysql" // MySQL database driver
)

// files are unique UUID values used by the database and filenames
type files map[string]struct{}

var (
	empty  = database.Empty{}
	ignore files
	paths  []string // a collection of directories
	d      = directories.Init(false)
)

// AddTarFile saves the result of a fileWalk file into a TAR archive at path as the source file name.
// Source: cloudfoundry/archiver (https://github.com/cloudfoundry/archiver/blob/master/compressor/write_tar.go)
func AddTarFile(path, name string, tw *tar.Writer) error {
	fi, err := os.Lstat(path)
	if err != nil {
		return err
	}
	var link string
	if fi.Mode()&os.ModeSymlink != 0 {
		if link, err = os.Readlink(path); err != nil {
			return err
		}
	}
	hdr, err := tar.FileInfoHeader(fi, link)
	if err != nil {
		return err
	}
	if fi.IsDir() && !os.IsPathSeparator(name[len(name)-1]) {
		name = name + "/"
	}
	if hdr.Typeflag == tar.TypeReg && name == "." {
		// archiving a single file
		hdr.Name = filepath.ToSlash(filepath.Base(path))
	} else {
		hdr.Name = filepath.ToSlash(name)
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if hdr.Typeflag == tar.TypeReg {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err = io.Copy(tw, file); err != nil {
			return err
		}
	}
	return nil
}

type results struct {
	count int   // results handled
	fails int   // results that failed
	bytes int64 // bytes counted
}
type scan struct {
	path   string       // directory to scan
	delete bool         // delete any detected orphan files
	human  bool         // humanize values shown by print output
	m      database.IDs // UUID values fetched from the database
}

// Clean walks through and scans directories containing UUID files and erases any orphans that cannot be matched to the database.
func Clean(target string, delete, human bool) {
	targets(target)
	// connect to the database
	rows, m := CreateUUIDMap()
	logs.Println("The following files do not match any UUIDs in the database")
	// parse directories
	var sum results
	for p := range paths {
		s := scan{path: paths[p], delete: delete, human: human, m: m}
		sum.calculate(s)
	}
	// output a summary of the results
	logs.Println(color.Notice.Sprintf("\nTotal orphaned files discovered %v out of %v", humanize.Comma(int64(sum.count)), humanize.Comma(int64(rows))))
	if sum.fails > 0 {
		logs.Print(fmt.Sprintf("due to errors %v files could not be deleted\n", sum.fails))
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
}

func (sum *results) calculate(s scan) {
	r, err := s.scanPath()
	logs.Log(err)
	sum.bytes += r.bytes
	sum.count += r.count
	sum.fails += r.fails
}

func targets(target string) int {
	if d.Base == "" {
		d = directories.Init(false)
	}
	paths = nil
	switch target {
	case "all":
		paths = append(paths, d.UUID, d.Emu, d.Backup, d.Img000, d.Img400, d.Img150)
	case "download":
		paths = append(paths, d.UUID, d.Backup)
	case "emulation":
		paths = append(paths, d.Emu)
	case "image":
		paths = append(paths, d.Img000, d.Img400, d.Img150)
	}
	return len(paths)
}

// CreateUUIDMap builds a map of all the unique UUID values stored in the Defacto2 database.
func CreateUUIDMap() (int, database.IDs) {
	db := database.Connect()
	defer db.Close()
	// query database
	var id, uuid string
	rows, err := db.Query("SELECT `id`,`uuid` FROM `files`")
	logs.Check(err)
	m := database.IDs{} // this map is to store all the UUID values used in the database
	// handle query results
	rc := 0 // row count
	for rows.Next() {
		err = rows.Scan(&id, &uuid)
		logs.Check(err)
		m[uuid] = database.Empty{} // store record `uuid` value as a key name in the map `m` with an empty value
		rc++
	}
	return rc, m
}

func (s *scan) archive(list []os.FileInfo) map[string]struct{} {
	a := make(map[string]struct{})
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
			a[fn] = empty
		}
	}
	return a
}

// backup is used by scanPath to backup matched orphans.
func backup(s *scan, list []os.FileInfo) {
	archive := s.archive(list)
	// identify which files should be backed up
	part := backupPart()
	if _, ok := part[s.path]; ok {
		t := time.Now().Format("2006-Jan-2-150405") // Mon Jan 2 15:04:05 MST 2006
		name := filepath.Join(d.Backup, fmt.Sprintf("bak-%v-%v.tar", part[s.path], t))
		basepath := s.path
		// create tar archive
		newTar, err := os.Create(name)
		logs.File("directory.backup", err)
		tw := tar.NewWriter(newTar)
		defer tw.Close()
		c := 0
		// walk through `path` and match any files marked for deletion
		// Partial source: https://github.com/cloudfoundry/archiver/blob/master/compressor/write_tar.go
		err = filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			name, err := walkName(basepath, path)
			if err != nil {
				return err
			}
			if _, ok := archive[name]; ok {
				c++
				if c == 1 {
					logs.Print("archiving these files before deletion\n\n")
				}
				return AddTarFile(path, name, tw)
			}
			return nil // no match
		})
		// if backup fails, then abort deletion
		if err != nil || c == 0 {
			// clean up any loose archives
			newTar.Close()
			err := os.Remove(name)
			logs.Check(err)
		}
	}
}

func backupPart() map[string]string {
	b := make(map[string]string)
	b[d.UUID] = "uuid"
	b[d.Img150] = "img-150xthumbs"
	b[d.Img400] = "img-400xthumbs"
	b[d.Img000] = "img-captures"
	return b
}

func walkName(basepath, path string) (string, error) {
	var name string
	var err error
	if os.IsPathSeparator(path[len(path)-1]) {
		name, err = filepath.Rel(basepath, path)
	} else {
		name, err = filepath.Rel(filepath.Dir(basepath), path)
	}
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(name), nil
}

// ignoreList is used by scanPath to filter files that should not be erased.
func ignoreList(path string) files {
	i := make(map[string]struct{})
	i["00000000-0000-0000-0000-000000000000"] = empty
	i["blank.png"] = empty
	if path == d.Emu {
		i["g_drive.zip"] = empty
		i["s_drive.zip"] = empty
		i["u_drive.zip"] = empty
		i["dosee-core.js"] = empty
		i["dosee-core.mem"] = empty
	}
	return i
}

// scanPath gets a list of filenames located in s.path and matches the results against the list generated by CreateUUIDMap.
func (s scan) scanPath() (results, error) {
	logs.Println(color.Primary.Sprintf("\nResults from %v", s.path))
	// query file system
	list, err := ioutil.ReadDir(s.path)
	if err != nil {
		var e *os.PathError
		if errors.As(err, &e) {
			logs.Println(color.Warn.Sprint("no such directory"))
		} else {
			return results{}, err
		}
	}

	// files to ignore
	ignore = ignoreList(s.path)
	// archive files that are to be deleted
	if s.delete {
		backup(&s, list)
	}
	// list and if requested, delete orphaned files
	r := parse(&s, &list)
	var dsc string
	if s.human {
		dsc = humanize.Bytes(uint64(r.bytes))
	} else {
		dsc = fmt.Sprintf("%v B", r.bytes)
	}
	logs.Print(fmt.Sprintf("\n%v orphaned files\n%v drive space consumed\n", r.count, dsc))
	return r, nil // number of orphaned files discovered, deletion failures, their cumulative size in bytes
}
