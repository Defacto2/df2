// Package archive handles collections of files that are either packaged together or compressed.
package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/pkg/archive/internal/content"
	"github.com/Defacto2/df2/pkg/archive/internal/demozoo"
	"github.com/Defacto2/df2/pkg/archive/internal/file"
	"github.com/Defacto2/df2/pkg/archive/internal/task"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/images"
	"github.com/Defacto2/df2/pkg/logs"
)

const (
	diz = ".diz"
	nfo = ".nfo"
	txt = ".txt"
)

var (
	ErrDir    = errors.New("is a directory")
	ErrFile   = errors.New("no such file")
	ErrNotArc = errors.New("format specified by source filename is not an archive format")
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
func Compress(files []string, buf io.Writer) []error {
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	w := tar.NewWriter(gw)
	defer w.Close()

	errs := []error{}
	for _, f := range files {
		if err := file.Add(w, f); err != nil {
			errs = append(errs, fmt.Errorf("compress: %w", err))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Delete removes the collection of files from the host filesystem.
func Delete(files []string) []error {
	errs := []error{}
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			errs = append(errs, fmt.Errorf("delete: %w", err))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Store adds collection of files to the tar writter.
func Store(files []string, buf io.Writer) []error {
	tw := tar.NewWriter(buf)
	defer tw.Close()

	errs := []error{}
	for _, f := range files {
		if err := file.Add(tw, f); err != nil {
			errs = append(errs, fmt.Errorf("store: %w", err))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Demozoo decompresses and parses archives fetched from https://demozoo.org.
func Demozoo(src, uuid string, varNames *[]string) (demozoo.Data, error) {
	dz := demozoo.Data{}
	if err := database.CheckUUID(uuid); err != nil {
		return demozoo.Data{}, fmt.Errorf("extract demozoo checkuuid %q: %w", uuid, err)
	}
	// create temp dir
	tempDir, err := os.MkdirTemp("", "extarc-")
	if err != nil {
		return demozoo.Data{}, fmt.Errorf("extract demozoo tempdir %q: %w", tempDir, err)
	}
	defer os.RemoveAll(tempDir)
	filename, err := database.GetFile(uuid)
	if err != nil {
		return demozoo.Data{}, fmt.Errorf("extract demozoo lookup id %q: %w", uuid, err)
	}
	if _, err = Restore(src, filename, tempDir); err != nil {
		return demozoo.Data{}, fmt.Errorf("extract demozoo restore %q: %w", filename, err)
	}
	zips, err := zips(tempDir)
	if err != nil {
		return demozoo.Data{}, err
	}
	if nfo := demozoo.NFO(src, zips, varNames); nfo != "" {
		if src == "" {
			dz.NFO = nfo
		} else if err := demozoo.MoveText(filepath.Join(tempDir, nfo), uuid); err != nil {
			return demozoo.Data{}, fmt.Errorf("extract demozo move nfo: %w", err)
		}
	}
	if dos := demozoo.DOS(src, zips, varNames); dos != "" {
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
		fn := strings.ToLower(file)
		ext := filepath.Ext(fn)
		switch ext {
		case diz, nfo, txt:
			// okay
		default:
			continue
		}
		switch {
		case fn == base+".nfo": // [archive name].nfo
			f[file] = demozoo.Lvl1
		case fn == base+".txt": // [archive name].txt
			f[file] = demozoo.Lvl2
		case ext == ".nfo": // [random].nfo
			f[file] = demozoo.Lvl3
		case fn == "file_id.diz": // BBS file description
			f[file] = demozoo.Lvl4
		case fn == base+".diz": // [archive name].diz
			f[file] = demozoo.Lvl5
		case fn == ".txt": // [random].txt
			f[file] = demozoo.Lvl6
		case fn == ".diz": // [random].diz
			f[file] = demozoo.Lvl7
		default: // currently lacking is [group name].nfo and [group name].txt priorities
		}
	}
	return f.Top()
}

// Proof decompresses and parses a hosted file archive.
// src is the path to the file including the uuid filename.
// filename is the original archive filename, usually kept in the database.
// uuid is used to rename the extracted assets such as image previews.
func Proof(src, filename, uuid string) error {
	if err := database.CheckUUID(uuid); err != nil {
		return fmt.Errorf("archive uuid %q: %w", uuid, err)
	}
	// create temp dir
	tempDir, err := os.MkdirTemp("", "proofextract-")
	if err != nil {
		return fmt.Errorf("archive tempdir %q: %w", tempDir, err)
	}
	defer os.RemoveAll(tempDir)
	if err = Unarchiver(src, filename, tempDir); err != nil {
		return fmt.Errorf("unarchiver: %w", err)
	}
	th, tx, err := task.Run(tempDir)
	if err != nil {
		return err
	}
	if n := th.Name; n != "" {
		if err := images.Generate(n, uuid, true); err != nil {
			return fmt.Errorf("archive generate img: %w", err)
		}
	}
	if n := tx.Name; n != "" {
		f := directories.Files(uuid)
		if _, err := file.Move(n, f.UUID+txt); err != nil {
			return fmt.Errorf("archive filemove %q: %w", n, err)
		}
		logs.Print("  »txt")
	}
	if x := true; !x {
		if err := file.Dir(tempDir); err != nil {
			return fmt.Errorf("archive dir %q: %w", tempDir, err)
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
func Read(src, name string) ([]string, string, error) {
	if info, err := os.Stat(src); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("read %s: %w", filepath.Base(src), ErrFile)
	} else if info.IsDir() {
		return nil, "", fmt.Errorf("read %s: %w", filepath.Base(src), ErrDir)
	}
	files, filename, err := Readr(src, name)
	if err != nil {
		return nil, "", fmt.Errorf("read uuid/filename: %w", err)
	}
	return files, filename, nil
}

// Restore unpacks or decompresses a given archive file to the destination.
// The archive format is selected implicitly. Restore relies on the filename
// extension to determine which decompression format to use, which must be
// supplied using filename.
// src is the absolute path to the archive file named as a unique id.
// filename is the original archive filename and file extension.
func Restore(src, name, dest string) ([]string, error) {
	err := Unarchiver(src, name, dest)
	if err != nil {
		return nil, fmt.Errorf("restore unarchiver: %w", err)
	}
	files, _, err := Readr(src, name)
	if err != nil {
		return nil, fmt.Errorf("restore readr: %w", err)
	}
	return files, nil
}
