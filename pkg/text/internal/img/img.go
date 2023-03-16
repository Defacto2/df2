package img

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/images"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
)

var (
	ErrNoSrc    = errors.New("requires a source path to a directory")
	ErrAnsiLove = errors.New("cannot access shared libraries: libansilove.so.1")
	ErrDest     = errors.New("dest argument requires a destination filename path")
)

const (
	png  = ".png"
	webp = ".webp"
)

// Generate a collection of site images.
func Generate(w io.Writer, cfg configger.Config, name, uuid string, amiga bool) error {
	stdout := func(s string) {
		fmt.Fprintf(w, "  %s", s)
	}
	const note = `
this command requires the installation of AnsiLove/C
installation instructions: https://github.com/ansilove/ansilove`
	f, err := directories.Files(cfg, uuid)
	if err != nil {
		return err
	}
	o := f.Img000 + png
	s, err := MakePng(name, f.Img000, amiga)
	if err != nil && err.Error() == `execute ansilove: executable file not found in $PATH` {
		log.Println(note)
		return fmt.Errorf("generate, ansilove not found: %w", err)
	}
	if err != nil && errors.Unwrap(err).Error() == "signal: killed" {
		tmp, err1 := Reduce(w, f.UUID, uuid)
		if err1 != nil {
			return fmt.Errorf("ansilove reduce: %w", err1)
		}
		s, err = MakePng(tmp, f.Img000, amiga)
		defer os.Remove(tmp)
	}
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	stdout(s)
	const thumbMedium = 400
	if err := resize(w, o); err != nil {
		return err
	}
	s, err = images.ToWebp(w, o, images.NewExt(o, webp), true)
	if err != nil {
		return fmt.Errorf("generate webp: %w", err)
	}
	stdout(s)
	s, err = images.ToThumb(o, f.Img400, thumbMedium)
	if err != nil {
		return fmt.Errorf("generate thumb %dpx: %w", thumbMedium, err)
	}
	stdout(s)
	return nil
}

func resize(w io.Writer, o string) error {
	var wp, hp int
	var err error
	if wp, hp, _, err = images.Info(o); (wp + hp) > images.WebpMaxSize {
		if err != nil {
			return fmt.Errorf("generate info: %w", err)
		}
		cw, ch := images.WebPCalc(wp, hp)
		s, err := images.ToPng(o, images.NewExt(o, png), ch, cw)
		if err != nil {
			return fmt.Errorf("generate calc: %w", err)
		}
		fmt.Fprintf(w, "  %s", s)
	}
	return nil
}

// Reduce the length of the textfile so it can be parsed by AnsiLove.
func Reduce(w io.Writer, src, uuid string) (string, error) {
	fmt.Fprint(w, " will attempt to reduce the length of file")

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

// ToPng converts any supported format to a compressed PNG image.
// helpful: https://www.programming-books.io/essential/go
// /images-png-jpeg-bmp-tiff-webp-vp8-gif-c84a45304ec3498081c67aa1ea0d9c49.
func MakePng(src, dest string, amiga bool) (string, error) {
	if src == "" {
		return "", fmt.Errorf("makepng: %w", ErrNoSrc)
	}
	if dest == "" {
		return "", fmt.Errorf("makepng: %w", ErrDest)
	}
	_, err := os.Stat(src)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("makepng missing %q: %w", src, err)
	} else if err != nil {
		return "", fmt.Errorf("makepng stat: %w", err)
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
		return "", fmt.Errorf("make ansilove %q: %w", args, ErrAnsiLove)
	} else if err != nil {
		return "", fmt.Errorf("make ansilove %q: %q %w", args, out, err)
	}
	size, err := checkSize(img)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("✓ text » png %v\n%s",
		humanize.Bytes(size), color.Secondary.Sprintf("%s", out)), nil
}

func checkSize(img string) (uint64, error) {
	_, err := os.Getwd()
	if err != nil {
		return 0, fmt.Errorf("makepng working directory: %w", err)
	}

	_, err = os.Executable()
	if err != nil {
		return 0, fmt.Errorf("makepng failed to obtain exe: %w", err)
	}

	stat, err := os.Stat(img)
	if err != nil && os.IsNotExist(err) {
		return 0, fmt.Errorf("makepng stat %q: %w", img, err)
	} else if err != nil {
		return 0, fmt.Errorf("makepng stat: %w", err)
	}
	return uint64(stat.Size()), nil
}
