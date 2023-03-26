// Package archive handles collections of files that are either packaged together or compressed.
package archive

import (
	"archive/tar"
	"compress/gzip"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/pkg/archive/internal/content"
	"github.com/Defacto2/df2/pkg/archive/internal/demozoo"
	"github.com/Defacto2/df2/pkg/archive/internal/file"
	"github.com/Defacto2/df2/pkg/archive/internal/task"
	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/images"
)

const (
	diz = ".diz"
	nfo = ".nfo"
	txt = ".txt"
)

var (
	ErrArchive = errors.New("format specified by source filename is not an archive format")
	ErrDir     = errors.New("is a directory")
	ErrFile    = errors.New("no such file")
	ErrWriter  = errors.New("writer must be a file object")
)

// Copy copies a file to the destination.
func Copy(name, dest string) (int64, error) {
	return file.Copy(name, dest)
}

// Move copies a file to the destination and then deletes the source.
func Move(name, dest string) (int64, error) {
	return file.Move(name, dest)
}

// Compress a collection of files using gzip and add them to the tar writter.
func Compress(w io.Writer, files []string) error {
	_, okf := w.(*os.File)
	_, okw := w.(*gzip.Writer)
	if !okf && !okw {
		return fmt.Errorf("archive compress %w", ErrWriter)
	}
	gw := gzip.NewWriter(w)
	defer gw.Close()
	nw := tar.NewWriter(gw)
	defer nw.Close()
	var errs error
	for _, f := range files {
		if err := file.Add(nw, f); err != nil {
			errs = errors.Join(errs, fmt.Errorf("archive compress: %w", err))
		}
	}
	return errs
}

// Delete removes the collection of files from the host filesystem.
func Delete(files []string) error {
	var errs error
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			errs = errors.Join(errs, fmt.Errorf("archive delete: %w", err))
		}
	}
	return errs
}

// Store adds collection of files to the tar writer.
func Store(w io.Writer, files []string) error {
	_, okf := w.(*os.File)
	_, okw := w.(*tar.Writer)
	if !okf && !okw {
		return fmt.Errorf("archive store %w", ErrWriter)
	}
	tw := tar.NewWriter(w)
	defer tw.Close()
	var errs error
	for _, f := range files {
		if err := file.Add(tw, f); err != nil {
			errs = errors.Join(errs, fmt.Errorf("archive store %w", err))
		}
	}
	return errs
}

type Demozoo struct {
	Source   string
	UUID     string
	VarNames *[]string
	Config   conf.Config
}

