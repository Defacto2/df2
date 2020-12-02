// Package text generates images from text files using the Ansilove/C program.
package text

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/dustin/go-humanize"
	"github.com/gookit/color"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/images"
)

var (
	ErrNoSrc    = errors.New("requires a source path to a directory")
	ErrAnsiLove = errors.New("cannot access shared libraries: libansilove.so.1")
	ErrDest     = errors.New("dest argument requires a destination filename path")
)

// generate a collection of site images.
func generate(name, id string) error {
	n, f := name, directories.Files(id)
	o := f.Img000 + png
	s, err := makePng(n, f.Img000)
	if err != nil && err.Error() == `execute ansilove: executable file not found in $PATH` {
		const note = `
this command requires the installation of AnsiLove/C
installation instructions: https://github.com/ansilove/ansilove
`
		fmt.Println(note)
		return fmt.Errorf("generate ansilove not found: %w", err)
	} else if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	fmt.Printf("  %s", s)
	// cap images to the webp limit of 16383 pixels
	const (
		webpMaxSize = 16383
		thumbSmall  = 150
		thumbMedium = 400
	)
	var w int
	if w, err = images.Width(o); w > webpMaxSize {
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}
		fmt.Printf("width %dpx", w)
		s, err = images.ToPng(o, images.NewExt(o, png), webpMaxSize)
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}
		fmt.Printf("  %s", s)
	}
	s, err = images.ToWebp(o, images.NewExt(o, webp), true)
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	fmt.Printf("  %s", s)
	s, err = images.ToThumb(o, f.Img400, thumbMedium)
	if err != nil {
		return fmt.Errorf("generate thumb %dpx: %w", thumbMedium, err)
	}
	fmt.Printf("  %s", s)
	s, err = images.ToThumb(o, f.Img150, thumbSmall)
	if err != nil {
		return fmt.Errorf("generate thumb %dpx: %w", thumbSmall, err)
	}
	fmt.Printf("  %s", s)
	return nil
}

// ToPng converts any supported format to a compressed PNG image.
// helpful: https://www.programming-books.io/essential/go/images-png-jpeg-bmp-tiff-webp-vp8-gif-c84a45304ec3498081c67aa1ea0d9c49
func makePng(src, dest string) (string, error) {
	if src == "" {
		return "", fmt.Errorf("make png: %w", ErrNoSrc)
	}
	_, err := os.Stat(src)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("make png missing %q: %w", src, err)
	} else if err != nil {
		return "", fmt.Errorf("make png stat: %w", err)
	}
	if dest == "" {
		return "", fmt.Errorf("make png: %w", ErrDest)
	}
	img := dest + png
	cmd := exec.Command("ansilove", "-r", "-o", img, src)
	out, err := cmd.Output()
	if err != nil && err.Error() == "exit status 127" {
		return "", fmt.Errorf("make png ansilove: %w", ErrAnsiLove)
	} else if err != nil {
		return "", fmt.Errorf("make png ansilove %q: %w", out, err)
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("make png working directory: %w", err)
	}
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("make png failed to obtain exe: %w", err)
	}
	fmt.Printf("%s - %s", wd, exe)
	stat, err := os.Stat(img)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("make png stat %q: %w", img, err)
	} else if err != nil {
		return "", fmt.Errorf("make png stat: %w", err)
	}
	return fmt.Sprintf("✓ text » png %v\n%s",
		humanize.Bytes(uint64(stat.Size())), color.Secondary.Sprintf("%s", out)), nil
}
