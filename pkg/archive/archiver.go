package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"unicode/utf8"

	"github.com/Defacto2/df2/pkg/archive/internal/arc"
	"github.com/mholt/archiver"
	"github.com/nwaples/rardecode"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Extractor extracts a file from the given archive file into the destination folder.
// The archive format is selected implicitly.
// Archiver relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func Extractor(src, filename, target, dest string) error {
	filename = strings.ToLower(filename)
	f, err := archiver.ByExtension(filename)
	if err != nil {
		return fmt.Errorf("extractor byextension %q: %w", filename, err)
	}
	if err := arc.Configure(f); err != nil {
		return fmt.Errorf("extractor configure: %w", err)
	}
	e, ok := f.(archiver.Extractor)
	if !ok {
		return fmt.Errorf("extractor %s (%T): %w", filename, f, ErrNotArc)
	}
	if err := e.Extract(src, target, dest); err != nil {
		return fmt.Errorf("extractor: %w", err)
	}
	return nil
}

// Readr returns a list of files within an rar, tar or zip archive.
// It has offers compatibility with compression formats.
func Readr(src, filename string) ([]string, error) {
	files := []string{}
	err := arc.Walkr(src, filename, func(f archiver.File) error {
		if f.IsDir() {
			return nil
		}
		var fn string
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
func Unarchiver(src, filename, dest string) error {
	f, err := archiver.ByExtension(filename)
	if err != nil {
		return fmt.Errorf("unarchiver byextension %q: %w", filename, err)
	}
	if err := arc.Configure(f); err != nil {
		return fmt.Errorf("unarchiver configure: %w", err)
	}
	un, ok := f.(archiver.Unarchiver)
	if !ok {
		return fmt.Errorf("unarchiver %s (%T): %w", filename, f, ErrNotArc)
	}
	if err := un.Unarchive(src, dest); err != nil {
		return fmt.Errorf("unarchiver: %w", err)
	}
	return nil
}
