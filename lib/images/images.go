// Package images generate thumbnails and converts between image formats.
package images

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/gif"  // register GIF decoding
	_ "image/jpeg" // register Jpeg decoding
	"image/png"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	gap "github.com/muesli/go-app-paths"
	"github.com/nickalie/go-webpbin"
	"github.com/spf13/viper"
	"github.com/yusukebe/go-pngquant"
	_ "golang.org/x/image/bmp"  // register BMP decoding
	_ "golang.org/x/image/tiff" // register TIFF decoding
	_ "golang.org/x/image/webp" // register WebP decoding
)

const (
	WebpMaxSize = 16383

	fperm os.FileMode = 0666
	fmode             = os.O_RDWR | os.O_CREATE
)

var (
	ErrFormat = errors.New("unsupported image format")
	ErrViper  = errors.New("viper directory locations cannot be read")
)

// Duplicate an image file and appends suffix to its name.
func Duplicate(filename, suffix string) (name string, err error) {
	src, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("duplicate %q: %w", filename, err)
	}
	defer src.Close()
	ext := filepath.Ext(filename)
	fn := strings.TrimSuffix(filename, ext)
	dest, err := os.OpenFile(fn+suffix+ext, fmode, fperm)
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

func valid(name string, err error) bool {
	if err != nil {
		return false
	}
	s, err := os.Stat(name)
	if err != nil {
		return false
	}
	if s.IsDir() || s.Size() < 1 {
		return false
	}
	return true
}

// Generate a collection of site images.
func Generate(src, id string, remove bool) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("generate stat %q: %w", src, err)
	}
	out := func(s string, e error) {
		if s != "" {
			logs.Printf("  %s", s)
			return
		}
		logs.Log(e)
	}
	if viper.GetString("directory.root") == "" {
		return ErrViper
	}
	f := directories.Files(id)
	// these funcs use dependencies that are not thread safe
	// convert to png
	pngLoc, webpLoc := NewExt(f.Img000, _png), NewExt(f.Img000, webp)
	pngOk, webpOk := false, false
	s, err := ToPng(src, pngLoc, 1500, 1500)
	out(s, err)
	pngOk = valid(pngLoc, err)
	// convert to webp
	if s, err = ToWebp(src, webpLoc, true); !errors.Is(err, ErrFormat) {
		out(s, err)
	}
	if err != nil && pngOk {
		s, err = ToWebp(pngLoc, webpLoc, true)
		out(s, err)
	}
	webpOk = valid(webpLoc, err)
	// make 400x400 thumbs
	s, err = ToThumb(src, f.Img400, 400)
	if err != nil && pngOk {
		s, err = ToThumb(pngLoc, f.Img400, 400)
	} else if err != nil && webpOk {
		s, err = ToThumb(webpLoc, f.Img400, 400)
	}
	out(s, err)
	// make 150x150 thumbs
	s, err = ToThumb(src, f.Img150, 150)
	if err != nil && pngOk {
		s, err = ToThumb(pngLoc, f.Img150, 150)
	} else if err != nil && webpOk {
		s, err = ToThumb(webpLoc, f.Img150, 150)
	}
	out(s, err)
	return cleanup(remove, src)
}

func cleanup(remove bool, src string) error {
	if remove {
		if err := os.Remove(src); err != nil {
			return fmt.Errorf("generate cleanup %q: %w", src, err)
		}
	}
	return nil
}

// Move a file from the source location to the destination.
// This is used in situations where os.rename() fails due to multiple partitions.
func Move(src, dest string) error {
	s, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("move open: %w", err)
	}
	defer s.Close()
	d, err := os.OpenFile(dest, fmode, fperm)
	if err != nil {
		return fmt.Errorf("move create: %w", err)
	}
	defer d.Close()
	_, err = io.Copy(s, d)
	if err != nil {
		return fmt.Errorf("move io copy: %w", err)
	}
	if err = os.Remove(src); err != nil {
		return fmt.Errorf("move remove: %w", err)
	}
	return nil
}

// Info returns the image metadata.
func Info(name string) (width, height int, format string, err error) {
	file, err := os.Open(name)
	if err != nil {
		return 0, 0, "", fmt.Errorf("info open %q: %w", name, err)
	}
	defer file.Close()
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, "", fmt.Errorf("info decode config: %w", err)
	}
	return config.Width, config.Height, format, file.Close()
}

// NewExt replaces or appends the extension to a file name.
func NewExt(name, ext string) string {
	e := filepath.Ext(name)
	if e == "" {
		return name + ext
	}
	fn := strings.TrimSuffix(name, e)
	return fn + ext
}

// ToPng converts any supported format to a compressed PNG image.
// helpful: https://www.programming-books.io/essential/go/images-png-jpeg-bmp-tiff-webp-vp8-gif-c84a45304ec3498081c67aa1ea0d9c49
func ToPng(src, dest string, width, height int) (s string, err error) {
	in, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("to png open %s: %w", src, err)
	}
	defer in.Close()
	// image.Decode will determine the format
	img, ext, err := image.Decode(in)
	if err != nil {
		return "", fmt.Errorf("to png decode: %w", err)
	}
	// cap image size
	if width > 0 || height > 0 {
		img = imaging.Thumbnail(img, width, height, imaging.Lanczos)
	}
	// use the 3rd party CLI tool, pngquant to compress the PNG data
	img, err = pngquant.Compress(img, "4")
	if err != nil {
		return "", fmt.Errorf("to png quant compress: %w", err)
	}
	// adjust any configs to the PNG image encoder
	cfg := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	// write the PNG data to img
	buf := new(bytes.Buffer)
	if err = cfg.Encode(buf, img); err != nil {
		return "", fmt.Errorf("to png buffer encode: %w", err)
	}
	// save the PNG to a file
	out, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("to png create %q: %w", dest, err)
	}
	defer out.Close()
	if _, err := buf.WriteTo(out); err != nil {
		return "", fmt.Errorf("to png buffer write to: %w", err)
	}
	return fmt.Sprintf("%v»png", strings.ToLower(ext)), nil
}

