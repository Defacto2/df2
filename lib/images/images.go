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
func Duplicate(name string, prefix string) (string, error) {
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
	_, err = io.Copy(dest, src)
	if err != nil {
		return "", err
	}
	return fn + prefix + ext, nil
}

// Generate a collection of site images.
func Generate(n string, id string, d directories.Dir) {
	f := directories.Files(id)
	// make these 4 image tasks multithread
	c := make(chan bool)
	go func() { e := ToPng(n, NewExt(f.Img000, ".png"), 1500); logs.Log(e); c <- true }()
	go func() { e := ToWebp(n, NewExt(f.Img000, ".webp")); logs.Log(e); c <- true }()
	go func() { e := ToThumb(n, f.Img400, 400); logs.Log(e); c <- true }()
	go func() { e := ToThumb(n, f.Img150, 150); logs.Log(e); c <- true }()
	<-c // sync 4 tasks
	os.Remove(n)
}

// NewExt replaces or appends the extension to a file name.
func NewExt(name string, extension string) string {
	e := filepath.Ext(name)
	if e == "" {
		return name + extension
	}
	fn := strings.TrimSuffix(name, e)
	return fn + extension
}

// ToPng converts any supported format to a compressed PNG image.
// helpful: https://www.programming-books.io/essential/go/images-png-jpeg-bmp-tiff-webp-vp8-gif-c84a45304ec3498081c67aa1ea0d9c49
func ToPng(src string, dest string, maxDimension int) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	// image.Decode will determine the format
	img, ext, err := image.Decode(in)
	if err != nil {
		return err
	}
	// cap image size
	if maxDimension > 0 {
		img = imaging.Thumbnail(img, maxDimension, maxDimension, imaging.Lanczos)
	}
	// use the 3rd party CLI tool, pngquant to compress the PNG data
	img, err = pngquant.Compress(img, "4")
	if err != nil {
		return err
	}
	// adjust any configs to the PNG image encoder
	cfg := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	// write the PNG data to img
	buf := new(bytes.Buffer)
	err = cfg.Encode(buf, img)
	if err != nil {
		return err
	}
	// save the PNG to a file
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	s, err := buf.WriteTo(out)
	if err != nil {
		return err
	}
	fmt.Printf("Converted %v to a compressed PNG, %v\n", ext, humanize.Bytes(uint64(s)))
	return nil
}

// ToThumb creates a thumb from an image that is size pixel in width and height.
func ToThumb(file string, saveDir string, size int) error {
	pfx := "_" + fmt.Sprintf("%v", size) + "x"
	cp, err := Duplicate(file, pfx)
	if err != nil {
		return err
	}
	in, err := imaging.Open(cp)
	if err != nil {
		return err
	}
	in = imaging.Resize(in, size, 0, imaging.Lanczos)
	in = imaging.CropAnchor(in, size, size, imaging.Center)
	// use the 3rd party CLI tool, pngquant to compress the PNG data
	in, err = pngquant.Compress(in, "4")
	if err != nil {
		return err
	}
	f := NewExt(cp, ".png")
	if err := imaging.Save(in, f); err != nil {
		return err
	}
	s := NewExt(saveDir, ".png")
	if err := os.Rename(f, s); err != nil {
		return err
	}
	fmt.Printf("Generating thumbnail x%v, %v\n", size, filesize(s))
	if err := os.Remove(cp); err != nil {
		return err
	}
	return nil
}

// ToWebp converts any supported format to a WebP image using a 3rd party library.
func ToWebp(src string, dest string) error {
	// skip if already a webp image
	if m, _ := mimetype.DetectFile(src); m.Extension() == ".webp" {
		return nil
	}
	err := webpbin.NewCWebP().
		Quality(70).
		InputFile(src).
		OutputFile(dest).
		Run()
	fmt.Println("Converted to WebP,", filesize(dest))
	return err
}

func filesize(name string) string {
	f, err := os.Stat(name)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%v", humanize.Bytes(uint64(f.Size())))
}
