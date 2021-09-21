package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/mholt/archiver"
	"github.com/nwaples/rardecode"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Extractor extracts a file from the given archive file into the destination folder.
// The archive format is selected implicitly.
// Archiver relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func Extractor(source, filename, extract, destination string) error {
	filename = strings.ToLower(filename)
	f, err := archiver.ByExtension(filename)
	if err != nil {
		return fmt.Errorf("extractor byextension %q: %w", filename, err)
	}
	if err := configure(f); err != nil {
		return fmt.Errorf("extractor configure: %w", err)
	}
	e, ok := f.(archiver.Extractor)
	if !ok {
		return fmt.Errorf("extractor %s (%T): %w", filename, f, ErrNotArc)
	}
	if err := e.Extract(source, extract, destination); err != nil {
		return fmt.Errorf("extractor: %w", err)
	}
	return nil
}

// Readr returns a list of files within an rar, tar or zip archive.
// It has offers compatibility with compression formats.
func Readr(archive, filename string) ([]string, error) {
	files := []string{}
	err := walkr(archive, filename, func(f archiver.File) error {
		if f.IsDir() {
			return nil
		}
		fn := ""
		switch h := f.Header.(type) {
		case zip.FileHeader:
			fn = h.Name
		case *tar.Header:
			fn = h.Name
		case *rardecode.FileHeader:
			fn = h.Name
		default:
			fn = f.Name()
		}
		b := []byte(fn)
		if utf8.Valid(b) {
			files = append(files, fn)
			return nil
		}
		// handle cheecky DOS era filenames with CP437 extended characters.
		r := transform.NewReader(bytes.NewReader(b), charmap.CodePage437.NewDecoder())
		result, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		files = append(files, string(result))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("readr: %w", err)
	}
	return files, nil
}

// Unarchiver unarchives the given archive file into the destination folder.
// The archive format is selected implicitly.
// Archiver relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func Unarchiver(source, filename, destination string) error {
	f, err := archiver.ByExtension(filename)
	if err != nil {
		return fmt.Errorf("unarchiver byextension %q: %w", filename, err)
	}
	if err := configure(f); err != nil {
		return fmt.Errorf("unarchiver configure: %w", err)
	}
	un, ok := f.(archiver.Unarchiver)
	if !ok {
		return fmt.Errorf("unarchiver %s (%T): %w", filename, f, ErrNotArc)
	}
	if err := un.Unarchive(source, destination); err != nil {
		return fmt.Errorf("unarchiver: %w", err)
	}
	return nil
}

func configure(f interface{}) (err error) {
	cfg := &archiver.Tar{
		OverwriteExisting:      true,
		MkdirAll:               true,
		ImplicitTopLevelFolder: false,
		ContinueOnError:        false,
	}
	switch v := f.(type) {
	case *archiver.Rar:
		// options: https://pkg.go.dev/github.com/mholt/archiver?tab=doc#Rar
		v.OverwriteExisting = true
		v.MkdirAll = true
		v.ImplicitTopLevelFolder = false
		v.ContinueOnError = false
	case *archiver.Tar:
		// options: https://pkg.go.dev/github.com/mholt/archiver?tab=doc#Tar
		v.OverwriteExisting = true
		v.MkdirAll = true
		v.ImplicitTopLevelFolder = false
		v.ContinueOnError = false
	case *archiver.TarBz2:
		v.Tar = cfg
	case *archiver.TarGz:
		v.Tar = cfg
	case *archiver.TarLz4:
		v.Tar = cfg
	case *archiver.TarSz:
		v.Tar = cfg
	case *archiver.TarXz:
		v.Tar = cfg
	case *archiver.Zip:
		// options: https://pkg.go.dev/github.com/mholt/archiver?tab=doc#Zip
		v.OverwriteExisting = true
		v.MkdirAll = true
		v.SelectiveCompression = true
		v.ImplicitTopLevelFolder = false
		v.ContinueOnError = false
	case *archiver.Gz,
		*archiver.Bz2,
		*archiver.Lz4,
		*archiver.Snappy,
		*archiver.Xz:
		// nothing to customise
	default:
		err = fmt.Errorf("configure %v: %w", f, ErrNoCustom)
		return err
	}
	return nil
}

// walkr calls walkFn for each file within the given archive file.
// The archive format is chosen implicitly.
// Archiver relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func walkr(archive, filename string, walkFn archiver.WalkFunc) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("walkr paniced with %s in archive %s: %v\n", filename, filepath.Base(archive), r)
		}
	}()

	filename = strings.ToLower(filename)
	a, err := archiver.ByExtension(filename)
	if err != nil {
		return fmt.Errorf("walkr byextension %q: %w", filename, err)
	}
	w, ok := a.(archiver.Walker)
	if !ok {
		return fmt.Errorf("walkr %s (%T): %w", filename, a, ErrWalkrFmt)
	}
	if err := w.Walk(archive, walkFn); err != nil {
		return fmt.Errorf("walkr %q: %w", filepath.Base(archive), err)
	}
	return nil
}
