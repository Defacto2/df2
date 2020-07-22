package archive

import (
	"fmt"

	"github.com/mholt/archiver"
)

func extractr(archive, filename, tempDir string) error {
	if err := Unarchiver(archive, filename, tempDir); err != nil {
		return err
	}
	return nil
}

// Readr returns a list of files within an rar, tar or zip archive.
// It has offers compatibility with compression formats.
func Readr(archive, filename string) (files []string, err error) {
	err = Walkr(archive, filename, func(f archiver.File) error {
		files = append(files, f.Name())
		return nil
	})
	return files, err
}

// Unarchiver unarchives the given archive file into the destination folder.
// The archive format is selected implicitly.
// Archiver relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func Unarchiver(source, filename, destination string) error {
	f, err := archiver.ByExtension(filename)
	if err != nil {
		return err
	}
	if err := configure(f); err != nil {
		return err
	}
	un, ok := f.(archiver.Unarchiver)
	if !ok {
		return fmt.Errorf("format specified by source filename is not an archive format: %s (%T)", filename, f)
	}
	return un.Unarchive(source, destination)
}

func configure(f interface{}) (err error) {
	tar := &archiver.Tar{
		OverwriteExisting:      true,
		MkdirAll:               true,
		ImplicitTopLevelFolder: true,
		ContinueOnError:        false,
	}
	switch v := f.(type) {
	case *archiver.Rar:
		v.OverwriteExisting = true
		v.MkdirAll = true
		v.ImplicitTopLevelFolder = true
		v.ContinueOnError = false
		//v.Password = os.Getenv("ARCHIVE_PASSWORD")
	case *archiver.Tar:
		// nothing to customize
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
		v.OverwriteExisting = true
		v.MkdirAll = true
		v.SelectiveCompression = true
		v.ImplicitTopLevelFolder = true
		v.ContinueOnError = false
	case *archiver.Gz, *archiver.Bz2, *archiver.Lz4, *archiver.Snappy, *archiver.Xz:
		// nothing to customize
	default:
		err = fmt.Errorf("config does not support customization: %s", f)
	}
	return err
}

// Walkr calls walkFn for each file within the given archive file.
// The archive format is chosen implicitly.
// Archiver relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func Walkr(archive, filename string, walkFn archiver.WalkFunc) error {
	a, err := archiver.ByExtension(filename)
	if err != nil {
		return err
	}
	w, ok := a.(archiver.Walker)
	if !ok {
		return fmt.Errorf("format specified by archive filename is not a walker format: %s (%T)", filename, a)
	}
	return w.Walk(archive, walkFn)
}
