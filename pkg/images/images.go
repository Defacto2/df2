// Package images generate thumbnails and converts between image formats.
package images

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"image"
	_ "image/gif"  // register GIF decoding.
	_ "image/jpeg" // register Jpeg decoding.
	"image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/images/internal/file"
	"github.com/Defacto2/df2/pkg/images/internal/imagemagick"
	"github.com/Defacto2/df2/pkg/images/internal/netpbm"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"github.com/nickalie/go-webpbin"
	"github.com/yusukebe/go-pngquant"
	_ "golang.org/x/image/bmp"  // register BMP decoding.
	_ "golang.org/x/image/tiff" // register TIFF decoding.
	_ "golang.org/x/image/webp" // register WebP decoding.
)

var ErrFormat = errors.New("unsupported image format")

const (
	WebpMaxSize int = 16383 // WebpMaxSize is the maximum pixel dimension of an webp image.

	thumbWidth int         = 400
	fperm      os.FileMode = 0o666
	fmode                  = os.O_RDWR | os.O_CREATE

	gif  = ".gif"
	jpg  = ".jpg"
	jpeg = ".jpeg"
	_png = ".png"
	tif  = ".tif"
	tiff = ".tiff"
	webp = ".webp"
)

// Fix generates any missing assets from downloads that are images.
func Fix(db *sql.DB, w io.Writer) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	dir, err := directories.Init(conf.Defaults(), false)
	if err != nil {
		return err
	}
	rows, err := db.Query(`SELECT id, uuid, filename, filesize ` +
		`FROM files WHERE platform="image" ORDER BY id ASC`)
	if err != nil {
		return fmt.Errorf("images fix query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("images fix rows: %w", rows.Err())
	}
	defer rows.Close()
	c := 0
	for rows.Next() {
		img := file.Image{}
		if err = rows.Scan(&img.ID, &img.UUID, &img.Name, &img.Size); err != nil {
			return fmt.Errorf("images fix rows scan: %w", err)
		}
		if directories.ArchiveExt(img.Name) {
			continue
		}
		if b, err := img.IsDir(&dir); err != nil {
			return err
		} else if b {
			continue
		}
		c++
		fmt.Fprintf(w, "%d. %v", c, img)
		if _, err := os.Stat(filepath.Join(dir.UUID, img.UUID)); errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(w, "%s\n", str.X())
			continue
		} else if err != nil {
			return fmt.Errorf("images fix stat: %w", err)
		}
		if err := Generate(w, filepath.Join(dir.UUID, img.UUID), img.UUID, false); err != nil {
			// return fmt.Errorf("images fix generate: %w", err)
			fmt.Fprintf(w, "%s", err)
		}
		fmt.Fprintln(w)
	}
	if c == 0 {
		fmt.Fprintln(w, "everything is okay, there is nothing to do")
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
func Generate(w io.Writer, src, id string, remove bool) error {
	if w == nil {
		w = io.Discard
	}
	if _, err := os.Stat(src); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("generate stat %q: %w", src, err)
	}
	f, err := directories.Files(conf.Defaults(), id)
	if err != nil {
		return err
	}
	// these funcs use dependencies that are not thread safe
	// convert to png
	pngDest, webpDest := ReplaceExt(_png, f.Img000), ReplaceExt(webp, f.Img000)
	const width = 1500
	s, err := ToPNG(src, pngDest, width, width)
	if err != nil {
		return fmt.Errorf("could not generate from %s: %s: %w", src, pngDest, err)
	}
	fmt.Fprintf(w, "  %s", s)
	// use netpbm or imagemagick to convert unsupported image formats into PNG
	if !file.Check(pngDest, err) {
		if err := Libraries(w, src, pngDest, remove); err != nil {
			if err := file.Remove(remove, src); err != nil {
				fmt.Fprintf(w, "could not remove %s: %s\n", src, err)
			}
		} else {
			src = pngDest // use the png created by the libraries
		}
	}
	// convert to webp
	s, err = ToWebp(w, src, webpDest, true)
	if err != nil {
		if !errors.Is(err, ErrFormat) {
			return fmt.Errorf("could not generate webp from %s: %s: %w", src, webpDest, err)
		}
		s, err = ToWebp(w, pngDest, webpDest, true)
		if err != nil {
			return fmt.Errorf("could not generate webp from %s: %s: %w", pngDest, webpDest, err)
		}
	}
	fmt.Fprintf(w, "  %s", s)
	webpOk := file.Check(webpDest, err)
	// make 400x400 thumbs
	s, err = ToThumb(src, f.Img400, thumbWidth)
	if err != nil {
		s, err = ToThumb(pngDest, f.Img400, thumbWidth)
	} else if err != nil && webpOk {
		s, err = ToThumb(webpDest, f.Img400, thumbWidth)
	}
	if err != nil {
		return fmt.Errorf("could not generate thumbs from any sources: %s: %w", src, err)
	}
	fmt.Fprintf(w, "  %s", s)
	return file.Remove(remove, src)
}

