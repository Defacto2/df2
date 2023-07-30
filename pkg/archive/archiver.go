package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/Defacto2/df2/pkg/archive/internal/arc"
	"github.com/Defacto2/df2/pkg/archive/internal/sys"
	"github.com/mholt/archiver"
	"github.com/nwaples/rardecode"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Extractor extracts the named file from the given archive file into the destination folder.
// The archive format is selected implicitly.
// Archiver relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func Extractor(name, src, target, dest string) error {
	name = strings.ToLower(name)
	f, err := archiver.ByExtension(name)
	if err != nil {
		return fmt.Errorf("extractor byextension %q: %w", name, err)
	}
	if err := arc.Configure(f); err != nil {
		return fmt.Errorf("extractor configure: %w", err)
	}
	e, ok := f.(archiver.Extractor)
	if !ok {
		return fmt.Errorf("extractor %s (%T): %w", name, f, ErrArchive)
	}

	// recover from panic caused by mholt/archiver.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("extractor panic %s: %v", name, r)
		}
	}()

	if err := e.Extract(src, target, dest); err != nil {
		// second attempt at extraction using a system archiver program
		if err := sys.Extract(name, src, target, dest); err != nil {
			return fmt.Errorf("extractor, system attempt: %w", err)
		}
		return fmt.Errorf("extractor: %w", err)
	}
	return nil
}

// Readr returns both a list of files within an rar, tar or zip archive,
// and a suitable archive filename string.
// If there are problems reading the archive due to an incorrect filename
// extension, the returned filename string will be corrected.
func Readr(w io.Writer, src, filename string) ([]string, string, error) {
	if w == nil {
		w = io.Discard
	}
	if files, err := readr(src, filename); err == nil {
		return files, filename, nil
	}
	files, ext, err := sys.Readr(w, src, filename)
	if errors.Is(err, sys.ErrWrongExt) {
		newname := sys.Rename(ext, filename)
		fmt.Fprintf(w, "rename to %s; ", newname)
		files, err = readr(src, newname)
		if err != nil {
			return nil, "", fmt.Errorf("readr fix: %w", err)
		}
		return files, newname, nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("readr: %w", err)
	}
	return files, filename, nil
}

func readr(src, filename string) ([]string, error) {
	files := []string{}
	return files, arc.Walkr(src, filename, func(f archiver.File) error {
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
		// handle cheeky DOS era filenames with CP437 extended characters.
		r := transform.NewReader(bytes.NewReader(b), charmap.CodePage437.NewDecoder())
		result, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		files = append(files, string(result))
		return nil
	})
}

// Unarchiver decompresses the given archive file into the destination folder.
// The archive format is selected implicitly.
// Archiver relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func Unarchiver(src, dest, filename string) error {
	f, err := archiver.ByExtension(filename)
	if err != nil {
		return fmt.Errorf("unarchiver byextension %q: %w", filename, err)
	}
	if err := arc.Configure(f); err != nil {
		return fmt.Errorf("unarchiver configure: %w", err)
	}
	un, ok := f.(archiver.Unarchiver)
	if !ok {
		return fmt.Errorf("unarchiver %s (%T): %w", filename, f, ErrArchive)
	}
	if err := un.Unarchive(src, dest); err != nil {
		return fmt.Errorf("unarchiver: %w", err)
	}
	return nil
}
