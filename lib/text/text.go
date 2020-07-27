package text

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/images"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
)

var (
	ErrNoSrc    = errors.New("requires a source path to a directory")
	ErrAnsiLove = errors.New("cannot access shared libraries: libansilove.so.1")
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
		return fmt.Errorf("generate make png: %w", err)
	}
	fmt.Printf("  %s", s)
	// cap images to the webp limit of 16383 pixels
	const limit = 16383
	if w, err := images.Width(o); w > limit {
		if err != nil {
			return fmt.Errorf("generate images width: %w", err)
		}
		fmt.Printf("width %dpx", w)
		s, err = images.ToPng(o, images.NewExt(o, png), limit)
		if err != nil {
			return fmt.Errorf("generate to png: %w", err)
		}
		fmt.Printf("  %s", s)
	}
	s, err = images.ToWebp(o, images.NewExt(o, webp), true)
	if err != nil {
		return fmt.Errorf("generate to webp: %w", err)
	}
	fmt.Printf("  %s", s)
	s, err = images.ToThumb(o, f.Img400, 400)
	if err != nil {
		return fmt.Errorf("generate to thumb 400px: %w", err)
	}
	fmt.Printf("  %s", s)
	s, err = images.ToThumb(o, f.Img150, 150)
	if err != nil {
		return fmt.Errorf("generate to thumb 150px: %w", err)
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
		return "", errors.New("makepng: dest argument requires a destination filename path")
	}
	img := dest + png
	cmd := exec.Command("ansilove", "-r", "-o", img, src)
	out, err := cmd.Output()
	if err != nil && err.Error() == "exit status 127" {
		return "", fmt.Errorf("makepng ansilove: %w", ErrAnsiLove)
	} else if err != nil {
		return "", fmt.Errorf("makepng ansilove: %w", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("makepng working directory: %w", err)
	}
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("makepng failed to obtain exe: %w", err)
	}
	fmt.Printf("%s - %s", wd, exe)
	stat, err := os.Stat(img)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("makepng stat %q: %w", img, err)
	} else if err != nil {
		return "", fmt.Errorf("makepng stat: %w", err)
	}
	return fmt.Sprintf("✓ text » png %v\n%s",
		humanize.Bytes(uint64(stat.Size())), color.Secondary.Sprintf("%s", out)), nil
}
