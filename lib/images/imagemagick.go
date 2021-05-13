package images

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// imagemagick tools require the installation of ImageMagick v6.
// ubuntu: sudo apt install imagemagick
//
// Convert Between Image Formats
// https://imagemagick.org/script/convert.php

var ErrFmt = errors.New("imagemagick does not support this image")

// ToMagick uses the magick convert command to convert an image to PNG.
func ToMagick(src, dest string) error {

	if _, err := IDMagick(src); err != nil {
		return err
	}

	// command on ubuntu: magick convert rose.jpg -resize 50% rose.png
	const file = "convert"
	var args = []string{src, dest}

	path, err := exec.LookPath(file)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, path, args...)
	fmt.Printf("running %s %s\n", file, args)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if len(out) > 0 {
		fmt.Println("magick:", out)
	}

	return nil
}

func IDMagick(src string) ([]byte, error) {

	const file = "identify"
	var args = []string{src}

	path, err := exec.LookPath(file)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	const unknown = "no decode delegate for this image format"
	if bytes.Contains(out, []byte(unknown)) {
		return nil, ErrFmt
	}

	return out, nil
}
