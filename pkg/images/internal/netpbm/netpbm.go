package netpbm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

// netpbm requires the installation of Netpbm.
// ubuntu: sudo apt install netpbm
//
// Netpbm is a package of graphics programs and a programming library.
// http://netpbm.sourceforge.net/doc/

var (
	ErrFmt = errors.New("netpdm package hasn't been configured to support this image")
	ErrSrc = errors.New("src file does not exist")
)

type Program string

type Idents map[string]Program

const (
	GEM  Program = "gemtopnm"  // GEM .img format to Netpbm format.
	Gif  Program = "giftopnm"  // GIF to PNM.
	Ilbm Program = "ilbmtoppm" // IFF ILBM to PPM.
	RLE  Program = "cistopbm"  // Compuserve RLE image to PBM.
	PNM  Program = "pnmtopng"  // Netpbm format to Portable Network Graphics.
)

func Extensions() Idents {
	ids := make(Idents)
	ids[".img"] = GEM
	ids[".ximg"] = GEM
	ids[".timg"] = GEM
	ids[".gif"] = Gif
	ids[".iff"] = Ilbm
	ids[".lbm"] = Ilbm
	ids[".bbm"] = Ilbm
	ids[".ilbm"] = Ilbm
	ids[".pic"] = Ilbm
	ids[".rle"] = RLE
	return ids
}

// Convert uses netpdm to convert a configured image to PNG.
func Convert(w io.Writer, src, dest string) error {
	if w == nil {
		w = io.Discard
	}
	if _, err := os.Stat(src); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w: %s", ErrSrc, src)
	} else if err != nil {
		return err
	}

	prog, err := ID(src)
	if err != nil {
		return err
	}

	input, err := exec.LookPath(string(prog))
	if err != nil {
		return err
	}
	output, err := exec.LookPath(string(PNM))
	if err != nil {
		return err
	}

	// shell command to replicate
	// ilbmtoppm < pic.ilbm | pnmtopng - > pic.png
	// < input overwrite
	// > output overwrite

	// hack to handle bash piping and redirections
	bash, err := exec.LookPath("bash")
	if err != nil {
		return err
	}
	cmdStr := fmt.Sprintf("%s < %s | %s > %s", input, src, output, dest)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, bash, "-c", cmdStr)
	fmt.Fprintf(w, "running %s\n", cmdStr)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if len(out) > 0 {
		fmt.Fprintf(w, "%s: %s", prog, string(out))
	}
	return nil
}

func ID(src string) (Program, error) {
	ext := filepath.Ext(src)
	prog := Extensions()[ext]
	if prog == "" {
		return "", ErrFmt
	}
	return prog, nil
}
