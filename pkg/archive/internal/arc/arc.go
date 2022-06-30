package arc

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
)

var (
	ErrNoCustom = errors.New("does not support customization")
	ErrWalkrFmt = errors.New("format specified by archive filename is not a walker format")
)

// Configure interface for the archiver,
// a cross-platform, multi-format archive utility and Go library.
func Configure(f any) error {
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
		return fmt.Errorf("configure %v: %w", f, ErrNoCustom)
	}
	return nil
}

// Walkr calls walkFn for each file within the given archive file.
// The archive format is chosen implicitly.
// Archiver relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func Walkr(src, filename string, walkFn archiver.WalkFunc) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("walkr paniced with %s in archive %s: %v\n", filename, filepath.Base(src), r)
		}
	}()
	filename = strings.ToLower(filename)
	a, err := archiver.ByExtension(filename)
	if err != nil {
		return fmt.Errorf("walkr byextension %q: %w", filename, err)
	}
	if w, ok := a.(archiver.Walker); !ok {
		return fmt.Errorf("walkr %s (%T): %w", filename, a, ErrWalkrFmt)
	} else if err := w.Walk(src, walkFn); err != nil {
		return fmt.Errorf("walkr %q: %w", filepath.Base(src), err)
	}
	return nil
}
