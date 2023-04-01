package img

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/images"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
)

var (
	ErrNamed     = errors.New("named path cannot be empty")
	ErrDest      = errors.New("dest path cannot be empty")
	ErrSharedLib = errors.New("cannot access shared libraries: libansilove.so.1")
	ErrType      = errors.New("mimetype is not a known text file")

	ErrResize = errors.New("resize only works with the png image format")
)

const (
	png  = ".png"
	webp = ".webp"
)

// Make both PNG and Webp preview images and a 400x PNG thumbnail from the named text files.
// Name is the source text file required for conversion to an image.
// UUID is the universal ID used for the image filename.
// When the amiga bool is true the image text will use an Amiga era Topaz+, 80x50 font.
func Make(w io.Writer, cfg conf.Config, name, uuid string, amiga bool) error {
	if w == nil {
		w = io.Discard
	}
	if err := Type(name); err != nil {
		return err
	}
	const note = `
this command requires the installation of AnsiLove/C
installation instructions: https://github.com/ansilove/ansilove`
	f, err := directories.Files(cfg, uuid)
	if err != nil {
		return err
	}
	s, err := Export(name, f.Img000, amiga)
	if err != nil {
		if err.Error() == `execute ansilove: executable file not found in $PATH` {
			fmt.Fprintln(w, note)
			return fmt.Errorf("generate, ansilove not found: %w", err)
		}
		if errors.Unwrap(err).Error() == "signal: killed" {
			tmp, err1 := Reduce(w, f.UUID, uuid)
			if err1 != nil {
				return fmt.Errorf("ansilove reduce: %w", err1)
			}
			s, err = Export(tmp, f.Img000, amiga)
			if err != nil {
				return fmt.Errorf("generate: %w", err)
			}
			defer os.Remove(tmp)
			err = nil
		}
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}
	}
	fmt.Fprintf(w, "  %s", s)
	const thumbMedium = 400
	src := f.Img000 + png
	if err := Resize(w, src); err != nil {
		return err
	}
	s, err = images.ToWebp(w, src, images.ReplaceExt(webp, src), true)
	if err != nil {
		return fmt.Errorf("generate webp: %w", err)
	}
	fmt.Fprintf(w, "  %s", s)
	s, err = images.ToThumb(src, f.Img400, thumbMedium)
	if err != nil {
		return fmt.Errorf("generate thumb %dpx: %w", thumbMedium, err)
	}
	fmt.Fprintf(w, "  %s", s)
	return nil
}

// Type checks the the document type of the named file and returns an error
// if it is a known binary format.
func Type(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("type cannot open the named file: %w: %s", err, name)
	}
	defer f.Close()
	dst := &bytes.Buffer{}
	inf, err := f.Stat()
	if err != nil {
		return fmt.Errorf("type cannot stat the named file: %w: %s", err, name)
	}
	// DetectContentType only needs the first 512B of a file
	const maxSize = int64(512)
	n := maxSize
	if inf.Size() < maxSize {
		n = inf.Size()
	}
	if _, err = io.CopyN(dst, f, n); err != nil {
		return fmt.Errorf("type cannot copy the first %dB of the named file: %w: %s", n, err, name)
	}
	mime := http.DetectContentType(dst.Bytes())
	s := strings.Split(mime, "/")
	if len(s) == 0 {
		return err // todo
	}
	const fallback = "application/octet-stream"
	if mime == fallback {
		// fallback is often returned by ANSI encoded files etc
		return nil
	}
	if t := s[0]; t == "text" {
		return nil
	}
	return fmt.Errorf("%w %s: %s", ErrType, mime, name)
}

// Resize the named image so it is compatible with the format restrictions of WebP.
func Resize(w io.Writer, name string) error {
	if w == nil {
		w = io.Discard
	}
	wpx, hpx, s, err := images.Info(name)
	if s != "png" {
		return fmt.Errorf("%s, %w: %s", s, ErrResize, name)
	}
	if (wpx + hpx) > images.WebpMaxSize {
		if err != nil {
			return fmt.Errorf("generate info: %w", err)
		}
		width, height := images.WebPCalc(wpx, hpx)
		s, err := images.ToPNG(name, name, width, height)
		if err != nil {
			return fmt.Errorf("generate calc: %w", err)
		}
		fmt.Fprintf(w, "  %s", s)
	}
	return nil
}

// Reduce the length of the named textfile so it can be parsed by AnsiLove.
func Reduce(w io.Writer, named, uuid string) (string, error) {
	if w == nil {
		w = io.Discard
	}
	fmt.Fprint(w, " will attempt to reduce the length of file")

	f, err := os.Open(named)
	if err != nil {
		return "", err
	}
	defer f.Close()

	dest, err := os.CreateTemp(os.TempDir(), uuid+"_reduce")
	if err != nil {
		return "", err
	}
	defer dest.Close()

	scan, nw := bufio.NewScanner(f), bufio.NewWriter(dest)
	scan.Split(bufio.ScanLines)

	const maxLines = 500
	line := 0
	for scan.Scan() {
		line++
		if line > maxLines {
			break
		}
		if _, err := nw.WriteString(scan.Text() + "\n"); err != nil {
			os.Remove(dest.Name())
			return "", err
		}
	}
	nw.Flush()
	return dest.Name(), nil
}

// Export any supported text based named file to a compressed PNG image.
// helpful: https://www.programming-books.io/essential/go
// /images-png-jpeg-bmp-tiff-webp-vp8-gif-c84a45304ec3498081c67aa1ea0d9c49.
func Export(name, dest string, amiga bool) (string, error) {
	if name == "" {
		return "", fmt.Errorf("makepng: %w", ErrNamed)
	}
	if dest == "" {
		return "", fmt.Errorf("makepng: %w", ErrDest)
	}
	if _, err := os.Stat(name); errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("makepng missing %q: %w", name, err)
	} else if err != nil {
		return "", fmt.Errorf("makepng stat: %w", err)
	}

	const ten = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), ten)
	defer cancel()
	const (
		suppressOutput = "-q" // ansilove -q # suppress output messages
		retina2xImage  = "-r" // ansilove -r # create Retina @2x output file
		destination    = "-o" // ansilove -o # specify output filename/path
		selectFont     = "-f" // ansilove -f # select font for supported formats: 80x25 (default), topaz+, 80x50, ...
	)
	saveAs := dest + png
	args := []string{suppressOutput, retina2xImage, destination, saveAs}
	if amiga {
		args = append(args, selectFont, "topaz+")
	}
	args = append(args, name)
	cmd := exec.CommandContext(ctx, "ansilove", args...)
	out, err := cmd.Output()
	if err != nil && err.Error() == "exit status 127" {
		return "", fmt.Errorf("make ansilove %q: %w", args, ErrSharedLib)
	} else if err != nil {
		return "", fmt.Errorf("make ansilove %q: %q %w", args, out, err)
	}
	i, err := Bytes(saveAs)
	if err != nil {
		return "", err
	}
	fmt.Println(saveAs)
	return fmt.Sprintf("✓ text » png %v\n%s",
		humanize.Bytes(i), color.Secondary.Sprintf("%s", out)), nil
}

func Bytes(name string) (uint64, error) {
	stat, err := os.Stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, fmt.Errorf("bytes stat %q: %w", name, err)
	} else if err != nil {
		return 0, fmt.Errorf("bytes stat: %w", err)
	}
	return uint64(stat.Size()), nil
}
