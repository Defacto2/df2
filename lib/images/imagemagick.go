package images

import (
	"errors"
	"fmt"
	"log"

	"gopkg.in/gographics/imagick.v2/imagick"
)

// imagemagick functions require the ImageMagick C library.
// Installation: sudo apt install libmagickwand-dev-6 netpbm

var ErrFmt = errors.New("imagemagick c does not support this image")

// ToMagick uses the imagemagick C library to convert an image to PNG.
func ToMagick(src, dest string) error {

	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if err := mw.ReadImage(src); err != nil {
		log.Println("imagick requirements: have you installed libmagickwand-dev-6 and netpbm?")
		log.Fatalln(err)
	}

	gif := mw.GetImageFormat()
	if gif == "" {
		return ErrFmt
	}

	if err := mw.WriteImage(dest); err != nil {
		return fmt.Errorf("imagick write: %w", err)
	}

	return nil
}
