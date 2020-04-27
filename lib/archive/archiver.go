package archive

import (
	"fmt"

	"github.com/mholt/archiver"
)

func extractr(archive, filename, tempDir string) error {
	err := Unarchiver(archive, filename, tempDir)
	if err != nil {
		return arErr(err)
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
	uaIface, err := archiver.ByExtension(filename)
	if err != nil {
		return err
	}
	u, ok := uaIface.(archiver.Unarchiver)
	if !ok {
		return fmt.Errorf("format specified by source filename is not an archive format: %s (%T)", filename, uaIface)
	}
	return u.Unarchive(source, destination)
}

// Walkr calls walkFn for each file within the given archive file.
// The archive format is chosen implicitly.
// Archiver relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func Walkr(archive, filename string, walkFn archiver.WalkFunc) error {
	wIface, err := archiver.ByExtension(filename)
	if err != nil {
		return err
	}
	w, ok := wIface.(archiver.Walker)
	if !ok {
		return fmt.Errorf("format specified by archive filename is not a walker format: %s (%T)", filename, wIface)
	}
	return w.Walk(archive, walkFn)
}

func arErr(err error) error {
	return fmt.Errorf("archiver extract: %v", err)
}
