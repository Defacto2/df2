package archive

import (
	"fmt"

	"github.com/mholt/archiver"
)

func extractr(archive, filename, tempDir string) error {
	if err := Unarchiver(archive, filename, tempDir); err != nil {
		return fmt.Errorf("extractr: %w", err)
	}
	return nil
}

// Readr returns a list of files within an rar, tar or zip archive.
// It has offers compatibility with compression formats.
func Readr(archive, filename string) (files []string, err error) {
	err = walkr(archive, filename, func(f archiver.File) error {
		files = append(files, f.Name())
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

// TODO: test ALL archive types
func configure(f interface{}) (err error) {
	tar := &archiver.Tar{
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
		// see tar var
	case *archiver.TarBz2:
		v.Tar = tar
	case *archiver.TarGz:
		v.Tar = tar
	case *archiver.TarLz4:
		v.Tar = tar
	case *archiver.TarSz:
		v.Tar = tar
	case *archiver.TarXz:
		v.Tar = tar
	case *archiver.Zip:
		// options: https://pkg.go.dev/github.com/mholt/archiver?tab=doc#Zip
		v.OverwriteExisting = true
		v.MkdirAll = true
		v.SelectiveCompression = true
		v.ImplicitTopLevelFolder = false
		v.ContinueOnError = false
	case *archiver.Gz, *archiver.Bz2, *archiver.Lz4, *archiver.Snappy, *archiver.Xz:
		// nothing to customize
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
	a, err := archiver.ByExtension(filename)
	if err != nil {
		return fmt.Errorf("walkr byextension %q: %w", filename, err)
	}
	w, ok := a.(archiver.Walker)
	if !ok {
		return fmt.Errorf("walkr %s (%T): %w", filename, a, ErrWalkrFmt)
	}
	if err := w.Walk(archive, walkFn); err != nil {
		return fmt.Errorf("walkr %q: %w", archive, err)
	}
	return nil
}
