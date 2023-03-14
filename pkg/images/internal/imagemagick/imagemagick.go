package imagemagick

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

// imagemagick tools require the installation of ImageMagick v6.
// ubuntu: sudo apt install imagemagick
//
// List all supported formats
// identify -list format
//
// Convert Between Image Formats
// https://imagemagick.org/script/convert.php

var ErrFmt = errors.New("imagemagick does not support this image")

// Convert uses the magick convert command to convert an image to PNG.
func Convert(w io.Writer, src, dest string) error {
	if _, err := ID(src); err != nil {
		return err
	}

	// command on ubuntu: magick convert rose.jpg -resize 50% rose.png
	const file = "convert"
	args := []string{src, dest}
	path, err := exec.LookPath(file)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, path, args...)
	fmt.Fprintf(w, "running %s %s\n", file, args)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if len(out) > 0 {
		fmt.Fprintln(w, "magick:", string(out))
	}

	return nil
}

func ID(src string) ([]byte, error) {
	const file = "identify"
	args := []string{src}

	path, err := exec.LookPath(file)
	if err != nil {
		return nil, err
	}

	const three = 3 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), three)
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