func Libraries(w io.Writer, src, pngDest string, remove bool) error {
	if w == nil {
		w = io.Discard
	}
	prog, err := netpbm.ID(src)
	if err != nil {
		return err
	}
	if prog != "" {
		err := netpbm.Convert(w, src, pngDest)
		if err != nil {
			return err
		}
		if !file.Check(pngDest, err) {
			return err
		}
		return nil
	}
	if err = imagemagick.Convert(w, src, pngDest); err != nil {
		return err
	} else if !file.Check(pngDest, err) {
		return file.Remove(remove, src)
	}
	return nil
}

// Copy a file from the source location to the destination.
func Copy(dest, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cp open: %w", err)
	}
	defer s.Close()
	d, err := os.OpenFile(dest, fmode, fperm)
	if err != nil {
		return fmt.Errorf("cp create: %w", err)
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	if err != nil {
		return fmt.Errorf("cp io copy: %w", err)
	}
	return nil
}

// Move a file from the source location to the destination.
// This is used in situations where os.rename() fails due to multiple partitions.
func Move(dest, src string) error {
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
	_, err = io.Copy(d, s)
	if err != nil {
		return fmt.Errorf("move io copy: %w", err)
	}
	if err = os.Remove(src); err != nil {
		return fmt.Errorf("move remove: %w", err)
	}
	return nil
}

// Info returns the image metadata.
func Info(name string) ( //nolint:nonamedreturns
	width int, height int, format string, err error,
) {
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

// ReplaceExt replaces or appends the extension to a file name.
func ReplaceExt(ext, name string) string {
	e := filepath.Ext(name)
	if e == "" {
		return name + ext
	}
	fn := strings.TrimSuffix(name, e)
	return fn + ext
}

// ToPNG converts any supported format to a compressed PNG image.
// helpful: https://www.programming-books.io/essential/
// go/images-png-jpeg-bmp-tiff-webp-vp8-gif-c84a45304ec3498081c67aa1ea0d9c49.
func ToPNG(src, dest string, width, height int) (string, error) {
	f, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("to png open %s: %w", src, err)
	}
	defer f.Close()
	// image.Decode will determine the format
	img, ext, err := image.Decode(f)
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
	p := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	// write the PNG data to img
	bb := &bytes.Buffer{}
	if err = p.Encode(bb, img); err != nil {
		return "", fmt.Errorf("to png buffer encode: %w", err)
	}
	// save the PNG to a file
	d, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("to png create %q: %w", dest, err)
	}
	defer d.Close()
	if _, err := bb.WriteTo(d); err != nil {
		return "", fmt.Errorf("to png buffer write to: %w", err)
	}
	return fmt.Sprintf("%v»png", strings.ToLower(ext)), nil
}

