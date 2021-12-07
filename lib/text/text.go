// Package text generates images from text files using the Ansilove/C program.
package text

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/images"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
)

var (
	ErrNoSrc    = errors.New("requires a source path to a directory")
	ErrAnsiLove = errors.New("cannot access shared libraries: libansilove.so.1")
	ErrDest     = errors.New("dest argument requires a destination filename path")
	ErrZero     = errors.New("source file is 0 bytes")
	ErrMeNo     = errors.New("no readme chosen")
	ErrMeUnk    = errors.New("unknown readme")
	ErrMeNF     = errors.New("readme not found in archive")
)

// reduce the length of the textfile so it can be parsed by AnsiLove.
func reduce(src, uuid string) (string, error) {
	fmt.Print(" will attempt to reduce the length of file")

	f, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer f.Close()

	save, err := os.CreateTemp(os.TempDir(), uuid+"_reduce")
	if err != nil {
		return "", err
	}
	defer save.Close()

	const maxLines = 500
	scanner, writer := bufio.NewScanner(f), bufio.NewWriter(save)
	scanner.Split(bufio.ScanLines)
	line := 0
	for scanner.Scan() {
		line++
		if _, err := writer.WriteString(scanner.Text() + "\n"); err != nil {
			os.Remove(save.Name())
			return "", err
		}
		if line > maxLines {
			break
		}
	}
	return save.Name(), nil
}

// generate a collection of site images.
func generate(name, uuid string, amiga bool) error {
	prnt := func(s string) {
		fmt.Printf("  %s", s)
	}
	const note = `
this command requires the installation of AnsiLove/C
installation instructions: https://github.com/ansilove/ansilove
`
	n, f := name, directories.Files(uuid)
	o := f.Img000 + png
	s, err := makePng(n, f.Img000, amiga)
	if err != nil && err.Error() == `execute ansilove: executable file not found in $PATH` {
		fmt.Println(note)
		return fmt.Errorf("generate ansilove not found: %w", err)
	} else if err != nil && errors.Unwrap(err).Error() == "signal: killed" {
		tmp, err1 := reduce(f.UUID, uuid)
		if err1 != nil {
			return fmt.Errorf("ansilove reduce: %w", err1)
		}
		s, err = makePng(tmp, f.Img000, amiga)
		defer os.Remove(tmp)
	}
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	prnt(s)
	const thumbMedium = 400
	var w, h int
	if w, h, _, err = images.Info(o); (w + h) > images.WebpMaxSize {
		if err != nil {
			return fmt.Errorf("generate info: %w", err)
		}
		cw, ch := images.WebPCalc(w, h)
		s, err = images.ToPng(o, images.NewExt(o, png), ch, cw)
		if err != nil {
			return fmt.Errorf("generate calc: %w", err)
		}
		prnt(s)
	}
	s, err = images.ToWebp(o, images.NewExt(o, webp), true)
	if err != nil {
		return fmt.Errorf("generate webp: %w", err)
	}
	prnt(s)
	s, err = images.ToThumb(o, f.Img400, thumbMedium)
	if err != nil {
		return fmt.Errorf("generate thumb %dpx: %w", thumbMedium, err)
	}
	prnt(s)
	return nil
}

// ToPng converts any supported format to a compressed PNG image.
// helpful: https://www.programming-books.io/essential/go/images-png-jpeg-bmp-tiff-webp-vp8-gif-c84a45304ec3498081c67aa1ea0d9c49
func makePng(src, dest string, amiga bool) (string, error) {
	if src == "" {
		return "", fmt.Errorf("make png: %w", ErrNoSrc)
	}
	if dest == "" {
		return "", fmt.Errorf("make png: %w", ErrDest)
	}
	_, err := os.Stat(src)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("make png missing %q: %w", src, err)
	} else if err != nil {
		return "", fmt.Errorf("make png stat: %w", err)
	}
	const ten = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), ten)
	defer cancel()
	// ansilove -q # suppress output messages
	// ansilove -r # create Retina @2x output file
	// ansilove -o # specify output filename/path
	// ansilove -f # select font for supported formats: 80x25 (default), topaz+, 80x50, ...
	img := dest + png
	args := []string{"-q", "-r", "-o", img}
	if amiga {
		args = append(args, "-f", "topaz+")
	}
	args = append(args, src)
	cmd := exec.CommandContext(ctx, "ansilove", args...)
	out, err := cmd.Output()
	if err != nil && err.Error() == "exit status 127" {
		return "", fmt.Errorf("make ansilove: %w", ErrAnsiLove)
	} else if err != nil {
		return "", fmt.Errorf("make ansilove %q: %w", out, err)
	}

	_, err = os.Getwd()
	if err != nil {
		return "", fmt.Errorf("make png working directory: %w", err)
	}

	_, err = os.Executable()
	if err != nil {
		return "", fmt.Errorf("make png failed to obtain exe: %w", err)
	}

	stat, err := os.Stat(img)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("make png stat %q: %w", img, err)
	} else if err != nil {
		return "", fmt.Errorf("make png stat: %w", err)
	}

	return fmt.Sprintf("✓ text » png %v\n%s",
		humanize.Bytes(uint64(stat.Size())), color.Secondary.Sprintf("%s", out)), nil
}
