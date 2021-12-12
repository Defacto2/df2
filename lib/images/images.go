// Package images generate thumbnails and converts between image formats.
package images

import (
	"bytes"
	"errors"
	"fmt"
	"image"

	// register GIF decoding.
	_ "image/gif"

	// register Jpeg decoding.
	_ "image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/images/internal/file"
	"github.com/Defacto2/df2/lib/images/internal/imagemagick"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gookit/color"
	"github.com/nickalie/go-webpbin"
	"github.com/spf13/viper"
	"github.com/yusukebe/go-pngquant"

	// register BMP decoding.
	_ "golang.org/x/image/bmp"

	// register TIFF decoding.
	_ "golang.org/x/image/tiff"

	// register WebP decoding.
	_ "golang.org/x/image/webp"
)

var (
	ErrFormat = errors.New("unsupported image format")
	ErrViper  = errors.New("viper directory locations cannot be read")
)

const (
	WebpMaxSize int         = 16383 // WebpMaxSize is the maximum pixel dimension of an webp image.
	thumbWidth  int         = 400
	fperm       os.FileMode = 0o666
	fmode                   = os.O_RDWR | os.O_CREATE

	gif  = ".gif"
	jpg  = ".jpg"
	jpeg = ".jpeg"
	_png = ".png"
	tif  = ".tif"
	tiff = ".tiff"
	webp = ".webp"
)

// Fix generates any missing assets from downloads that are images.
func Fix(simulate bool) error {
	dir := directories.Init(false)
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(`SELECT id, uuid, filename, filesize FROM files WHERE platform="image" ORDER BY id ASC`)
	if err != nil {
		return fmt.Errorf("images fix query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("images fix rows: %w", rows.Err())
	}
	defer rows.Close()
	c := 0
	for rows.Next() {
		var img file.Image
		if err = rows.Scan(&img.ID, &img.UUID, &img.Name, &img.Size); err != nil {
			return fmt.Errorf("images fix rows scan: %w", err)
		}
		if directories.ArchiveExt(img.Name) {
			continue
		}
		if !img.IsDir(&dir) {
			c++
			logs.Printf("%d. %v", c, img)
			if _, err := os.Stat(filepath.Join(dir.UUID, img.UUID)); os.IsNotExist(err) {
				logs.Printf("%s\n", str.X())
				continue
			} else if err != nil {
				return fmt.Errorf("images fix stat: %w", err)
			}
			if simulate {
				logs.Printf("%s\n", color.Question.Sprint("?"))
				continue
			}
			if err := Generate(filepath.Join(dir.UUID, img.UUID), img.UUID, false); err != nil {
				return fmt.Errorf("images fix generate: %w", err)
			}
			logs.Print("\n")
		}
	}
	if simulate && c > 0 {
		logs.Simulate()
	} else if c == 0 {
		logs.Println("everything is okay, there is nothing to do")
	}
	return nil
}

// Duplicate an image file and appends the suffix to its name.
func Duplicate(name, suffix string) (string, error) {
	src, err := os.Open(name)
	if err != nil {
		return "", fmt.Errorf("duplicate %q: %w", name, err)
	}
	defer src.Close()
	ext := filepath.Ext(name)
	fn := strings.TrimSuffix(name, ext)
	dest, err := os.OpenFile(fn+suffix+ext, fmode, fperm)
	if err != nil {
		return "", fmt.Errorf("duplicate open file %q: %w", fn, err)
	}
	defer dest.Close()
	if _, err = io.Copy(dest, src); err != nil {
		return "", fmt.Errorf("duplicate io copy: %w", err)
	}
	return fmt.Sprintf("%s%s%s", fn, suffix, ext), nil
}

// Generate a collection of site images.
func Generate(name, id string, remove bool) error {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return fmt.Errorf("generate stat %q: %w", name, err)
	}
	out := func(s string, e error) {
		if s != "" {
			logs.Printf("  %s", s)
			return
		}
		logs.Log(e)
	}
	if viper.GetString("directory.root") == "" {
		return fmt.Errorf("%w: directory.root", ErrViper)
	}
	f := directories.Files(id)
	// these funcs use dependencies that are not thread safe
	// convert to png
	pngDest, webpDest := NewExt(f.Img000, _png), NewExt(f.Img000, webp)
	const width = 1500
	s, err := ToPng(name, pngDest, width, width)
	out(s, err)
	// use imagemagick to convert unsupported image formats into PNG
	if !file.Check(pngDest, err) {
		err1 := imagemagick.Convert(name, pngDest)
		if err1 != nil {
			out(s, err1)
			return file.Remove(remove, name)
		}
		if !file.Check(pngDest, err1) {
			return file.Remove(remove, name)
		}
		name = pngDest
	}
	// convert to webp
	if s, err = ToWebp(name, webpDest, true); !errors.Is(err, ErrFormat) {
		out(s, err)
	}
	if err != nil {
		s, err = ToWebp(pngDest, webpDest, true)
		out(s, err)
	}
	webpOk := file.Check(webpDest, err)
	// make 400x400 thumbs
	s, err = ToThumb(name, f.Img400, thumbWidth)
	if err != nil {
		s, err = ToThumb(pngDest, f.Img400, thumbWidth)
	} else if err != nil && webpOk {
		s, err = ToThumb(webpDest, f.Img400, thumbWidth)
	}
	out(s, err)
	return file.Remove(remove, name)
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
// helpful: https://www.programming-books.io/essential/
// go/images-png-jpeg-bmp-tiff-webp-vp8-gif-c84a45304ec3498081c67aa1ea0d9c49.
func ToPng(src, dest string, width, height int) (string, error) {
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
func ToThumb(src, dest string, sizeSquared int) (string, error) {
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
func ToWebp(src, dest string, vendorTempDir bool) (string, error) {
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
		return "", fmt.Errorf("to webp mimetype %q != %s: %w",
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
		webp.Dest(file.Vendor())
	}
	if strings.HasSuffix(src, "-cropped.png") {
		defer os.Remove(src)
	}
	if err = webp.Run(); err != nil {
		if err1 := file.RemoveWebP(dest); err1 != nil {
			return "", fmt.Errorf("to webp cleanup: %w", err1)
		}
		return "", fmt.Errorf("to webp run: %w", err)
	}
	return "»webp", nil
}

// cropWebP crops an image to be usable size for WebP conversion.
func cropWebP(name string) (string, error) {
	w, h, _, err := Info(name)
	if err != nil {
		return "", fmt.Errorf("to webp size: %w", err)
	}
	if w+h > WebpMaxSize {
		fmt.Printf("crop to %dx%d", w, h)
		cropW, cropH := WebPCalc(w, h)
		ext := filepath.Ext(name)
		fn := strings.TrimSuffix(name, ext)
		crop := fn + "-cropped" + ext
		_, err := ToPng(name, crop, cropW, cropH)
		if err != nil {
			return "", fmt.Errorf("webp crop: %w", err)
		}
		return crop, nil
	}
	return name, nil
}

// WebPCalc calculates the largest permitted sizes for a valid WebP crop.
func WebPCalc(width, height int) (w, h int) {
	if width+height <= WebpMaxSize {
		return width, height
	}
	if width == height {
		const split = 2
		r := WebpMaxSize / split
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
func Width(name string) (int, error) {
	width, _, _, err := Info(name)
	if err != nil {
		return 0, fmt.Errorf("width: %w", err)
	}
	return width, nil
}