// ToThumb creates a thumb from an image that is pixel squared in size.
func ToThumb(src, dest string, sizeSquared int) (s string, err error) {
	pfx := "_" + fmt.Sprintf("%v", sizeSquared) + "x"
	cp, err := Duplicate(src, pfx)
	if err != nil {
		return "", fmt.Errorf("to thumb duplicate: %w", err)
	}
	in, err := imaging.Open(cp)
	if err != nil {
		return "", fmt.Errorf("to thumb imaging open: %w", err)
	}
	in = imaging.Resize(in, sizeSquared, 0, imaging.Lanczos)
	in = imaging.CropAnchor(in, sizeSquared, sizeSquared, imaging.Center)
	// use the 3rd party CLI tool, pngquant to compress the PNG data
	in, err = pngquant.Compress(in, "4")
	if err != nil {
		return "", fmt.Errorf("to thumb png quant compress: %w", err)
	}
	f := NewExt(cp, _png)
	if err := imaging.Save(in, f); err != nil {
		return "", fmt.Errorf("to thumb imaging save: %w", err)
	}
	ext := NewExt(dest, _png)
	if err := os.Rename(f, ext); err != nil {
		var le *os.LinkError // invalid cross-device link
		if errors.As(err, &le) {
			if err = Move(f, ext); err != nil {
				return "", fmt.Errorf("to thumb move: %w", err)
			}
		} else {
			return "", fmt.Errorf("to thumb rename: %w", err)
		}
	}
	if _, err := os.Stat(cp); err == nil {
		if err := os.Remove(cp); err != nil {
			return "", fmt.Errorf("to thumb remove %q: %w", cp, err)
		}
	}
	return fmt.Sprintf("»%vx", sizeSquared), nil
}

// ToWebp converts any supported format to a WebP image using a 3rd party library.
// Input format can be either PNG, JPEG, TIFF, WebP or raw Y'CbCr samples.
func ToWebp(src, dest string, vendorTempDir bool) (s string, err error) {
	valid := func(a []string, x string) bool {
		for _, n := range a {
			if x == n {
				return true
			}
		}
		return false
	}
	v := []string{_png, jpg, jpeg, tif, tiff, webp}
	// skip if already a webp image, or handle all other errors
	m, err := mimetype.DetectFile(src)
	if m.Extension() == webp {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("to webp mimetype detect: %w", err)
	}
	if !valid(v, m.Extension()) {
		return "", fmt.Errorf("to webp extension %q != %s: %w",
			m.Extension(), strings.Join(v, " "), ErrFormat)
	}
	src, err = cropWebP(src)
	if err != nil {
		return "", fmt.Errorf("to webp crop: %w", err)
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
		if err = cleanupWebP(dest); err != nil {
			return "", fmt.Errorf("to webp cleanup: %w", err)
		}
		return "", fmt.Errorf("to webp run: %w", err)
	}
	return "»webp", nil
}

func cleanupWebP(name string) error {
	s, err := os.Stat(name)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("cleanWebP stat: %w", err)
	}
	if s.Size() == 0 {
		if err := os.Remove(name); err != nil {
			return fmt.Errorf("cleanWebP remove: %w", err)
		}
	}
	return nil
}

// cropWebP crops an image to be usable size for WebP conversion.
func cropWebP(src string) (string, error) {
	w, h, _, err := Info(src)
	if err != nil {
		return "", fmt.Errorf("to webp size: %w", err)
	}
	if w+h > WebpMaxSize {
		cropW, cropH := WebPCalc(w, h)
		ext := filepath.Ext(src)
		fn := strings.TrimSuffix(src, ext)
		crop := fn + "-cropped" + ext
		_, err := ToPng(src, crop, cropW, cropH)
		if err != nil {
			return "", fmt.Errorf("webp crop: %w", err)
		}
		defer os.Remove(crop)
		return crop, nil
	}
	return src, nil
}

// WebPCalc calculates the largest permitted sizes for a valid WebP crop.
func WebPCalc(width, height int) (w, h int) {
	if width+height <= WebpMaxSize {
		return width, height
	}
	if width == height {
		r := WebpMaxSize / 2
		return r, r
	}
	big, small := height, width
	if width > height {
		big, small = width, height
	}
	r := big - small + (WebpMaxSize - big)
	if width > height {
		return small, r
	}
	return r, small
}

// Width returns the image width in pixels.
func Width(name string) (width int, err error) {
	width, _, _, err = Info(name)
	if err != nil {
		return 0, fmt.Errorf("width: %w", err)
	}
	return width, nil
}

// vendorPath is the absolute path to store webpbin vendor downloads.
func vendorPath() string {
	fp, err := gap.NewScope(gap.User, logs.GapUser).CacheDir()
	if err != nil {
		h, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(fmt.Errorf("vendorPath userhomedir: %w", err))
		}
		return path.Join(h, ".vendor/df2")
	}
	return fp
}
