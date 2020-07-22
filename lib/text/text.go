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

const ansiloveErr = `
this command requires the installation of AnsiLove/C
installation instructions: https://github.com/ansilove/ansilove
`

// generate a collection of site images.
func generate(name, id string) error {
	n, f := name, directories.Files(id)
	o := f.Img000 + png
	s, err := makePng(n, f.Img000)
	if err != nil && err.Error() == `execute ansilove: executable file not found in $PATH` {
		fmt.Println(ansiloveErr)
		return err
	} else if err != nil {
		return err
	}
	fmt.Printf("  %s", s)
	// cap images to the webp limit of 16383 pixels
	const limit = 16383
	if w, err := images.Width(o); w > limit {
		if err != nil {
			return err
		}
		fmt.Printf("width %dpx", w)
		s, err = images.ToPng(o, images.NewExt(o, png), limit)
		if err != nil {
			return err
		}
		fmt.Printf("  %s", s)
	}
	s, err = images.ToWebp(o, images.NewExt(o, webp), true)
	if err != nil {
		return err
	}
	fmt.Printf("  %s", s)
	s, err = images.ToThumb(o, f.Img400, 400)
	if err != nil {
		return err
	}
	fmt.Printf("  %s", s)
	s, err = images.ToThumb(o, f.Img150, 150)
	if err != nil {
		return err
	}
	fmt.Printf("  %s", s)
	return nil
}

// ToPng converts any supported format to a compressed PNG image.
// helpful: https://www.programming-books.io/essential/go/images-png-jpeg-bmp-tiff-webp-vp8-gif-c84a45304ec3498081c67aa1ea0d9c49
func makePng(src, dest string) (string, error) {
	if src == "" {
		return "", errors.New("makepng: src argument requires a source directory path")
	}
	_, err := os.Stat(src)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("src file does not exist: %q: %s", src, err)
	} else if err != nil {
		return "", err
	}
	if dest == "" {
		return "", errors.New("makepng: dest argument requires a destination filename path")
	}
	img := dest + png
	cmd := exec.Command("ansilove", "-r", "-o", img, src)
	out, err := cmd.Output()
	if err != nil && err.Error() == "exit status 127" {
		return "", errors.New("ansilove: cannot access shared libraries: libansilove.so.1")
	} else if err != nil {
		return "", fmt.Errorf("ansilove: %s", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("no working directory: %s", err)
	}
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to obtain exe: %s", err)
	}
	fmt.Printf("%s - %s", wd, exe)
	stat, err := os.Stat(img)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("makepng: new image is not found: %s", err)
	} else if err != nil {
		return "", err
	}
	ss := stat.Size()
	return fmt.Sprintf("✓ text » png %v\n%s",
		humanize.Bytes(uint64(ss)), color.Secondary.Sprintf("%s", out)), nil
}
