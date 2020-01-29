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
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/disintegration/imaging"
	"github.com/dustin/go-humanize"
	"github.com/gabriel-vasile/mimetype"
	"github.com/nickalie/go-webpbin"
	"github.com/yusukebe/go-pngquant"
	_ "golang.org/x/image/bmp"  // register BMP decoding
	_ "golang.org/x/image/tiff" // register TIFF decoding
	_ "golang.org/x/image/webp" // register WebP decoding
)

// Duplicate an image file and appends prefix to its name.
func Duplicate(name, prefix string) (string, error) {
	src, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer src.Close()
	ext := filepath.Ext(name)
	fn := strings.TrimSuffix(name, ext)
	dest, err := os.OpenFile(fn+prefix+ext, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer dest.Close()
	if _, err = io.Copy(dest, src); err != nil {
		return "", err
	}
	return fn + prefix + ext, nil
}

// Generate a collection of site images.
func Generate(name, id string, rem bool) {
	var n string = name
	out := func(s string, e error) {
		if s != "" {
			print("  ", s)
		} else {
			logs.Log(e)
		}
	}
	f := directories.Files(id)
	// these funcs use dependencies that are not thread safe
	s, err := ToPng(n, NewExt(f.Img000, ".png"), 1500)
	out(s, err)
	s, err = ToWebp(n, NewExt(f.Img000, ".webp"))
	out(s, err)
	s, err = ToThumb(n, f.Img400, 400)
	out(s, err)
	s, err = ToThumb(n, f.Img150, 150)
	out(s, err)
	if rem {
		os.Remove(n)
	}
}

// NewExt replaces or appends the extension to a file name.
func NewExt(name, extension string) string {
	e := filepath.Ext(name)
	if e == "" {
		return name + extension
	}
	fn := strings.TrimSuffix(name, e)
	return fn + extension
}

// Info returns the image height, width and format.
func Info(name string) (int, int, string, error) {
	file, err := os.Open(name)
	if err != nil {
		return 0, 0, "", err
	}
	defer file.Close()
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, "", err
	}
	return config.Height, config.Width, format, nil
}

// Width returns the image width in pixels.
func Width(name string) (int, error) {
	_, w, _, err := Info(name)
	return w, err
}

func getImageDimension(imagePath string) (int, int) {
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
	}
	return image.Width, image.Height
}

// ToPng converts any supported format to a compressed PNG image.
// helpful: https://www.programming-books.io/essential/go/images-png-jpeg-bmp-tiff-webp-vp8-gif-c84a45304ec3498081c67aa1ea0d9c49
func ToPng(src, dest string, maxDimension int) (string, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()
	// image.Decode will determine the format
	img, ext, err := image.Decode(in)
	if err != nil {
		return "", err
	}
	// cap image size
	if maxDimension > 0 {
		img = imaging.Thumbnail(img, maxDimension, maxDimension, imaging.Lanczos)
	}
	// use the 3rd party CLI tool, pngquant to compress the PNG data
	img, err = pngquant.Compress(img, "4")
	if err != nil {
		return "", err
	}
	// adjust any configs to the PNG image encoder
	cfg := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	// write the PNG data to img
	buf := new(bytes.Buffer)
	err = cfg.Encode(buf, img)
	if err != nil {
		return "", err
	}
	// save the PNG to a file
	out, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer out.Close()
	s, err := buf.WriteTo(out)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %v » png %v", "✓", strings.ToLower(ext), humanize.Bytes(uint64(s))), nil
}

// ToThumb creates a thumb from an image that is size pixel in width and height.
func ToThumb(file, saveDir string, size int) (string, error) {
	pfx := "_" + fmt.Sprintf("%v", size) + "x"
	cp, err := Duplicate(file, pfx)
	if err != nil {
		return "", err
	}
	in, err := imaging.Open(cp)
	if err != nil {
		return "", err
	}
	in = imaging.Resize(in, size, 0, imaging.Lanczos)
	in = imaging.CropAnchor(in, size, size, imaging.Center)
	// use the 3rd party CLI tool, pngquant to compress the PNG data
	in, err = pngquant.Compress(in, "4")
	if err != nil {
		return "", err
	}
	f := NewExt(cp, ".png")
	if err := imaging.Save(in, f); err != nil {
		return "", err
	}
	s := NewExt(saveDir, ".png")
	if err := os.Rename(f, s); err != nil {
		return "", err
	}
	if _, err := os.Stat(cp); err == nil {
		if err := os.Remove(cp); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%s %vx%v %v", "✓", size, size, filesize(s)), nil
}

// ToWebp converts any supported format to a WebP image using a 3rd party library.
func ToWebp(src, dest string) (string, error) {
	// skip if already a webp image
	if m, _ := mimetype.DetectFile(src); m.Extension() == ".webp" {
		return "", nil
	}
	err := webpbin.NewCWebP().
		Quality(70).
		InputFile(src).
		OutputFile(dest).
		Run()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s » webp %v", "✓", filesize(dest)), nil
}

func filesize(name string) string {
	f, err := os.Stat(name)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%v", humanize.Bytes(uint64(f.Size())))
}
