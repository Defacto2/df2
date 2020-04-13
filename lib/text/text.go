package text

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/images"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
)

/*
   https://github.com/ansilove/ansilove
   /usr/local/bin/ansilove
   ansilove [-dhiqrsv] [-b bits] [-c columns] [-f font] [-m mode] [-o file]
            [-R factor] file

     -b bits     Set to 9 to render 9th column of block characters (default:
                 8).
     -c columns  Adjust number of columns for ANSI, BIN, and TND files.
     -d          Enable DOS aspect ratio.
     -f font     Select font for supported formats.
     -h          Show help.
     -i          Enable iCE colors.
     -m mode     Set rendering mode for ANS files. Valid options are:
                 ced     Black on gray, with 78 columns.
                 transparent
                         Render with transparent background.
                 workbench
                         Use Amiga Workbench palette.
     -o file     Specify output filename/path.
     -q          Suppress output messages (quiet).
     -r          Create Retina @2x output file.
     -R factor   Create Retina output file with custom scale factor.
     -s          Show SAUCE record without generating output.
     -v          Show version information.
*/

// Generate a collection of site images.
func Generate(name, id string) {
	var n string = name
	out := func(s string, e error) {
		if s != "" {
			print("  ", s)
		} else {
			logs.Log(e)
		}
	}
	f := directories.Files(id)
	o := f.Img000 + ".png"
	s, err := ToPng(n, f.Img000)
	if err != nil && err.Error() == `execute ansilove: executable file not found in $PATH` {
		fmt.Println("\n\nthis command requires the installation of AnsiLove/C")
		fmt.Println("installation instructions: https://github.com/ansilove/ansilove")
		fmt.Println()
		logs.Check(err)
	}
	out(s, err)
	// cap images to the webp limit of 16383 pixels
	if w, err := images.Width(o); w > 16383 {
		out(fmt.Sprintf("width %dpx", w), err)
		s, err = images.ToPng(o, images.NewExt(o, ".png"), 16383)
		out(s, err)
	}
	s, err = images.ToWebp(o, images.NewExt(o, ".webp"))
	out(s, err)
	s, err = images.ToThumb(o, f.Img400, 400)
	out(s, err)
	s, err = images.ToThumb(o, f.Img150, 150)
	out(s, err)
	//os.Remove(n)
}

// ToPng converts any supported format to a compressed PNG image.
// helpful: https://www.programming-books.io/essential/go/images-png-jpeg-bmp-tiff-webp-vp8-gif-c84a45304ec3498081c67aa1ea0d9c49
func ToPng(src, dest string) (string, error) {
	img := dest + ".png"
	cmd := exec.Command("ansilove", "-r", "-o", img, src)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	stat, err := os.Stat(img)
	if os.IsNotExist(err) {
		return "", err
	}
	ss := stat.Size()
	return fmt.Sprintf("✓ text » png %v\n%s", humanize.Bytes(uint64(ss)), color.Secondary.Sprintf("%s", out)), nil
}
