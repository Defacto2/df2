// Package archive handles collections of files that are either packaged together or compressed.
package archive

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/archive/internal/content"
	"github.com/Defacto2/df2/lib/archive/internal/demozoo"
	"github.com/Defacto2/df2/lib/archive/internal/file"
	"github.com/Defacto2/df2/lib/archive/internal/task"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/images"
	"github.com/Defacto2/df2/lib/logs"
)

const (
	bat  = ".bat"
	bmp  = ".bmp"
	com  = ".com"
	diz  = ".diz"
	exe  = ".exe"
	gif  = ".gif"
	jpg  = ".jpg"
	nfo  = ".nfo"
	png  = ".png"
	tiff = ".tiff"
	txt  = ".txt"
	webp = ".webp"
)

var (
	ErrSameArgs = errors.New("name and dest cannot be the same")
	ErrNotArc   = errors.New("format specified by source filename is not an archive format")
)

func Copy(name, dest string) (written int64, err error) {
	return file.Copy(name, dest)
}

func Move(name, dest string) (written int64, err error) {
	return file.Move(name, dest)
}

// Demozoo decompresses and parses archives fetched from Demozoo.org.
func Demozoo(name, uuid string, varNames *[]string) (demozoo.Data, error) {
	dz := demozoo.Data{}
	if err := database.CheckUUID(uuid); err != nil {
		return demozoo.Data{}, fmt.Errorf("extract demozoo checkuuid %q: %w", uuid, err)
	}
	// create temp dir
	tempDir, err := ioutil.TempDir("", "extarc-")
	if err != nil {
		return demozoo.Data{}, fmt.Errorf("extract demozoo tempdir %q: %w", tempDir, err)
	}
	defer os.RemoveAll(tempDir)
	filename, err := database.LookupFile(uuid)
	if err != nil {
		return demozoo.Data{}, fmt.Errorf("extract demozoo lookup id %q: %w", uuid, err)
	}
	if _, err = Restore(name, filename, tempDir); err != nil {
		return demozoo.Data{}, fmt.Errorf("extract demozoo restore %q: %w", filename, err)
	}
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return demozoo.Data{}, fmt.Errorf("extract demozoo readdir %q: %w", tempDir, err)
	}
	zips := make(content.Contents)
	for i, f := range files {
		var zip content.File
		zip.Path = tempDir // filename gets appended by z.scan()
		zip.Scan(f)
		if err = zip.MIME(); err != nil {
			return demozoo.Data{}, fmt.Errorf("extract demozoo filemime %q: %w", f, err)
		}
		zips[i] = zip
	}
	if nfo := demozoo.NFO(name, zips, varNames); nfo != "" {
		if ok, err := demozoo.MoveText(filepath.Join(tempDir, nfo), uuid); err != nil {
			return demozoo.Data{}, fmt.Errorf("extract demozo move nfo: %w", err)
		} else if !ok {
			dz.NFO = nfo
		}
	}
	if dos := demozoo.DOS(name, zips, varNames); dos != "" {
		dz.DOSee = dos
	}
	return dz, nil
}

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

// Proof decompresses and parses an archive.
// uuid is used to rename the extracted assets such as image previews.
func Proof(src, filename, uuid string) error {
	if err := database.CheckUUID(uuid); err != nil {
		return fmt.Errorf("archive uuid %q: %w", uuid, err)
	}
	// create temp dir
	tempDir, err := ioutil.TempDir("", "proofextract-")
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

// Read returns a list of files within an rar, tar, zip or 7z archive.
// uuid is the absolute path to the archive file named as a unique id.
// filename is the original archive filename and file extension.
func Read(uuid, name string) ([]string, error) {
	files, err := Readr(uuid, name)
	if err != nil {
		return nil, fmt.Errorf("read uuid/filename: %w", err)
	}
	return files, nil
}

// Restore unpacks or decompresses a given archive file to the destination.
// The archive format is selected implicitly. Restore relies on the filename
// extension to determine which decompression format to use, which must be
// supplied using filename.
// uuid is the absolute path to the archive file named as a unique id.
// filename is the original archive filename and file extension.
func Restore(uuid, name, dest string) ([]string, error) {
	err := Unarchiver(uuid, name, dest)
	if err != nil {
		return nil, fmt.Errorf("restore unarchiver: %w", err)
	}
	files, err := Readr(uuid, name)
	if err != nil {
		return nil, fmt.Errorf("restore readr: %w", err)
	}
	return files, nil
}