// Demozoo decompresses and parses archives fetched from https://demozoo.org.
func (z Demozoo) Decompress(db *sql.DB, w io.Writer) (demozoo.Data, error) {
	if db == nil {
		return demozoo.Data{}, database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	dz := demozoo.Data{}
	if err := database.CheckUUID(z.UUID); err != nil {
		return demozoo.Data{}, fmt.Errorf("extract demozoo checkuuid %q: %w", z.UUID, err)
	}
	// create temp dir
	tmp, err := os.MkdirTemp("", "extarc-")
	if err != nil {
		return demozoo.Data{}, fmt.Errorf("extract demozoo tempdir %q: %w", tmp, err)
	}
	defer os.RemoveAll(tmp)
	name := filepath.Base(z.Source)
	if z.UUID != database.TestID {
		name, err = database.GetFile(db, z.UUID)
		if err != nil {
			return demozoo.Data{}, fmt.Errorf("extract demozoo lookup id %q: %w", z.UUID, err)
		}
	}
	if _, err = Restore(w, z.Source, name, tmp); err != nil {
		return demozoo.Data{}, fmt.Errorf("extract demozoo restore %q: %w", name, err)
	}

	zips, err := zips(tmp)
	if err != nil {
		return demozoo.Data{}, err
	}
	if nfo := demozoo.NFO(z.Source, zips, z.VarNames); nfo != "" {
		if z.Source == "" {
			dz.NFO = nfo
		} else if err := demozoo.MoveText(w, z.Config, filepath.Join(tmp, nfo), z.UUID); err != nil {
			return demozoo.Data{}, fmt.Errorf("extract demozo move nfo: %w", err)
		}
	}
	if dos := demozoo.DOS(w, z.Source, zips, z.VarNames); dos != "" {
		dz.DOSee = dos
	}
	return dz, nil
}

func zips(name string) (content.Contents, error) {
	files, err := os.ReadDir(name)
	if err != nil {
		return nil, fmt.Errorf("extract demozoo readdir %q: %w", name, err)
	}
	zips := make(content.Contents)
	for i, file := range files {
		f, err := file.Info()
		if err != nil {
			fmt.Fprintf(os.Stdout, "extract demozoo file info error: %s\n", err)
		}
		var zip content.File
		zip.Path = name // filename gets appended by z.scan()
		zip.Scan(f)
		if err = zip.MIME(); err != nil {
			return nil, fmt.Errorf("extract demozoo filemime %q: %w", f, err)
		}
		zips[i] = zip
	}
	return zips, nil
}

// NFO attempts to discover a archive package NFO or information textfile from a collection of files.
// For better results the name of the archive file should be provided.
func NFO(name string, files ...string) string {
	f := make(demozoo.Finds)
	for _, file := range files {
		base := strings.TrimSuffix(name, filepath.Ext(name))
		fname := strings.ToLower(file)
		ext := filepath.Ext(fname)
		switch ext {
		case diz, nfo, txt:
			// okay
		default:
			continue
		}
		f = nfoFile(f, file, fname, base, ext)
	}
	return f.Top()
}

func nfoFile(f demozoo.Finds, file, name, base, ext string) demozoo.Finds {
	switch {
	case name == base+".nfo":
		// [archive name].nfo
		f[file] = demozoo.Lvl1
	case name == base+".txt":
		// [archive name].txt
		f[file] = demozoo.Lvl2
	case ext == ".nfo":
		// [random].nfo
		f[file] = demozoo.Lvl3
	case name == "file_id.diz":
		// BBS file description
		f[file] = demozoo.Lvl4
	case name == base+".diz":
		// [archive name].diz
		f[file] = demozoo.Lvl5
	case name == ".txt":
		// [random].txt
		f[file] = demozoo.Lvl6
	case name == ".diz":
		// [random].diz
		f[file] = demozoo.Lvl7
	default:
		// currently lacking is [group name].nfo and [group name].txt priorities
	}
	return f
}

type Proof struct {
	Source string
	Name   string
	UUID   string
	Config conf.Config
}

// Proof decompresses and parses a hosted file archive.
// src is the path to the file including the uuid filename.
// filename is the original archive filename, usually kept in the database.
// uuid is used to rename the extracted assets such as image previews.
func (p Proof) Decompress(w io.Writer) error {
	if w == nil {
		w = io.Discard
	}
	if err := database.CheckUUID(p.UUID); err != nil {
		return fmt.Errorf("archive uuid %q: %w", p.UUID, err)
	}
	// create temp dir
	tmp, err := os.MkdirTemp("", "proofextract-")
	if err != nil {
		return fmt.Errorf("archive tempdir %q: %w", tmp, err)
	}
	defer os.RemoveAll(tmp)
	if err = Unarchiver(p.Source, tmp, p.Name); err != nil {
		return fmt.Errorf("unarchiver: %w", err)
	}
	th, tx, err := task.Run(tmp)
	if err != nil {
		return err
	}
	if n := th.Name; n != "" {
		if err := images.Generate(w, n, p.UUID, true); err != nil {
			return fmt.Errorf("archive generate img: %w", err)
		}
	}
	if n := tx.Name; n != "" {
		f, err := directories.Files(p.Config, p.UUID)
		if err != nil {
			return fmt.Errorf("archive config: %w", err)
		}
		if f.UUID != database.TestID {
			if _, err := file.Move(n, f.UUID+txt); err != nil {
				return fmt.Errorf("archive filemove %q: %w", n, err)
			}
		}
		fmt.Fprint(w, "  »txt")
	}
	if x := true; !x {
		if err := file.Dir(w, tmp); err != nil {
			return fmt.Errorf("archive dir %q: %w", tmp, err)
		}
	}
	return nil
}

// Read returns both a list of files within an rar, tar, zip or 7z archive;
// as-well as a suitable filename string for the archive. This filename is
// useful when the original archive filename has been given an invalid file
// extension.
// src is the absolute path to the archive file named as a unique id.
// name is the original archive filename and file extension.
func Read(w io.Writer, src, name string) ([]string, string, error) {
	if w == nil {
		w = io.Discard
	}
	st, err := os.Stat(src)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, "", fmt.Errorf("read %s: %w", filepath.Base(src), ErrFile)
	}
	if st.IsDir() {
		return nil, "", fmt.Errorf("read %s: %w", filepath.Base(src), ErrDir)
	}
	files, fname, err := Readr(w, src, name)
	if err != nil {
		return nil, "", fmt.Errorf("read uuid/filename: %w", err)
	}
	return files, fname, nil
}

// Restore unpacks or decompresses a given archive file to the destination.
// The archive format is selected implicitly. Restore relies on the filename
// extension to determine which decompression format to use, which must be
// supplied using filename.
// src is the absolute path to the archive file named as a unique id.
// filename is the original archive filename and file extension.
func Restore(w io.Writer, src, name, dest string) ([]string, error) {
	if w == nil {
		w = io.Discard
	}
	err := Unarchiver(src, dest, name)
	if err != nil {
		return nil, fmt.Errorf("restore unarchiver: %w", err)
	}
	files, _, err := Readr(w, src, name)
	if err != nil {
		return nil, fmt.Errorf("restore readr: %w", err)
	}
	return files, nil
}
