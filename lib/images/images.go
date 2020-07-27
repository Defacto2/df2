package images

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"  // register GIF decoding
	_ "image/jpeg" // register Jpeg decoding
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/disintegration/imaging"
	"github.com/dustin/go-humanize"
	"github.com/gabriel-vasile/mimetype"
	gap "github.com/muesli/go-app-paths"
	"github.com/nickalie/go-webpbin"
	"github.com/yusukebe/go-pngquant"
	_ "golang.org/x/image/bmp"  // register BMP decoding
	_ "golang.org/x/image/tiff" // register TIFF decoding
	_ "golang.org/x/image/webp" // register WebP decoding
)

const (
	fp os.FileMode = 0666
	fm             = os.O_RDWR | os.O_CREATE
)

var scope = gap.NewScope(gap.User, "df2")

// Duplicate an image file and appends suffix to its name.
func Duplicate(filename, suffix string) (name string, err error) {
	src, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("duplicate %q: %w", filename, err)
	}
	defer src.Close()
	ext := filepath.Ext(filename)
	fn := strings.TrimSuffix(filename, ext)
	dest, err := os.OpenFile(fn+suffix+ext, fm, fp)
	if err != nil {
		return "", fmt.Errorf("duplicate open file %q: %w", fn, err)
	}
	defer dest.Close()
	if _, err = io.Copy(dest, src); err != nil {
		return "", fmt.Errorf("duplicate io copy: %w", err)
	}
	name = fmt.Sprintf("%s%s%s", fn, suffix, ext)
	return name, nil
}

// Generate a collection of site images.
func Generate(name, id string, remove bool) error {
	n := name
	out := func(s string, e error) {
		if s != "" {
			logs.Printf("  %s", s)
		} else {
			logs.Log(e)
		}
	}
	f := directories.Files(id)
	// these funcs use dependencies that are not thread safe
	s, err := ToPng(n, NewExt(f.Img000, _png), 1500)
	out(s, err)
	s, err = ToWebp(n, NewExt(f.Img000, webp), true)
	out(s, err)
	s, err = ToThumb(n, f.Img400, 400)
	out(s, err)
	s, err = ToThumb(n, f.Img150, 150)
	out(s, err)
	if remove {
		if err := os.Remove(n); err != nil {
			return fmt.Errorf("generate remove %q: %w", n, err)
		}
	}
	return nil
}

// NewExt replaces or appends the extension to a file name.
func NewExt(path, ext string) (name string) {
	e := filepath.Ext(path)
	if e == "" {
		return path + ext
	}
	fn := strings.TrimSuffix(path, e)
	return fn + ext
}

// Info returns the image metadata.
func Info(name string) (height int, width int, format string, err error) {
	file, err := os.Open(name)
	if err != nil {
		return 0, 0, "", fmt.Errorf("info open %q: %w", name, err)
	}
	defer file.Close()
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, "", fmt.Errorf("info decode config: %w", err)
	}
	return config.Height, config.Width, format, file.Close()
}

// Width returns the image width in pixels.
func Width(name string) (width int, err error) {
	_, width, _, err = Info(name)
	if err != nil {
		return 0, fmt.Errorf("width: %w", err)
	}
	return width, nil
}

// ToPng converts any supported format to a compressed PNG image.
// helpful: https://www.programming-books.io/essential/go/images-png-jpeg-bmp-tiff-webp-vp8-gif-c84a45304ec3498081c67aa1ea0d9c49
func ToPng(src, dest string, maxDimension int) (print string, err error) {
	in, err := os.Open(src)
	if err != nil {
		return print, fmt.Errorf("to png open %s: %w", src, err)
	}
	defer in.Close()
	// image.Decode will determine the format
	img, ext, err := image.Decode(in)
	if err != nil {
		return print, fmt.Errorf("to png decode: %w", err)
	}
	// cap image size
	if maxDimension > 0 {
		img = imaging.Thumbnail(img, maxDimension, maxDimension, imaging.Lanczos)
	}
	// use the 3rd party CLI tool, pngquant to compress the PNG data
	img, err = pngquant.Compress(img, "4")
	if err != nil {
		return print, fmt.Errorf("to png quant compress: %w", err)
	}
	// adjust any configs to the PNG image encoder
	cfg := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	// write the PNG data to img
	buf := new(bytes.Buffer)
	if err = cfg.Encode(buf, img); err != nil {
		return print, fmt.Errorf("to png buffer encode: %w", err)
	}
	// save the PNG to a file
	out, err := os.Create(dest)
	if err != nil {
		return print, fmt.Errorf("to png create %q: %w", dest, err)
	}
	defer out.Close()
	if _, err := buf.WriteTo(out); err != nil {
		return print, fmt.Errorf("to png buffer write to: %w", err)
	}
	return fmt.Sprintf("%v»png", strings.ToLower(ext)), nil
}

// ToThumb creates a thumb from an image that is pixel squared in size.
func ToThumb(src, dest string, sizeSquared int) (print string, err error) {
	pfx := "_" + fmt.Sprintf("%v", sizeSquared) + "x"
	cp, err := Duplicate(src, pfx)
	if err != nil {
		return print, fmt.Errorf("to thumb duplicate: %w", err)
	}
	in, err := imaging.Open(cp)
	if err != nil {
		return print, fmt.Errorf("to thumb imaging open: %w", err)
	}
	in = imaging.Resize(in, sizeSquared, 0, imaging.Lanczos)
	in = imaging.CropAnchor(in, sizeSquared, sizeSquared, imaging.Center)
	// use the 3rd party CLI tool, pngquant to compress the PNG data
	in, err = pngquant.Compress(in, "4")
	if err != nil {
		return print, fmt.Errorf("to thumb png quant compress: %w", err)
	}
	f := NewExt(cp, _png)
	if err := imaging.Save(in, f); err != nil {
		return print, fmt.Errorf("to thumb imaging save: %w", err)
	}
	s := NewExt(dest, _png)
	if err := os.Rename(f, s); err != nil {
		return print, fmt.Errorf("to thumb rename: %w", err)
	}
	if _, err := os.Stat(cp); err == nil {
		if err := os.Remove(cp); err != nil {
			return print, fmt.Errorf("to thumb remove %q: %w", cp, err)
		}
	}
	return fmt.Sprintf("»%vx", sizeSquared), nil
}

// ToWebp converts any supported format to a WebP image using a 3rd party library.
// Input format can be either PNG, JPEG, TIFF, WebP or raw Y'CbCr samples.
func ToWebp(src, dest string, vendorTempDir bool) (print string, err error) {
	// skip if already a webp image, or handle all other errors
	if m, err := mimetype.DetectFile(src); m.Extension() == webp {
		return print, nil
	} else if err != nil {
		return print, fmt.Errorf("to webp mimetype detect: %w", err)
	}
	const percent = 70
	webp := webpbin.NewCWebP().
		Quality(percent).
		InputFile(src).
		OutputFile(dest)
	if vendorTempDir {
		webp.Dest(vendorPath())
	}
	if err = webp.Run(); err != nil {
		return print, fmt.Errorf("to webp run: %w", err)
	}
	return "»webp", nil
}

// vendorPath is the absolute path to store webpbin vendor downloads.
func vendorPath() string {
	fp, err := scope.CacheDir()
	if err != nil {
		h, _ := os.UserHomeDir()
		return path.Join(h, ".vendor/df2")
	}
	return fp
}

func filesize(name string) string {
	f, err := os.Stat(name)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%v", humanize.Bytes(uint64(f.Size())))
}