// ToThumb creates a thumb from an image that is pixel squared in size.
func ToThumb(src, dest string, sizeSquared int) (string, error) {
	s := "_" + fmt.Sprintf("%v", sizeSquared) + "x"
	cp, err := Duplicate(src, s)
	if err != nil {
		return "", fmt.Errorf("to thumb duplicate: %w", err)
	}
	img, err := imaging.Open(cp)
	if err != nil {
		return "", fmt.Errorf("to thumb imaging open: %w", err)
	}
	img = imaging.Resize(img, sizeSquared, 0, imaging.Lanczos)
	img = imaging.CropAnchor(img, sizeSquared, sizeSquared, imaging.Center)
	// use the 3rd party CLI tool, pngquant to compress the PNG data
	img, err = pngquant.Compress(img, "4")
	if err != nil {
		return "", fmt.Errorf("to thumb png quant compress: %w", err)
	}
	f := ReplaceExt(_png, cp)
	if err := imaging.Save(img, f); err != nil {
		return "", fmt.Errorf("to thumb imaging save: %w", err)
	}
	pngDest := ReplaceExt(_png, dest)
	if err := os.Rename(f, pngDest); err != nil {
		var le *os.LinkError // invalid cross-device link
		if !errors.As(err, &le) {
			return "", fmt.Errorf("to thumb rename: %w", err)
		}
		// attempt to fix cross-device link
		if err = Move(pngDest, f); err != nil {
			return "", fmt.Errorf("to thumb move: %w", err)
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
// Input format can be either GIF, PNG, JPEG, TIFF, WebP or raw Y'CbCr samples.
func ToWebp(w io.Writer, src, dest string, vendorTempDir bool) (string, error) {
	if w == nil {
		w = io.Discard
	}
	input, rm, err := checkWebP(src)
	if err != nil {
		return "", err
	}
	if rm {
		defer os.Remove(input)
	}
	input, err = cropWebP(w, input)
	if err != nil {
		return "", fmt.Errorf("to webp crop: %w", err)
	}
	const percent = 70
	webp := webpbin.NewCWebP().
		Quality(percent).
		InputFile(input).
		OutputFile(dest)
	if vendorTempDir {
		webp.Dest(file.Vendor())
	}
	if strings.HasSuffix(input, "-cropped.png") {
		defer os.Remove(input)
	}
	if err = webp.Run(); err != nil {
		if err1 := file.Remove0byte(dest); err1 != nil {
			return "", fmt.Errorf("to webp cleanup: %w", err1)
		}
		return "", fmt.Errorf("to webp run: %w", err)
	}
	return "»webp", nil
}

// checkWebP checks the validity of and returns the absolute path of the input source image,
// a returned true value means the input source is temporary and after use, it should be deleted.
func checkWebP(src string) (string, bool, error) {
	valid := func(a []string, x string) bool {
		for _, n := range a {
			if x == n {
				return true
			}
		}
		return false
	}
	v := []string{_png, jpg, jpeg, tif, tiff, webp}
	input := src
	// skip if already a webp image, or handle all other errors
	m, err := mimetype.DetectFile(src)
	switch {
	case m.Extension() == gif:
		// Dec 2022, https://github.com/nickalie/go-webpbin
		// currently does not support the library,
		// gif2webp -- Tool for converting GIF images to WebP
		f, err := os.CreateTemp("", "df2-gifToWebp.png")
		if err != nil {
			return "", false, err
		}
		_, err = ToPNG(src, f.Name(), 0, 0)
		if err != nil {
			return "", true, fmt.Errorf("to webp gif-topng: %w", err)
		}
		input = f.Name()
	case m.Extension() == webp:
		return "", false, nil
	case err != nil:
		return "", false, fmt.Errorf("to webp mimetype detect: %w", err)
	case !valid(v, m.Extension()):
		return "", false, fmt.Errorf("to webp mimetype %q != %s: %w",
			m.Extension(), strings.Join(v, " "), ErrFormat)
	}
	if m.Extension() == gif {
		return input, true, nil
	}
	return input, false, nil
}

// cropWebP crops an image to be usable size for WebP conversion.
func cropWebP(w io.Writer, name string) (string, error) {
	if w == nil {
		w = io.Discard
	}
	wp, hp, _, err := Info(name)
	if err != nil {
		return "", fmt.Errorf("to webp size: %w", err)
	}
	if wp+hp > WebpMaxSize {
		fmt.Fprintf(w, "crop to %dx%d", w, hp)
		cropW, cropH := WebPCalc(wp, hp)
		ext := filepath.Ext(name)
		fn := strings.TrimSuffix(name, ext)
		crop := fn + "-cropped" + ext
		if _, err := ToPNG(name, crop, cropW, cropH); err != nil {
			return "", fmt.Errorf("webp crop: %w", err)
		}
		return crop, nil
	}
	return name, nil
}

// WebPCalc calculates the largest permitted sizes for a valid WebP crop.
func WebPCalc(width, height int) (w, h int) { //nolint:nonamedreturns
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
